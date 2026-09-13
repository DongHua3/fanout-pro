package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// version 由构建时通过 -ldflags 注入。
var version = "dev"

func main() {
	var (
		webPort  = flag.Int("web", 8899, "Web 管理端口")
		maxSlots = flag.Int("max", 20, "最多同时运行的隧道数")
		workDir  = flag.String("dir", "/var/lib/fanout", "工作目录")
	)
	panelMode := flag.String("panel", "", "节点链接后端: 留空按界面设置/自动探测, 3x-ui, native, xray-cf-lite")
	publicIP := flag.String("ip", "", "母机公网 IPv4，用于分享链接/SOCKS5 地址；留空则自动探测")
	showVersion := flag.Bool("version", false, "显示版本后退出")
	flag.Parse()

	if *publicIP == "" {
		*publicIP = os.Getenv("FANOUT_PUBLIC_IP")
	}

	if *showVersion {
		fmt.Println("fanout", version)
		return
	}

	if os.Geteuid() != 0 {
		log.Fatal("需要 root 权限（要创建 netns 和改 iptables）")
	}
	if err := os.MkdirAll(*workDir, 0700); err != nil {
		log.Fatalf("创建工作目录失败: %v", err)
	}
	setPublicIPOverride(*publicIP)
	go hostPublicIP() // 预热探测，别让首个请求阻塞
	if err := prepareHost(); err != nil {
		log.Fatal(err)
	}

	configurePanel(*workDir, *panelMode)
	if p, err := openPanel(); err != nil {
		log.Printf("节点链接后端暂不可用（可在 Web 界面查看原因）: %v", err)
	} else {
		log.Printf("节点链接后端: %s", p.Describe())
	}

	mgr := NewManager(*maxSlots, *workDir)
	log.Printf("正在拉取节点列表...")
	if n, err := mgr.RefreshNodes(); err != nil {
		log.Printf("拉取失败（可在 Web 界面重试）: %v", err)
	} else {
		log.Printf("已获取 %d 个节点", n)
	}

	if n, err := mgr.restoreState(); err != nil {
		log.Printf("恢复上次状态失败: %v", err)
	} else if n > 0 {
		log.Printf("正在恢复上次的 %d 条隧道", n)
		// 3x-ui 模式重启不会自动重写面板出站，旧版本升上来时面板里的 socks
		// 出站没有认证字段，端口一旦要认证就连不上，这里恢复后对账一次
		go mgr.ReconcileOutbounds()
	}

	go mgr.WatchHealth()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("正在清理所有隧道...")
		mgr.Shutdown()
		closePanel()
		os.Exit(0)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/api/nodes", apiNodes(mgr))
	mux.HandleFunc("/api/nodes/start", apiNodesStart(mgr))
	mux.HandleFunc("/api/tunnels", apiTunnels(mgr))
	mux.HandleFunc("/api/start", apiStart(mgr))
	mux.HandleFunc("/api/stop", apiStop(mgr))
	mux.HandleFunc("/api/swap", apiSwap(mgr))
	mux.HandleFunc("/api/retry", apiRetry(mgr))
	mux.HandleFunc("/api/cred", apiCred(mgr))
	mux.HandleFunc("/api/refresh", apiRefresh(mgr))
	mux.HandleFunc("/api/regions", apiRegions(mgr))
	mux.HandleFunc("/api/provision", apiProvision(mgr))
	mux.HandleFunc("/api/jobs", apiJobs(mgr))
	mux.HandleFunc("/api/jobs/dismiss", apiJobDismiss(mgr))
	mux.HandleFunc("/api/xui", apiXUIStatus)
	mux.HandleFunc("/api/xui/inbounds", apiXUIInbounds(mgr))
	mux.HandleFunc("/api/xui/bind", apiXUIBind(mgr))
	mux.HandleFunc("/api/xui/clone", apiXUIClone(mgr))
	mux.HandleFunc("/api/xui/detail", apiXUIDetail)
	mux.HandleFunc("/api/xui/links", apiXUILinks(mgr))
	mux.HandleFunc("/api/xui/delete", apiXUIDelete(mgr))
	mux.HandleFunc("/api/panel/inbound/new", apiInboundCreate(mgr))
	mux.HandleFunc("/api/panel/inbound/update", apiInboundUpdate(mgr))
	mux.HandleFunc("/api/panel/client/add", apiClientAdd(mgr))
	mux.HandleFunc("/api/panel/client/del", apiClientDelete(mgr))
	mux.HandleFunc("/api/panel/client/reset", apiClientReset(mgr))
	mux.HandleFunc("/api/panel/mode", apiPanelMode(*workDir))
	mux.HandleFunc("/api/port/check", apiPortCheck)
	auth, created, err := NewAuth(*workDir)
	if err != nil {
		log.Fatalf("初始化访问口令失败: %v", err)
	}
	if created {
		log.Printf("已生成访问口令，见 %s", filepath.Join(*workDir, "password"))
	}

	mux.HandleFunc("/api/exits", apiExits(mgr, auth))
	mux.HandleFunc("/api/sub", apiSubscription(mgr))
	mux.HandleFunc("/sub", apiSubscription(mgr))

	bpCreated, err := initBasePath(*workDir)
	if err != nil {
		log.Fatalf("初始化访问路径失败: %v", err)
	}
	if bpCreated {
		log.Printf("已生成访问路径，见 %s", filepath.Join(*workDir, "basepath"))
	}

	// 用户显式给了 -web 就以命令行为准，否则沿用界面上存过的端口
	portExplicit := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "web" {
			portExplicit = true
		}
	})
	webCfg, err := loadWebSettings(*workDir, *webPort, portExplicit)
	if err != nil {
		log.Fatalf("加载 Web 设置失败: %v", err)
	}

	srv := newWebServer(StripBasePath(auth.Wrap(mux)))
	// 设置面板：改密码 / 改路径 / 改端口 / 改本地监听 / 域名与 SSL 设置。
	mux.HandleFunc("/api/settings", apiSettings(auth, srv))
	mux.HandleFunc("/api/ssl/inspect", apiSSLInspect)
	mux.HandleFunc("/api/update/check", apiUpdateCheck)
	mux.HandleFunc("/api/update/apply", apiUpdateApply)

	scheme := "http"
	if webCfg.IsTLSEnabled() {
		scheme = "https"
	}
	hostDisplay := "<本机IP>"
	if webCfg.Domain != "" {
		hostDisplay = webCfg.Domain
	}
	log.Printf("管理界面: %s://%s%s%s/", scheme, hostDisplay, webCfg.listenAddrString(), currentBasePath())
	log.Printf("SOCKS5 端口在 %d-%d 之间随机分配", randPortMin, randPortMax)
	if err := srv.serve(); err != nil {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type NodeWithQuality struct {
	Node
	Quality QualityInfo `json:"quality"`
	Running bool        `json:"running"`
	Slot    int         `json:"slot,omitempty"`
}

func apiNodes(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodes, fetched := m.Nodes()

		activeHosts := map[string]int{}
		for _, t := range m.Tunnels() {
			activeHosts[t.Node.HostName] = t.Slot
			activeHosts[sanitizeTag(t.Node.HostName)] = t.Slot
		}

		ips := make([]string, len(nodes))
		for i, n := range nodes {
			ips[i] = n.IP
		}
		qc := GetQualityCache(m.workDir)
		qmap := qc.BatchEvaluate(ips)

		regionFilter := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("region")))
		typeFilter := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
		kw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kw")))

		list := make([]NodeWithQuality, 0, len(nodes))
		for _, n := range nodes {
			if regionFilter != "" && strings.ToUpper(n.CountryCode) != regionFilter {
				continue
			}

			q, ok := qmap[n.IP]
			if !ok {
				q = fallbackQuality(n.IP)
			}

			if typeFilter != "" && strings.ToLower(q.Type) != typeFilter {
				continue
			}

			if kw != "" {
				text := strings.ToLower(fmt.Sprintf("%s %s %s %s %s %s %s",
					n.HostName, n.IP, n.Country, n.CountryCode, q.ISP, q.Org, q.ASN))
				if !strings.Contains(text, kw) {
					continue
				}
			}

			slot, running := activeHosts[n.HostName]
			if !running {
				slot, running = activeHosts[sanitizeTag(n.HostName)]
			}

			list = append(list, NodeWithQuality{
				Node:    n,
				Quality: q,
				Running: running,
				Slot:    slot,
			})
		}

		sortBy := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("sort")))
		switch sortBy {
		case "speed":
			sort.Slice(list, func(i, j int) bool {
				if list[i].SpeedMbps != list[j].SpeedMbps {
					return list[i].SpeedMbps > list[j].SpeedMbps
				}
				return list[i].Quality.Score > list[j].Quality.Score
			})
		case "ping":
			sort.Slice(list, func(i, j int) bool {
				pi := list[i].Ping
				pj := list[j].Ping
				if pi > 0 && pj > 0 {
					if pi != pj {
						return pi < pj
					}
					return list[i].SpeedMbps > list[j].SpeedMbps
				}
				if pi > 0 {
					return true
				}
				if pj > 0 {
					return false
				}
				return list[i].SpeedMbps > list[j].SpeedMbps
			})
		default: // "quality"
			sort.Slice(list, func(i, j int) bool {
				if list[i].Quality.Score != list[j].Quality.Score {
					return list[i].Quality.Score > list[j].Quality.Score
				}
				if list[i].Quality.Type != list[j].Quality.Type {
					if list[i].Quality.Type == "residential" {
						return true
					}
					if list[j].Quality.Type == "residential" {
						return false
					}
				}
				if list[i].SpeedMbps != list[j].SpeedMbps {
					return list[i].SpeedMbps > list[j].SpeedMbps
				}
				return list[i].Ping < list[j].Ping
			})
		}

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(list) {
				list = list[:limit]
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"nodes":   list,
			"total":   len(list),
			"fetched": fetched,
		})
	}
}

