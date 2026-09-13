package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNormalizeListenAddr(t *testing.T) {
	cases := map[string]string{
		"":          "",
		"0.0.0.0":   "",
		"all":       "",
		"127.0.0.1": "127.0.0.1",
	}
	for in, want := range cases {
		got, err := normalizeListenAddr(in)
		if err != nil {
			t.Fatalf("normalizeListenAddr(%q) 意外报错: %v", in, err)
		}
		if got != want {
			t.Fatalf("normalizeListenAddr(%q)=%q，想要 %q", in, got, want)
		}
	}
	if _, err := normalizeListenAddr("not-an-ip"); err == nil {
		t.Fatal("非法监听地址应当报错")
	}
}

func TestValidatePort(t *testing.T) {
	for _, p := range []int{1, 8899, 65535} {
		if err := validatePort(p); err != nil {
			t.Fatalf("端口 %d 应合法: %v", p, err)
		}
	}
	for _, p := range []int{0, -1, 70000} {
		if err := validatePort(p); err == nil {
			t.Fatalf("端口 %d 应非法", p)
		}
	}
}

func TestSetBasePathValidatesAndPersists(t *testing.T) {
	dir := t.TempDir()
	if _, err := initBasePath(dir); err != nil {
		t.Fatalf("initBasePath: %v", err)
	}
	bp, err := setBasePath("myPanel_1")
	if err != nil {
		t.Fatalf("setBasePath: %v", err)
	}
	if bp != "/myPanel_1" || currentBasePath() != "/myPanel_1" {
		t.Fatalf("basePath 未生效: %q / %q", bp, currentBasePath())
	}
	if _, err := os.ReadFile(dir + "/basepath"); err != nil {
		t.Fatalf("basepath 未落盘: %v", err)
	}
	if _, err := setBasePath("bad/slash"); err == nil {
		t.Fatal("带非法字符的路径应被拒")
	}
	// 空串表示去掉前缀
	if bp, err := setBasePath(""); err != nil || bp != "" {
		t.Fatalf("空路径应清空前缀: %q %v", bp, err)
	}
}

func TestAuthSetPassword(t *testing.T) {
	dir := t.TempDir()
	auth, _, err := NewAuth(dir)
	if err != nil {
		t.Fatalf("NewAuth: %v", err)
	}
	if err := auth.SetPassword("newsecret"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}
	if !auth.check("newsecret") {
		t.Fatal("新口令应校验通过")
	}
	if auth.check("wrong") {
		t.Fatal("旧口令不应再通过")
	}
	if err := auth.SetPassword(""); err == nil {
		t.Fatal("空口令应被拒")
	}
	if err := auth.SetPassword("ab"); err == nil {
		t.Fatal("过短口令应被拒")
	}
}

func TestWebServerReloadSwitchesPort(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadWebSettings(dir, 0, false); err != nil {
		t.Fatalf("loadWebSettings: %v", err)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	srv := newWebServer(h)

	// 用两个系统分配的空闲端口验证切换
	p1 := freePort(t)
	if err := srv.reload(WebSettings{Port: p1, ListenAddr: "127.0.0.1"}); err != nil {
		t.Fatalf("reload p1: %v", err)
	}
	waitServe(t, p1)

	p2 := freePort(t)
	if err := srv.applyWebSettings(WebSettings{Port: p2, ListenAddr: "127.0.0.1"}); err != nil {
		t.Fatalf("applyWebSettings p2: %v", err)
	}
	waitServe(t, p2)

	// 旧端口应在优雅关闭后不再接受连接
	time.Sleep(1500 * time.Millisecond)
	if c, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p1)), 300*time.Millisecond); err == nil {
		c.Close()
		t.Fatalf("旧端口 %d 切换后仍在监听", p1)
	}

	// 非法端口应被拒，且不影响现有监听
	if err := srv.applyWebSettings(WebSettings{Port: 70000, ListenAddr: "127.0.0.1"}); err == nil {
		t.Fatal("非法端口应被拒")
	}
	waitServe(t, p2)
}

