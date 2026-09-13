package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestRenameExitSuffix(t *testing.T) {
	cases := []struct{ remark, label, want string }{
		{"KR-248", "JP-132", "JP-132"},
		{"线路A-KR-248", "JP-132", "线路A-JP-132"},
		{"inbound-47525-KR-248", "JP-132", "inbound-47525-JP-132"},
		{"无格式", "JP-132", "无格式"},
		{"", "JP-132", ""},
	}
	for _, c := range cases {
		got := renameExitSuffix(c.remark, c.label)
		if got != c.want {
			t.Errorf("renameExitSuffix(%q) = %q, want %q", c.remark, got, c.want)
		}
	}
}

func TestResolvedInboundTagPrefersAPITag(t *testing.T) {
	stream := json.RawMessage(`{"network":"ws"}`)
	got := resolvedInboundTag("in-12080-tcp", 12080, stream)
	if got != "in-12080-tcp" {
		t.Fatalf("resolvedInboundTag() = %q, want API tag %q", got, "in-12080-tcp")
	}
}

func TestResolvedInboundTagFallsBackForLegacyAPI(t *testing.T) {
	stream := json.RawMessage(`{"network":"ws"}`)
	got := resolvedInboundTag("", 12080, stream)
	if got != "in-12080-ws" {
		t.Fatalf("resolvedInboundTag() = %q, want reconstructed tag %q", got, "in-12080-ws")
	}
}

// 面板没开 SSL 时会打印 "Warning: Panel is not secure with SSL"，
// 它包含 "Panel is secure with SSL" 这个子串，曾被误判成 https（issue #8）。
func TestXUISSLFromSettings(t *testing.T) {
	const off = `current panel settings as follows:
Warning: Panel is not secure with SSL
hasDefaultCredential: false
port: 37285
webBasePath: /abc123/
`
	const on = `current panel settings as follows:
Panel is secure with SSL
port: 2053
webBasePath: /xyz/
`
	const silent = `current panel settings as follows:
port: 2053
webBasePath: /xyz/
`
	cases := []struct {
		name       string
		text       string
		wantOn     bool
		wantStated bool
	}{
		{"未启用 SSL", off, false, true},
		{"已启用 SSL", on, true, true},
		{"没提 SSL", silent, false, false},
	}
	for _, c := range cases {
		on, stated := xuiSSLFromSettings(c.text)
		if on != c.wantOn || stated != c.wantStated {
			t.Errorf("%s: xuiSSLFromSettings() = (%v,%v), want (%v,%v)",
				c.name, on, stated, c.wantOn, c.wantStated)
		}
	}
}

// cert 为空时正则里的 \s* 会跨行，把下一行的 "key:" 当成证书路径（issue #8）。
func TestXUICertConfigured(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"证书为空", "cert:\nkey:\n", false},
		{"证书为空带空格", "cert: \nkey: \n", false},
		{"证书已配置", "cert: /root/cert.crt\nkey: /root/private.key\n", true},
	}
	for _, c := range cases {
		if got := xuiCertConfigured(c.text); got != c.want {
			t.Errorf("%s: xuiCertConfigured() = %v, want %v", c.name, got, c.want)
		}
	}
}

// 端口/路径的取值同样不能跨行。
func TestXUISettingFieldsStayOnOwnLine(t *testing.T) {
	const text = `current panel settings as follows:
Warning: Panel is not secure with SSL
hasDefaultCredential: false
port: 37285
webBasePath: /abc123/
`
	pm := reXUIPort.FindStringSubmatch(text)
	if pm == nil || pm[1] != "37285" {
		t.Fatalf("reXUIPort 解析失败: %v", pm)
	}
	bm := reXUIBase.FindStringSubmatch(text)
	if bm == nil || bm[1] != "/abc123/" {
		t.Fatalf("reXUIBase 解析失败: %v", bm)
	}
	// 值为空时宁可解析不出（调用方会报错），也不能把下一行当成路径
	if m := reXUIBase.FindStringSubmatch("webBasePath:\nport: 1\n"); m != nil {
		t.Fatalf("webBasePath 为空时不应吃掉下一行: %v", m)
	}
}