// apiNodesStart 在节点大厅中根据主机名直接开启出口，并可选直接挂载指定入站。
func apiNodesStart(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Query().Get("host")
		bindTag := r.URL.Query().Get("bind_tag")
		bindPort := r.URL.Query().Get("bind_port")

		if r.Method == http.MethodPost && host == "" {
			var body struct {
				Host     string `json:"host"`
				BindTag  string `json:"bind_tag"`
				BindPort string `json:"bind_port"`
			}
			if json.NewDecoder(r.Body).Decode(&body) == nil {
				host = body.Host
				if body.BindTag != "" {
					bindTag = body.BindTag
				}
				if body.BindPort != "" {
					bindPort = body.BindPort
				}
			}
		}

		if host == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 host 参数"})
			return
		}

		targetTag := bindTag
		if targetTag == "" {
			targetTag = bindPort
		}

		t, err := m.StartByHost(host)
		if err != nil {
			if strings.Contains(err.Error(), "已在运行中") && targetTag != "" {
				var runningTunnel *Tunnel
				for _, tun := range m.Tunnels() {
					if tun.Node.HostName == host || tun.Node.IP == host || sanitizeTag(tun.Node.HostName) == sanitizeTag(host) {
						runningTunnel = tun
						break
					}
				}
				if runningTunnel != nil {
					if runningTunnel.Status == "up" {
						if p, pErr := openPanel(); pErr == nil {
							_ = p.Bind(targetTag, runningTunnel.Node.HostName, m.Tunnels())
							invalidateInbounds()
						}
					} else {
						go func(rt *Tunnel) {
							for i := 0; i < 40; i++ {
								time.Sleep(1 * time.Second)
								if rt.Status == "up" {
									if p, pErr := openPanel(); pErr == nil {
										_ = p.Bind(targetTag, rt.Node.HostName, m.Tunnels())
										invalidateInbounds()
									}
									break
								}
								if rt.Status == "failed" || rt.Status == "stopped" {
									break
								}
							}
						}(runningTunnel)
					}
					writeJSON(w, http.StatusOK, runningTunnel)
					return
				}
			}
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		if targetTag != "" {
			go func() {
				for i := 0; i < 40; i++ {
					time.Sleep(1 * time.Second)
					if t.Status == "up" {
						if p, err := openPanel(); err == nil {
							_ = p.Bind(targetTag, t.Node.HostName, m.Tunnels())
							invalidateInbounds()
						}
						break
					}
					if t.Status == "failed" || t.Status == "stopped" {
						break
					}
				}
			}()
		}

		writeJSON(w, http.StatusOK, t)
	}
}