func waitServe(t *testing.T, port int) {
	t.Helper()
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
	for i := 0; i < 40; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("端口 %d 未在预期时间内提供服务", port)
}

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("取空闲端口失败: %v", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

// 界面上改过端口会落盘。之后再带 -web 启动时，命令行必须说话算话，
// 否则用户敲了参数却连不上，还没有任何提示（ct-54 上真实踩到）。
func TestLoadWebSettingsExplicitFlagWins(t *testing.T) {
	dir := t.TempDir()

	// 首次启动：建档存 8899
	if _, err := loadWebSettings(dir, 8899, false); err != nil {
		t.Fatalf("首次: %v", err)
	}

	// 不带 -web 重启：沿用盘上的 8899，不被默认值覆盖
	s, err := loadWebSettings(dir, 8899, false)
	if err != nil {
		t.Fatalf("沿用: %v", err)
	}
	if s.Port != 8899 {
		t.Fatalf("没显式指定时应沿用盘上的值，实际 %d", s.Port)
	}

	// 显式 -web 80：以命令行为准
	s, err = loadWebSettings(dir, 80, true)
	if err != nil {
		t.Fatalf("显式指定: %v", err)
	}
	if s.Port != 80 {
		t.Fatalf("显式 -web 应压过盘上的值，实际 %d", s.Port)
	}

	// 且要落盘，下次不带参数启动仍是 80
	s, err = loadWebSettings(dir, 8899, false)
	if err != nil {
		t.Fatalf("复读: %v", err)
	}
	if s.Port != 80 {
		t.Fatalf("显式指定的端口应已写回，实际 %d", s.Port)
	}
}

func TestNormalizeDomain(t *testing.T) {
	validCases := map[string]string{
		"example.com":                         "example.com",
		"  sub.EXAMPLE.com  ":                 "sub.example.com",
		"https://panel.example.com/":          "panel.example.com",
		"http://PANEL.example.com:8443/path":  "panel.example.com",
		"my-domain.net":                       "my-domain.net",
		"a.b.c.d.org":                         "a.b.c.d.org",
		"localhost":                           "localhost",
		"127.0.0.1":                           "127.0.0.1",
		"https://[::1]:8899/foo":              "::1",
		"trailing.dot.com.":                   "trailing.dot.com",
		"multi.dot.com...":                    "multi.dot.com",
		"":                                    "",
	}
	for in, want := range validCases {
		got, err := normalizeDomain(in)
		if err != nil {
			t.Errorf("normalizeDomain(%q) 报错: %v", in, err)
		}
		if got != want {
			t.Errorf("normalizeDomain(%q) = %q, want %q", in, got, want)
		}
	}

	invalidCases := []string{
		"-bad.com",
		"bad-.com",
		"bad domain.com",
		"bad@domain.com",
		"bad_domain.com",
		"*.example.com",
		strings.Repeat("a", 64) + ".com",
		strings.Repeat("a.", 130) + "com",
	}
	for _, in := range invalidCases {
		if _, err := normalizeDomain(in); err == nil {
			t.Errorf("normalizeDomain(%q) 应报错但未报错", in)
		}
	}
}

func generateTestCert(t *testing.T, dir, cn string, hosts []string, validFor time.Duration) (string, string) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}

	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("serialNumber: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Fanout Test Org"},
			CommonName:   cn,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("CreateCertificate: %v", err)
	}

	certPath := filepath.Join(dir, cn+".crt")
	certOut, err := os.Create(certPath)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	_ = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey: %v", err)
	}
	keyPath := filepath.Join(dir, cn+".key")
	keyOut, err := os.Create(keyPath)
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	_ = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	keyOut.Close()

	return certPath, keyPath
}