// vmess 的分享链接是 base64 编码的 JSON，按 ":端口?" 匹配一条也筛不出来。
func TestLinkForPortHandlesVMess(t *testing.T) {
	conf := map[string]any{
		"v": "2", "ps": "t-ws", "add": "localhost", "port": 40978,
		"id": "06a30cf8-2c06-4aeb-82fe-b4c7f1ab0159", "net": "ws", "path": "/abc",
	}
	blob, _ := json.Marshal(conf)
	link := "vmess://" + base64.StdEncoding.EncodeToString(blob)

	fixed, ok := linkForPort(link, 40978, "1.2.3.4")
	if !ok {
		t.Fatal("端口对得上却被筛掉了")
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(fixed, "vmess://"))
	if err != nil {
		t.Fatalf("解码失败: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(decoded, &got); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if got["add"] != "1.2.3.4" {
		t.Errorf("add = %v, want 1.2.3.4", got["add"])
	}
	if int(toFloat(got["port"])) != 40978 {
		t.Errorf("端口被改坏了: %v", got["port"])
	}

	if _, ok := linkForPort(link, 12345, "1.2.3.4"); ok {
		t.Error("端口对不上时不该返回")
	}
}

func TestLinkForPortHandlesURIStyle(t *testing.T) {
	link := "vless://uuid@localhost:26387?security=none&type=tcp#t-tcp"
	fixed, ok := linkForPort(link, 26387, "1.2.3.4")
	if !ok {
		t.Fatal("端口对得上却被筛掉了")
	}
	if !strings.Contains(fixed, "@1.2.3.4:26387") {
		t.Errorf("地址没换: %s", fixed)
	}
	if _, ok := linkForPort(link, 99999, "1.2.3.4"); ok {
		t.Error("端口对不上时不该返回")
	}
}

func TestExtractPortFromTag(t *testing.T) {
	cases := []struct {
		tag  string
		want int
	}{
		{"inbound-443", 443},
		{"in-443-tcp", 443},
		{"in-8443-ws", 8443},
		{"443", 443},
		{"2053", 2053},
		{"vless-node", 0},
		{"", 0},
	}
	for _, c := range cases {
		got := extractPortFromTag(c.tag)
		if got != c.want {
			t.Errorf("extractPortFromTag(%q) = %d, want %d", c.tag, got, c.want)
		}
	}
}

func TestFindBoundHost(t *testing.T) {
	bound := map[string]string{
		"inbound-443": "jp-01",
		"in-8443-tcp": "us-02",
		"custom-2053": "kr-03",
	}

	// 1. 精准匹配 tag
	if got := findBoundHost(bound, "inbound-443", "", 443); got != "jp-01" {
		t.Errorf("findBoundHost exact tag failed: got %q, want jp-01", got)
	}

	// 2. tag 形式不同但端口一致 (如 tag 为 in-443-tcp，规则里为 inbound-443)
	if got := findBoundHost(bound, "in-443-tcp", "", 443); got != "jp-01" {
		t.Errorf("findBoundHost port alias fallback failed: got %q, want jp-01", got)
	}

	// 3. tag 为纯端口号 "443"
	if got := findBoundHost(bound, "443", "", 443); got != "jp-01" {
		t.Errorf("findBoundHost pure port fallback failed: got %q, want jp-01", got)
	}

	// 4. apiTag 匹配
	if got := findBoundHost(bound, "random-tag", "in-8443-tcp", 8443); got != "us-02" {
		t.Errorf("findBoundHost apiTag fallback failed: got %q, want us-02", got)
	}

	// 5. 后缀端口匹配
	if got := findBoundHost(bound, "unrelated", "", 2053); got != "kr-03" {
		t.Errorf("findBoundHost suffix port fallback failed: got %q, want kr-03", got)
	}

	// 6. 未绑定入站
	if got := findBoundHost(bound, "inbound-9999", "", 9999); got != "" {
		t.Errorf("findBoundHost unbound inbound should return empty, got %q", got)
	}
}

func TestXUIBindAndUnbindTolerance(t *testing.T) {
	currentSetting := map[string]any{
		"outbounds": []any{
			map[string]any{"tag": "direct", "protocol": "freedom"},
		},
		"routing": map[string]any{
			"rules": []any{},
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/panel/api/inbounds/list"):
			inbounds := []map[string]any{
				{
					"id":             1,
					"port":           443,
					"protocol":       "vless",
					"remark":         "reality-443",
					"enable":         true,
					"tag":            "inbound-443",
					"streamSettings": `{"network":"tcp"}`,
				},
			}
			resp := map[string]any{"success": true, "msg": "ok", "obj": inbounds}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.HasSuffix(path, "/panel/api/xray/update"):
			_ = r.ParseForm()
			settingStr := r.FormValue("xraySetting")
			var newSetting map[string]any
			if err := json.Unmarshal([]byte(settingStr), &newSetting); err == nil {
				currentSetting = newSetting
			}
			resp := map[string]any{"success": true, "msg": "updated"}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.HasSuffix(path, "/panel/api/xray/"):
			settingBlob, _ := json.Marshal(currentSetting)
			cfg := xrayConfig{
				OutboundTestURL: "http://test.url",
				XraySetting:     settingBlob,
			}
			cfgBlob, _ := json.Marshal(cfg)
			resp := map[string]any{"success": true, "msg": "ok", "obj": string(cfgBlob)}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.HasSuffix(path, "/panel/api/server/restartXrayService"):
			resp := map[string]any{"success": true, "msg": "restarted"}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	u, _ := url.Parse(mockServer.URL)
	port, _ := strconv.Atoi(u.Port())
	x := &XUI{
		Host:     u.Hostname(),
		Port:     port,
		BasePath: "",
		Scheme:   "http",
		token:    "test-token",
		client:   mockServer.Client(),
	}

	tunnels := []*Tunnel{
		{Slot: 1, Port: 20001, Status: "up", Node: Node{HostName: "jp-01"}},
		{Slot: 2, Port: 20002, Status: "up", Node: Node{HostName: "us-02"}},
	}

	// 1. 初始状态：未绑定（直连）
	inbounds, err := x.Inbounds(map[string]bool{"jp-01": true, "us-02": true})
	if err != nil {
		t.Fatalf("Inbounds failed: %v", err)
	}
	if inbounds[0].BoundTo != "" {
		t.Fatalf("Expected initial inbound to be unbound, got %q", inbounds[0].BoundTo)
	}

	// 2. 绑定 443 到 jp-01
	if err := x.Bind("inbound-443", "jp-01", tunnels); err != nil {
		t.Fatalf("Bind to jp-01 failed: %v", err)
	}

	// 检查绑定状态
	inbounds, err = x.Inbounds(map[string]bool{"jp-01": true, "us-02": true})
	if err != nil {
		t.Fatalf("Inbounds failed: %v", err)
	}
	if inbounds[0].BoundTo != "jp-01" {
		t.Fatalf("Expected BoundTo to be jp-01, got %q", inbounds[0].BoundTo)
	}

	// 3. 使用别名 "in-443-tcp" 配合 "direct" 解除绑定（模拟多别名容差解绑恢复直连）
	if err := x.Bind("in-443-tcp", "direct", tunnels); err != nil {
		t.Fatalf("Bind to direct failed: %v", err)
	}

	// 检查 routing rules 必须完全清空 fanout 规则
	routing, _ := currentSetting["routing"].(map[string]any)
	rules, _ := routing["rules"].([]any)
	for _, r := range rules {
		m := r.(map[string]any)
		tag, _ := m["outboundTag"].(string)
		if strings.HasPrefix(tag, "fanout-") {
			t.Fatalf("Unbind failed: fanout rule still exists: %v", m)
		}
	}

	// 再次查询 Inbounds，必须确认已解除绑定（直连）
	inbounds, err = x.Inbounds(map[string]bool{"jp-01": true, "us-02": true})
	if err != nil {
		t.Fatalf("Inbounds failed: %v", err)
	}
	if inbounds[0].BoundTo != "" {
		t.Fatalf("Expected BoundTo to be empty after unbind, got %q", inbounds[0].BoundTo)
	}

	// 4. 使用纯端口 "443" 重新绑定到 us-02
	if err := x.Bind("443", "us-02", tunnels); err != nil {
		t.Fatalf("Bind using port '443' failed: %v", err)
	}

	inbounds, err = x.Inbounds(map[string]bool{"jp-01": true, "us-02": true})
	if err != nil {
		t.Fatalf("Inbounds failed: %v", err)
	}
	if inbounds[0].BoundTo != "us-02" {
		t.Fatalf("Expected BoundTo to be us-02, got %q", inbounds[0].BoundTo)
	}

	// 验证 rule 中包含多版本容差 tag (如 inbound-443 与 in-443-tcp)
	routing, _ = currentSetting["routing"].(map[string]any)
	rules, _ = routing["rules"].([]any)
	var foundFanoutRule map[string]any
	for _, r := range rules {
		m := r.(map[string]any)
		if strings.HasPrefix(m["outboundTag"].(string), "fanout-") {
			foundFanoutRule = m
			break
		}
	}
	if foundFanoutRule == nil {
		t.Fatalf("Expected fanout rule to be added")
	}
	ruleTags := toStringSlice(foundFanoutRule["inboundTag"])
	hasInbound443, hasIn443TCP := false, false
	for _, tg := range ruleTags {
		if tg == "inbound-443" {
			hasInbound443 = true
		}
		if tg == "in-443-tcp" {
			hasIn443TCP = true
		}
	}
	if !hasInbound443 || !hasIn443TCP {
		t.Errorf("Expected routing rule to include both inbound-443 and in-443-tcp, got: %v", ruleTags)
	}

	// 5. 再次解绑（传空字符串 ""）
	if err := x.Bind("443", "", tunnels); err != nil {
		t.Fatalf("Bind empty unbind failed: %v", err)
	}
	inbounds, err = x.Inbounds(map[string]bool{"jp-01": true, "us-02": true})
	if err != nil {
		t.Fatalf("Inbounds failed: %v", err)
	}
	if inbounds[0].BoundTo != "" {
		t.Fatalf("Expected BoundTo to be empty after second unbind, got %q", inbounds[0].BoundTo)
	}
}

func TestXUIRebindTolerance(t *testing.T) {
	currentSetting := map[string]any{
		"outbounds": []any{
			map[string]any{"tag": "direct", "protocol": "freedom"},
		},
		"routing": map[string]any{
			"rules": []any{
				map[string]any{
					"type":        "field",
					"inboundTag":  []any{"inbound-443"},
					"outboundTag": "fanout-jp-01",
				},
			},
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case strings.HasSuffix(path, "/panel/api/inbounds/list"):
			inbounds := []map[string]any{
				{
					"id":             1,
					"port":           443,
					"protocol":       "vless",
					"remark":         "reality-443",
					"enable":         true,
					"tag":            "inbound-443",
					"streamSettings": `{"network":"tcp"}`,
				},
			}
			resp := map[string]any{"success": true, "msg": "ok", "obj": inbounds}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.HasSuffix(path, "/panel/api/xray/update"):
			_ = r.ParseForm()
			settingStr := r.FormValue("xraySetting")
			var newSetting map[string]any
			if err := json.Unmarshal([]byte(settingStr), &newSetting); err == nil {
				currentSetting = newSetting
			}
			resp := map[string]any{"success": true, "msg": "updated"}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.HasSuffix(path, "/panel/api/xray/"):
			settingBlob, _ := json.Marshal(currentSetting)
			cfg := xrayConfig{
				OutboundTestURL: "http://test.url",
				XraySetting:     settingBlob,
			}
			cfgBlob, _ := json.Marshal(cfg)
			resp := map[string]any{"success": true, "msg": "ok", "obj": string(cfgBlob)}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.HasSuffix(path, "/panel/api/server/restartXrayService"):
			resp := map[string]any{"success": true, "msg": "restarted"}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.Contains(path, "/panel/api/inbounds/get"):
			resp := map[string]any{"success": true, "msg": "ok", "obj": map[string]any{"id": 1, "remark": "reality-443"}}
			_ = json.NewEncoder(w).Encode(resp)

		case strings.Contains(path, "/panel/api/inbounds/update"):
			resp := map[string]any{"success": true, "msg": "ok"}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	u, _ := url.Parse(mockServer.URL)
	port, _ := strconv.Atoi(u.Port())
	x := &XUI{
		Host:     u.Hostname(),
		Port:     port,
		BasePath: "",
		Scheme:   "http",
		token:    "test-token",
		client:   mockServer.Client(),
	}

	target := &Tunnel{Slot: 2, Port: 20002, Status: "up", Node: Node{HostName: "us-02"}}
	tunnels := []*Tunnel{target}

	// 触发 Rebind，从 jp-01 换到 us-02
	if err := x.Rebind("jp-01", target, tunnels); err != nil {
		t.Fatalf("Rebind failed: %v", err)
	}

	inbounds, err := x.Inbounds(map[string]bool{"us-02": true})
	if err != nil {
		t.Fatalf("Inbounds failed: %v", err)
	}
	if inbounds[0].BoundTo != "us-02" {
		t.Errorf("Rebind should update BoundTo to us-02, got %q", inbounds[0].BoundTo)
	}
}