func apiTunnels(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, m.Tunnels())
	}
}

func apiStart(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Query().Get("host")
		if host == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 host 参数"})
			return
		}
		nodes, _ := m.Nodes()
		for _, n := range nodes {
			if n.HostName == host {
				t, err := m.Start(n)
				if err != nil {
					writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
					return
				}
				writeJSON(w, http.StatusOK, t)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "节点不存在，可能列表已过期"})
	}
}

func apiStop(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slot, err := strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slot 参数无效"})
			return
		}
		if err := m.Stop(slot); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"ok": "已停止"})
	}
}

func apiRefresh(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		n, err := m.RefreshNodes()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"count": n})
	}
}

// apiSwap 就地把一个出口换到别的节点，端口不变。
func apiSwap(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slot, err := strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slot 参数无效"})
			return
		}
		if err := m.Swap(slot); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"ok": "正在换节点"})
	}
}

// apiRetry 手动重试连接当前槽位的节点。
func apiRetry(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slot, err := strconv.Atoi(r.URL.Query().Get("slot"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slot 参数无效"})
			return
		}
		if err := m.Retry(slot); err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"ok": "正在重试连接"})
	}
}

// apiRegions 给新建向导用：各地区还剩多少空闲节点。
func apiRegions(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, m.Regions())
	}
}