func TestInspectCertificate(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := generateTestCert(t, dir, "panel.local", []string{"panel.local", "127.0.0.1"}, 30*24*time.Hour)

	details, err := inspectCertificate(certPath, keyPath)
	if err != nil {
		t.Fatalf("inspectCertificate 失败: %v", err)
	}

	if details.SubjectCN != "panel.local" {
		t.Errorf("SubjectCN = %q, want panel.local", details.SubjectCN)
	}
	if details.DaysLeft <= 0 {
		t.Errorf("DaysLeft = %d, 应大于 0", details.DaysLeft)
	}
	if details.IsExpired {
		t.Errorf("证书不应判定为过期")
	}
	hasLocal := false
	for _, s := range details.SANs {
		if s == "panel.local" || s == "127.0.0.1" {
			hasLocal = true
		}
	}
	if !hasLocal {
		t.Errorf("SANs 中应包含 panel.local 或 127.0.0.1: %v", details.SANs)
	}

	// 测试尚有 12 小时到期的证书：应向上取整为 1 天，且判定为未过期
	soonCert, soonKey := generateTestCert(t, dir, "soon.local", []string{"soon.local"}, 12*time.Hour)
	soonDetails, err := inspectCertificate(soonCert, soonKey)
	if err != nil {
		t.Fatalf("inspect soon cert 报错: %v", err)
	}
	if soonDetails.IsExpired {
		t.Errorf("还有 12 小时的证书不应判定为已过期")
	}
	if soonDetails.DaysLeft != 1 {
		t.Errorf("还有 12 小时的证书 DaysLeft 应为 1，实际 %d", soonDetails.DaysLeft)
	}

	// 测试过期证书
	expCert, expKey := generateTestCert(t, dir, "expired.local", []string{"expired.local"}, -2*time.Hour)
	expDetails, err := inspectCertificate(expCert, expKey)
	if err != nil {
		t.Fatalf("inspect expired cert 报错: %v", err)
	}
	if !expDetails.IsExpired {
		t.Errorf("过期证书应判定为 IsExpired=true")
	}
	if expDetails.DaysLeft != 0 {
		t.Errorf("已过期证书 DaysLeft 应为 0，实际 %d", expDetails.DaysLeft)
	}

	// 测试路径不存在
	if _, err := inspectCertificate(filepath.Join(dir, "none.crt"), keyPath); err == nil {
		t.Errorf("证书不存在时应当报错")
	}
	// 测试私钥配对错误
	otherCert, _ := generateTestCert(t, dir, "other.local", []string{"other.local"}, 24*time.Hour)
	if _, err := inspectCertificate(otherCert, keyPath); err == nil {
		t.Errorf("公私钥不匹配时应当报错")
	}
}

func TestWebServerSamePortTLSSwitch(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadWebSettings(dir, 0, false); err != nil {
		t.Fatalf("loadWebSettings: %v", err)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fanout-test-response"))
	})
	srv := newWebServer(h)

	p := freePort(t)
	// 1. 先以 HTTP 启动
	if err := srv.reload(WebSettings{Port: p, ListenAddr: "127.0.0.1"}); err != nil {
		t.Fatalf("reload HTTP: %v", err)
	}
	waitServe(t, p)
	if srv.IsTLS() {
		t.Fatal("启动时应为 HTTP 模式")
	}

	// 2. 同端口切换为 HTTPS (原生 TLS)
	certPath, keyPath := generateTestCert(t, dir, "localhost", []string{"127.0.0.1", "localhost"}, 24*time.Hour)
	tlsSettings := WebSettings{
		Port:       p,
		ListenAddr: "127.0.0.1",
		CertFile:   certPath,
		KeyFile:    keyPath,
		SSLMode:    "custom",
	}
	if err := srv.applyWebSettings(tlsSettings); err != nil {
		t.Fatalf("applyWebSettings to HTTPS: %v", err)
	}

	// 等待并验证 HTTPS 服务可用
	waitServeTLS(t, p)
	if !srv.IsTLS() {
		t.Fatal("切换后应为 TLS 模式")
	}

	// 此时尝试以明文 HTTP 访问该端口应失败
	httpClient := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
		Timeout:   300 * time.Millisecond,
	}
	// 此时尝试以明文 HTTP 访问该端口，应报错或返回 400 Bad Request（Client sent an HTTP request to an HTTPS server）
	if resp, err := httpClient.Get("http://127.0.0.1:" + strconv.Itoa(p) + "/"); err == nil {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			t.Fatal("HTTPS 模式下明文 HTTP 请求不应正常返回 200 OK")
		}
	}

	// 3. 同端口切回 HTTP 明文
	httpSettings := WebSettings{
		Port:       p,
		ListenAddr: "127.0.0.1",
		SSLMode:    "none",
	}
	if err := srv.applyWebSettings(httpSettings); err != nil {
		t.Fatalf("applyWebSettings back to HTTP: %v", err)
	}

	waitServe(t, p)
	if srv.IsTLS() {
		t.Fatal("切回后应为 HTTP 模式")
	}
}