// apiCred 改一个出口的 SOCKS5 用户名口令。两个参数都留空表示随机重置。
func apiCred(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		slot, err := strconv.Atoi(q.Get("slot"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "slot 参数无效"})
			return
		}
		cred, err := m.SetCred(slot, SocksCred{
			User: strings.TrimSpace(q.Get("user")),
			Pass: strings.TrimSpace(q.Get("pass")),
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{
			"user": cred.User,
			"pass": cred.Pass,
		})
	}
}

// apiSettings 管理界面自身的设置：改密码 / 改路径 / 改端口 / 改本地监听 / 域名与 SSL 设置。
// GET 返回当前值（不含明文口令）；POST 按传入的字段逐项应用，任一项失败即整体回报。
func apiSettings(auth *Auth, srv *webServer) http.HandlerFunc {
	type settingsReq struct {
		Password   *string `json:"password"`    // 非空则改口令
		BasePath   *string `json:"base_path"`   // 提供即改访问路径（空串=去掉前缀）
		Port       *int    `json:"port"`        // 提供即改监听端口
		ListenAddr *string `json:"listen_addr"` // 提供即改监听地址
		Domain     *string `json:"domain"`      // 域名
		CertFile   *string `json:"cert_file"`   // 证书文件路径
		KeyFile    *string `json:"key_file"`    // 私钥文件路径
		SSLMode    *string `json:"ssl_mode"`    // SSL 模式
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var in settingsReq
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
				return
			}

			// 改口令
			if in.Password != nil && *in.Password != "" {
				if err := auth.SetPassword(*in.Password); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
					return
				}
			}
			// 改访问路径
			if in.BasePath != nil {
				if _, err := setBasePath(*in.BasePath); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
					return
				}
			}
			// 改端口 / 监听地址 / 域名 / SSL 设置：合成一份新的 WebSettings 一起应用
			if in.Port != nil || in.ListenAddr != nil || in.Domain != nil || in.CertFile != nil || in.KeyFile != nil || in.SSLMode != nil {
				next := getWebSettings()
				if in.Port != nil {
					next.Port = *in.Port
				}
				if in.ListenAddr != nil {
					next.ListenAddr = *in.ListenAddr
				}
				if in.Domain != nil {
					next.Domain = *in.Domain
				}
				if in.CertFile != nil {
					next.CertFile = *in.CertFile
				}
				if in.KeyFile != nil {
					next.KeyFile = *in.KeyFile
				}
				if in.SSLMode != nil {
					next.SSLMode = *in.SSLMode
				}
				if err := srv.applyWebSettings(next); err != nil {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
					return
				}
			}
		}

		cfg := getWebSettings()
		listen := cfg.ListenAddr
		if listen == "" {
			listen = "0.0.0.0"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"base_path":    currentBasePath(),
			"port":         cfg.Port,
			"listen_addr":  listen,
			"domain":       cfg.Domain,
			"cert_file":    cfg.CertFile,
			"key_file":     cfg.KeyFile,
			"ssl_mode":     cfg.SSLMode,
			"is_tls":       srv.IsTLS(),
			"has_password": true,
			"version":      version,
		})
	}
}

// apiSSLInspect 提供对传入或已配置证书的实时解析与有效性诊断。
func apiSSLInspect(w http.ResponseWriter, r *http.Request) {
	certFile := ""
	keyFile := ""

	if r.Method == http.MethodPost {
		var req struct {
			CertFile string `json:"cert_file"`
			KeyFile  string `json:"key_file"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		certFile = req.CertFile
		keyFile = req.KeyFile
	}
	if certFile == "" {
		certFile = r.URL.Query().Get("cert_file")
	}
	if keyFile == "" {
		keyFile = r.URL.Query().Get("key_file")
	}

	if certFile == "" || keyFile == "" {
		cfg := getWebSettings()
		if certFile == "" {
			certFile = cfg.CertFile
		}
		if keyFile == "" {
			keyFile = cfg.KeyFile
		}
	}

	if certFile == "" || keyFile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "请提供证书文件 (cert_file) 和私钥文件 (key_file) 路径",
		})
		return
	}

	detail, err := inspectCertificate(certFile, keyFile)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// apiUpdateCheck 问 GitHub 最新 release，回报当前/最新版本与更新内容。
func apiUpdateCheck(w http.ResponseWriter, r *http.Request) {
	st, err := checkUpdate()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "检查更新失败: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// apiUpdateApply 下载最新版替换二进制并重启服务。成功后进程会被拉起成新版本。
func apiUpdateApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "用 POST"})
		return
	}
	st, err := checkUpdate()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "检查更新失败: " + err.Error()})
		return
	}
	if !st.HasUpdate {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restarting": false, "message": "已经是最新版"})
		return
	}
	if err := applyUpdate(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// 先把响应发回去，restartSelf 已排在延迟后触发
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restarting": true, "latest": st.Latest})
}

// apiExits 返回主界面需要的一切：出口以及挂在它上面的入站。
func apiExits(m *Manager, a *Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		view := m.ExitsOf()
		if a != nil {
			view.SubToken = a.Password()
		}
		writeJSON(w, http.StatusOK, view)
	}
}

// apiSubscription 提供标准 Base64 订阅源，包含母机直连与全部出口节点
func apiSubscription(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x, err := openPanel()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		list, err := x.Inbounds(nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		var ids []int
		for _, ib := range list {
			if ib.Enable {
				ids = append(ids, ib.ID)
			}
		}
		host := publicHost(r)
		links, err := x.InboundLinks(ids, host)
		if err != nil && len(links) == 0 {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		allText := strings.Join(links, "\n")
		b64 := base64.StdEncoding.EncodeToString([]byte(allText))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Subscription-Userinfo", "upload=0; download=0; total=1073741824000; expire=0")
		w.Header().Set("Profile-Update-Interval", "24")
		_, _ = w.Write([]byte(b64))
	}
}

// apiProvision 接收"开 N 个某地区的出口"这个意图，返回作业 id 供轮询。
func apiProvision(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		count, err := strconv.Atoi(q.Get("count"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "count 参数无效"})
			return
		}
		tpl := 0
		if s := q.Get("template"); s != "" {
			if tpl, err = strconv.Atoi(s); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "template 参数无效"})
				return
			}
		}
		job, err := m.Provision(ProvisionRequest{
			Region: q.Get("region"), Count: count, TemplateID: tpl,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"job": job.ID()})
	}
}

func apiJobs(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, m.jobs.Views())
	}
}

func apiJobDismiss(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.jobs.Dismiss(r.URL.Query().Get("id"))
		writeJSON(w, http.StatusOK, map[string]string{"ok": "已关闭"})
	}
}

// apiXUIStatus 报告当前的节点链接后端：接管的 3x-ui，或 fanout 自己跑的 Xray。
func apiXUIStatus(w http.ResponseWriter, r *http.Request) {
	p, err := openPanel()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"available": false,
			"reason":    err.Error(),
		})
		return
	}
	// xray-cf-lite 模式下节点由 xray-cf-lite 管，fanout 只改路由，不提供新建入口。
	_, isXCL := p.(*XCL)
	resp := map[string]any{
		"available": true,
		"kind":      p.Kind(),
		"describe":  p.Describe(),
		// 自建/3x-ui 能建入站；xray-cf-lite 只能改路由
		"can_create": !isXCL,
	}
	if x, ok := p.(*XUI); ok {
		resp["port"] = x.Port
		resp["base_path"] = x.BasePath
		resp["scheme"] = x.Scheme
		resp["host"] = x.Host
	}
	writeJSON(w, http.StatusOK, resp)
}

// apiPanelMode 读取/切换节点链接后端。
// GET 返回当前模式与本机可选模式；POST {"mode":"..."} 运行时切换，空 mode = 恢复自动探测。
func apiPanelMode(workDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var in struct {
				Mode string `json:"mode"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
				return
			}
			p, err := switchPanelMode(in.Mode)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			invalidateInbounds()
			writeJSON(w, http.StatusOK, map[string]any{
				"mode":     currentPanelMode(),
				"kind":     p.Kind(),
				"describe": p.Describe(),
			})
			return
		}
		resp := map[string]any{
			"mode":  currentPanelMode(),
			"modes": availablePanelModes(workDir),
		}
		if p, err := openPanel(); err == nil {
			resp["kind"] = p.Kind()
			resp["describe"] = p.Describe()
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// apiXUIInbounds 列出面板里已有的入站及其绑定状态。
func apiXUIInbounds(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := cachedInbounds(liveHosts(m))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, list)
	}
}