func waitServeTLS(t *testing.T, port int) {
	t.Helper()
	urlStr := "https://127.0.0.1:" + strconv.Itoa(port) + "/"
	tr := &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 500 * time.Millisecond}
	for i := 0; i < 40; i++ {
		resp, err := client.Get(urlStr)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("HTTPS 端口 %d 未在预期时间内提供服务", port)
}

func TestCertificateDynamicReload(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadWebSettings(dir, 0, false); err != nil {
		t.Fatalf("loadWebSettings: %v", err)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	srv := newWebServer(h)

	p := freePort(t)
	certA, keyA := generateTestCert(t, dir, "cert-a.example.com", []string{"127.0.0.1", "cert-a.example.com"}, 24*time.Hour)

	if err := srv.reload(WebSettings{
		Port:       p,
		ListenAddr: "127.0.0.1",
		CertFile:   certA,
		KeyFile:    keyA,
		SSLMode:    "custom",
	}); err != nil {
		t.Fatalf("reload with certA: %v", err)
	}
	waitServeTLS(t, p)

	// 连接检查当前证书为 cert-a
	dialTLS := func() string {
		conn, err := tls.Dial("tcp", "127.0.0.1:"+strconv.Itoa(p), &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Fatalf("tls.Dial: %v", err)
		}
		defer conn.Close()
		state := conn.ConnectionState()
		if len(state.PeerCertificates) == 0 {
			t.Fatal("未收到对端证书")
		}
		return state.PeerCertificates[0].Subject.CommonName
	}

	if cn := dialTLS(); cn != "cert-a.example.com" {
		t.Fatalf("当前证书 CN = %q, want cert-a.example.com", cn)
	}

	// 动态更新到 cert-b (地址不变，仅更新证书路径/指针)
	certB, keyB := generateTestCert(t, dir, "cert-b.example.com", []string{"127.0.0.1", "cert-b.example.com"}, 24*time.Hour)
	if err := srv.applyWebSettings(WebSettings{
		Port:       p,
		ListenAddr: "127.0.0.1",
		CertFile:   certB,
		KeyFile:    keyB,
		SSLMode:    "custom",
	}); err != nil {
		t.Fatalf("applyWebSettings with certB: %v", err)
	}

	// 再次连接应立即获取到新证书 cert-b，实现 0ms 无缝热重载
	if cn := dialTLS(); cn != "cert-b.example.com" {
		t.Fatalf("动态更新后证书 CN = %q, want cert-b.example.com", cn)
	}
}

func TestPublicHostAndLinkPortDomain(t *testing.T) {
	dir := t.TempDir()
	if _, err := loadWebSettings(dir, 8899, false); err != nil {
		t.Fatalf("loadWebSettings: %v", err)
	}

	// 未配置域名时
	req := &http.Request{Host: "192.168.1.100:8899"}
	host := publicHost(req)
	if host == "panel.test.com" {
		t.Fatal("未配置域名时不应返回域名")
	}

	// 配置域名后
	webSettingsMu.Lock()
	webSettingsCur.Domain = "panel.test.com"
	webSettingsMu.Unlock()

	host = publicHost(req)
	if host != "panel.test.com" {
		t.Fatalf("配置域名后 publicHost 应返回 panel.test.com，实际 %q", host)
	}

	// 测试 linkForPort 对各种旧主机名（localhost、IP、旧域名）的正则替换
	cases := []struct {
		in         string
		port       int
		publicHost string
		want       string
	}{
		{
			in:         "vless://uuid-1@localhost:20001?security=none#HK@Node:20001",
			port:       20001,
			publicHost: "panel.test.com",
			want:       "vless://uuid-1@panel.test.com:20001?security=none#HK@Node:20001",
		},
		{
			in:         "vless://uuid-2@[::1]:20002?security=none#test2",
			port:       20002,
			publicHost: "panel.test.com",
			want:       "vless://uuid-2@panel.test.com:20002?security=none#test2",
		},
		{
			in:         "vless://uuid-1@localhost:20001?security=none#test1",
			port:       20001,
			publicHost: "2001:db8::1",
			want:       "vless://uuid-1@[2001:db8::1]:20001?security=none#test1",
		},
		{
			in:         "trojan://pass@1.2.3.4:20003?security=tls#test3",
			port:       20003,
			publicHost: "panel.test.com",
			want:       "trojan://pass@panel.test.com:20003?security=tls#test3",
		},
		{
			in:         "ss://base64@old.domain.com:20004#My@Remark:20004",
			port:       20004,
			publicHost: "panel.test.com",
			want:       "ss://base64@panel.test.com:20004#My@Remark:20004",
		},
		{
			in:         "vmess://uuid-uri@old.domain.com:20006?remarks=test",
			port:       20006,
			publicHost: "panel.test.com",
			want:       "vmess://uuid-uri@panel.test.com:20006?remarks=test",
		},
	}

	for _, tc := range cases {
		got, ok := linkForPort(tc.in, tc.port, tc.publicHost)
		if !ok {
			t.Fatalf("linkForPort(%q, port=%d) 匹配失败", tc.in, tc.port)
		}
		if got != tc.want {
			t.Fatalf("linkForPort 替换结果错误:\n got:  %s\n want: %s", got, tc.want)
		}
	}

	// 测试子串端口防误判（端口 20001 不应被判定为属于端口 200）
	if _, ok := linkForPort("vless://uuid-1@localhost:20001?security=none", 200, "panel.test.com"); ok {
		t.Fatal("子串端口 200 不应匹配到 20001 的链接")
	}

	// 测试 VMess JSON 链接的主机替换
	vmessConf := map[string]any{
		"v": "2", "ps": "vmess-test", "add": "10.0.0.1", "port": 20005, "id": "uuid-vmess",
	}
	blob, _ := json.Marshal(vmessConf)
	vmessLink := "vmess://" + base64.StdEncoding.EncodeToString(blob)
	fixedVMess, ok := linkForPort(vmessLink, 20005, "panel.test.com")
	if !ok {
		t.Fatal("VMess linkForPort 匹配失败")
	}
	dec, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(fixedVMess, "vmess://"))
	var gotConf map[string]any
	if err := json.Unmarshal(dec, &gotConf); err != nil {
		t.Fatalf("VMess 解码 JSON 失败: %v", err)
	}
	if gotConf["add"] != "panel.test.com" {
		t.Fatalf("VMess add 字段未替换为域名，实际 %v", gotConf["add"])
	}
}