// liveHosts 返回当前有连通隧道的节点标识集合。
func liveHosts(m *Manager) map[string]bool {
	live := map[string]bool{}
	for _, t := range m.Tunnels() {
		if t.Status == "up" {
			live[sanitizeTag(t.Node.HostName)] = true
		}
	}
	return live
}

// apiXUIBind 把某个入站绑定到某条出口，host 为空串或 "direct" 表示解绑恢复直连。
func apiXUIBind(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tag := r.URL.Query().Get("tag")
		id := r.URL.Query().Get("id")
		port := r.URL.Query().Get("port")
		host := r.URL.Query().Get("host")

		if r.Method == http.MethodPost && tag == "" && id == "" && port == "" && host == "" {
			var body struct {
				Tag  string `json:"tag"`
				ID   string `json:"id"`
				Port string `json:"port"`
				Host string `json:"host"`
			}
			if json.NewDecoder(r.Body).Decode(&body) == nil {
				tag = body.Tag
				id = body.ID
				port = body.Port
				host = body.Host
			}
		}

		target := tag
		if target == "" {
			if id != "" {
				target = id
			} else if port != "" {
				target = port
			}
		}
		if target == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 tag/id/port 参数"})
			return
		}

		if host == "direct" || host == "none" {
			host = ""
		}

		x, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		if err := x.Bind(target, host, m.Tunnels()); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		invalidateInbounds()

		status := "bound"
		if host == "" {
			status = "direct"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"status": status,
			"host":   host,
		})
	}
}

// apiXUIClone 以某个入站为模板，为所有已连通的隧道各复制一个入站并绑好出口。
func apiXUIClone(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id 参数无效"})
			return
		}

		tunnels := m.Tunnels()
		// 用节点主机名而非槽位号：槽位在重启后会重排，指代会错位
		var hosts []string
		if raw := r.URL.Query().Get("hosts"); raw != "" {
			for _, part := range strings.Split(raw, ",") {
				if h := strings.TrimSpace(part); h != "" {
					hosts = append(hosts, h)
				}
			}
		} else {
			for _, t := range tunnels {
				if t.Status == "up" {
					hosts = append(hosts, t.Node.HostName)
				}
			}
		}
		if len(hosts) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有可用的隧道"})
			return
		}

		var reqPort int
		if v := r.URL.Query().Get("port"); v != "" {
			var err error
			reqPort, err = strconv.Atoi(v)
			if err != nil || reqPort < 1 || reqPort > 65535 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "端口超出合法范围 (1 ~ 65535)"})
				return
			}
		}
		x, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		var ports []int
		if len(hosts) == 1 && reqPort > 0 {
			p, err := x.CloneToTunnelWithPort(id, hosts[0], reqPort, tunnels)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
				return
			}
			ports = []int{p}
		} else {
			ports, err = x.CloneToTunnels(id, hosts, tunnels)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error(), "created": ports})
				return
			}
		}
		invalidateInbounds()
		writeJSON(w, http.StatusOK, map[string]any{"created": ports})
	}
}

// apiPortCheck 实时探测指定端口在操作系统上是否真正空闲可用
func apiPortCheck(w http.ResponseWriter, r *http.Request) {
	port, err := strconv.Atoi(r.URL.Query().Get("port"))
	if err != nil || port < 1 || port > 65535 {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "reason": "端口超出合法范围 (1 ~ 65535)"})
		return
	}
	// 检查外部 xray 配置
	ext := externalUsedPorts()
	if ext[port] {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "reason": fmt.Sprintf("端口 %d 已被外部 Xray 配置占用", port)})
		return
	}
	// 检查当前 panel 占用的端口
	if p, err := openPanel(); err == nil {
		if inbounds, err := p.Inbounds(nil); err == nil {
			for _, ib := range inbounds {
				if ib.Port == port {
					writeJSON(w, http.StatusOK, map[string]any{"available": false, "reason": fmt.Sprintf("端口 %d 已被现有节点 %s 占用", port, ib.Remark)})
					return
				}
			}
		}
	}
	// 操作系统真实 bind 测试
	if !portAvailable(port) {
		writeJSON(w, http.StatusOK, map[string]any{"available": false, "reason": fmt.Sprintf("端口 %d 已被系统其它程序（如 Nginx/3x-ui/Web服务）占用", port)})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"available": true})
}

// apiXUIDetail 返回某个入站的详情，含客户端与可直接复制的分享链接。
func apiXUIDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id 参数无效"})
		return
	}
	x, err := openPanel()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	host := r.URL.Query().Get("host")
	if host == "" {
		host = publicHost(r)
	}
	detail, err := x.InboundDetail(id, host)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

// publicHost 决定分享链接里的连接地址。优先使用设置中绑定的域名；
// 未配置域名时，以母机公网 IPv4 为准；探测不到再退回访问 fanout 时用的主机名。
func publicHost(r *http.Request) string {
	cfg := getWebSettings()
	if cfg.Domain != "" {
		return cfg.Domain
	}
	if ip := hostPublicIP(); ip != "" {
		return ip
	}
	host := ""
	if r != nil {
		host = r.Host
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	if host == "" || host == "127.0.0.1" || host == "::1" || host == "localhost" {
		return "<服务器IP>"
	}
	return host
}

// ExportItem 描述一个入站的分享链接与其出口归属。
type ExportItem struct {
	ID         int          `json:"id"`
	Port       int          `json:"port"`
	Protocol   string       `json:"protocol"`
	Remark     string       `json:"remark"`
	IsDirect   bool         `json:"is_direct"`
	ExitRegion string       `json:"exit_region,omitempty"`
	ExitIP     string       `json:"exit_ip,omitempty"`
	ExitHost   string       `json:"exit_host,omitempty"`
	Clients    []ClientInfo `json:"clients,omitempty"`
	Links      []string     `json:"links"`
}

// apiXUILinks 批量导出多个入站的分享链接，并标注母机直连与出口归属。
func apiXUILinks(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}

		var ids []int
		if raw := r.URL.Query().Get("ids"); raw != "" {
			for _, part := range strings.Split(raw, ",") {
				if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
					ids = append(ids, n)
				}
			}
		} else {
			list, err := x.Inbounds(nil)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
				return
			}
			for _, ib := range list {
				if ib.Enable {
					ids = append(ids, ib.ID)
				}
			}
		}
		if len(ids) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有可导出的入站"})
			return
		}

		host := r.URL.Query().Get("host")
		if host == "" {
			host = publicHost(r)
		}

		// 构建出口归属映射
		type exitInfo struct {
			region string
			ip     string
			host   string
		}
		inboundExitMap := make(map[int]exitInfo)
		if m != nil {
			ev := m.ExitsOf()
			for _, ex := range ev.Exits {
				for _, ib := range ex.Inbounds {
					inboundExitMap[ib.ID] = exitInfo{
						region: ex.Region,
						ip:     ex.ExitIP,
						host:   ex.Host,
					}
				}
			}
		}

		allLinks := make([]string, 0)
		items := make([]ExportItem, 0, len(ids))
		for _, id := range ids {
			detail, err := x.InboundDetail(id, host)
			if err != nil {
				continue
			}
			allLinks = append(allLinks, detail.Links...)
			ex, isExit := inboundExitMap[id]
			items = append(items, ExportItem{
				ID:         detail.ID,
				Port:       detail.Port,
				Protocol:   detail.Protocol,
				Remark:     detail.Remark,
				IsDirect:   !isExit,
				ExitRegion: ex.region,
				ExitIP:     ex.ip,
				ExitHost:   ex.host,
				Clients:    detail.Clients,
				Links:      detail.Links,
			})
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"links": allLinks,
			"items": items,
		})
	}
}