func TestSafeFallbackDegradesToHTTP(t *testing.T) {
	dir := t.TempDir()
	p := freePort(t)

	// 直接在 settings.json 中写入损坏/不存在的证书路径
	badSettings := WebSettings{
		Port:       p,
		ListenAddr: "127.0.0.1",
		CertFile:   "/path/to/nonexistent.crt",
		KeyFile:    "/path/to/nonexistent.key",
		SSLMode:    "custom",
	}
	blob, _ := json.Marshal(badSettings)
	settingsFile := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(settingsFile, blob, 0600); err != nil {
		t.Fatalf("写坏配置失败: %v", err)
	}

	if _, err := loadWebSettings(dir, p, false); err != nil {
		t.Fatalf("loadWebSettings: %v", err)
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fallback-ok"))
	})
	srv := newWebServer(h)

	go func() {
		_ = srv.serve()
	}()

	waitServe(t, p)
	if srv.IsTLS() {
		t.Fatal("启动时证书损坏应当降级为 HTTP，不应是 TLS")
	}

	// 核心断言：安全降级决不能抹除磁盘上的 settings.json 证书路径
	diskBytes, err := os.ReadFile(settingsFile)
	if err != nil {
		t.Fatalf("读取磁盘 settings.json 失败: %v", err)
	}
	var onDisk WebSettings
	if err := json.Unmarshal(diskBytes, &onDisk); err != nil {
		t.Fatalf("解析磁盘 settings.json 失败: %v", err)
	}
	if onDisk.CertFile != badSettings.CertFile || onDisk.KeyFile != badSettings.KeyFile {
		t.Fatalf("安全降级不应抹除磁盘证书配置: cert=%q, key=%q", onDisk.CertFile, onDisk.KeyFile)
	}
}