// apiXUIDelete 删除入站。停掉出口后它的入站会留下来，用户需要一个清理入口。
func apiXUIDelete(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ids []int
		for _, part := range strings.Split(r.URL.Query().Get("ids"), ",") {
			if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
				ids = append(ids, n)
			}
		}
		if len(ids) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有指定要删除的入站"})
			return
		}
		x, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		err = x.DeleteInbounds(ids, m.Tunnels())
		invalidateInbounds()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"deleted": len(ids)})
	}
}

// apiInboundCreate 新建一个入站。两种后端都支持：自建模式写自己的库，
// 接管 3x-ui 时走面板的 inbounds/add API。
// apiInboundUpdate 改入站的端口、备注与启停。两种后端都支持。
func apiInboundUpdate(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		q := r.URL.Query()
		id, err := strconv.Atoi(q.Get("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id 参数无效"})
			return
		}

		// 只有真正传了的参数才改，没传的保持原样
		var patch InboundPatch
		if v := q.Get("port"); v != "" {
			port, err := strconv.Atoi(v)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "端口无效"})
				return
			}
			if port < 1 || port > 65535 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "端口超出合法范围 (1 ~ 65535)"})
				return
			}
			patch.Port = &port
		}
		if q.Has("remark") {
			remark := q.Get("remark")
			patch.Remark = &remark
		}
		if v := q.Get("enable"); v != "" {
			enable := v == "1"
			patch.Enable = &enable
		}

		err = p.UpdateInbound(id, patch, m.Tunnels())
		invalidateInbounds()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"ok": "已保存"})
	}
}

// clientAction 把三个客户端操作的公共部分收拢：解析 id/email 再调后端。
func clientAction(m *Manager, what string,
	do func(p Panel, id int, email string, tunnels []*Tunnel) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		id, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id 参数无效"})
			return
		}
		err = do(p, id, r.URL.Query().Get("email"), m.Tunnels())
		invalidateInbounds()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"ok": what})
	}
}

func apiClientAdd(m *Manager) http.HandlerFunc {
	return clientAction(m, "已添加", func(p Panel, id int, email string, t []*Tunnel) error {
		return p.AddClient(id, email, t)
	})
}

func apiClientDelete(m *Manager) http.HandlerFunc {
	return clientAction(m, "已删除", func(p Panel, id int, email string, t []*Tunnel) error {
		return p.DeleteClient(id, email, t)
	})
}

func apiClientReset(m *Manager) http.HandlerFunc {
	return clientAction(m, "已重置", func(p Panel, id int, email string, t []*Tunnel) error {
		return p.ResetClient(id, email, t)
	})
}

func apiInboundCreate(m *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, err := openPanel()
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}

		q := r.URL.Query()
		var port int
		if v := q.Get("port"); v != "" {
			var err error
			port, err = strconv.Atoi(v)
			if err != nil || port < 1 || port > 65535 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "端口超出合法范围 (1 ~ 65535)"})
				return
			}
		}
		ib, err := p.CreateInbound(NewInboundSpec{
			Protocol: q.Get("protocol"),
			Network:  q.Get("network"),
			Port:     port,
			Remark:   q.Get("remark"),
			Path:     q.Get("path"),
			Host:     q.Get("host"),
			Security: q.Get("security"),
			Vision:   q.Get("vision") == "1",

			ServerName: q.Get("sni"),
			CertFile:   q.Get("cert"),
			KeyFile:    q.Get("key"),

			Dest:        q.Get("dest"),
			ServerNames: q.Get("server_names"),
			ShortID:     q.Get("sid"),
			Fingerprint: q.Get("fp"),
		}, m.Tunnels())
		invalidateInbounds()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":       ib.ID,
			"port":     ib.Port,
			"protocol": ib.Protocol,
			"remark":   ib.Remark,
			"network":  ib.Network,
			"security": ib.Security,
		})
	}
}
