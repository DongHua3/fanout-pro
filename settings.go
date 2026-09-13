package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// WebSettings 是管理界面自身的可改配置：监听端口、监听地址（本地/全接口）、
// 域名以及原生 TLS / SSL 证书配置。
// 落盘持久化，界面改完重启或热生效。访问口令与访问路径各有专门的文件
// （password / basepath），不放这里，但都能在设置面板里改。
type WebSettings struct {
	// Port 是管理界面监听端口。
	Port int `json:"port"`
	// ListenAddr 是监听地址：空或 0.0.0.0 表示所有网卡；127.0.0.1 表示只本机。
	ListenAddr string `json:"listen_addr"`
	// Domain 是绑定的域名（如 panel.example.com）。优先用于节点分享链接与订阅源。
	Domain string `json:"domain,omitempty"`
	// CertFile 是 SSL 证书文件绝对路径（.crt / .pem）。
	CertFile string `json:"cert_file,omitempty"`
	// KeyFile 是 SSL 私钥文件绝对路径（.key）。
	KeyFile string `json:"key_file,omitempty"`
	// SSLMode 是 SSL 模式："none" | "custom" | "acme" | "caddy"。
	SSLMode string `json:"ssl_mode,omitempty"`
}

// IsTLSEnabled 返回当前配置是否应当以原生 TLS 监听。
func (s WebSettings) IsTLSEnabled() bool {
	if s.SSLMode == "none" || s.SSLMode == "caddy" {
		return false
	}
	return strings.TrimSpace(s.CertFile) != "" && strings.TrimSpace(s.KeyFile) != ""
}

// CertDetails 包含证书解析与诊断详情。
type CertDetails struct {
	SubjectCN string    `json:"subject_cn"`
	Issuer    string    `json:"issuer"`
	SANs      []string  `json:"sans"`
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	DaysLeft  int       `json:"days_left"`
	IsExpired bool      `json:"is_expired"`
	CertPath  string    `json:"cert_path"`
	KeyPath   string    `json:"key_path"`
}

var (
	webSettingsMu   sync.RWMutex
	webSettingsCur  WebSettings
	webSettingsPath string
)

func webSettingsFilePath(dir string) string { return filepath.Join(dir, "settings.json") }

// loadWebSettings 读盘并返回当前配置。
//
// portExplicit 表示用户在命令行显式给了 -web。界面上改过端口之后会落盘，
// 之前这里一律以盘上为准，导致再带 -web 启动会被静默忽略——用户敲了参数却
// 连不上，也没有任何提示。显式指定时以命令行为准并写回，让参数说话算话。
func loadWebSettings(dir string, defaultPort int, portExplicit bool) (WebSettings, error) {
	webSettingsPath = webSettingsFilePath(dir)

	s := WebSettings{Port: defaultPort, ListenAddr: ""}
	blob, err := os.ReadFile(webSettingsPath)
	switch {
	case os.IsNotExist(err):
		webSettingsMu.Lock()
		webSettingsCur = s
		webSettingsMu.Unlock()
		return s, saveWebSettings()
	case err != nil:
		return s, err
	}
	if err := json.Unmarshal(blob, &s); err != nil {
		return s, err
	}
	if s.Port == 0 {
		s.Port = defaultPort
	}
	changed := false
	if portExplicit && s.Port != defaultPort {
		s.Port = defaultPort
		changed = true
	}
	if s.Domain != "" {
		if norm, err := normalizeDomain(s.Domain); err == nil && norm != s.Domain {
			s.Domain = norm
			changed = true
		}
	}
	webSettingsMu.Lock()
	webSettingsCur = s
	webSettingsMu.Unlock()
	if changed {
		return s, saveWebSettings()
	}
	return s, nil
}

func getWebSettings() WebSettings {
	webSettingsMu.RLock()
	defer webSettingsMu.RUnlock()
	return webSettingsCur
}

func saveWebSettings() error {
	webSettingsMu.RLock()
	blob, err := json.MarshalIndent(webSettingsCur, "", "  ")
	webSettingsMu.RUnlock()
	if err != nil {
		return err
	}
	tmp := webSettingsPath + ".tmp"
	if err := os.WriteFile(tmp, blob, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, webSettingsPath)
}

// normalizeListenAddr 把用户填的监听地址规整成合法值：空 / 0.0.0.0 / 127.0.0.1 / 具体 IP。
func normalizeListenAddr(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" || addr == "0.0.0.0" || strings.EqualFold(addr, "all") {
		return "", nil
	}
	if ip := net.ParseIP(addr); ip != nil {
		return addr, nil
	}
	return "", fmt.Errorf("监听地址必须是合法 IP，或留空表示所有网卡")
}

// validatePort 校验端口范围。
func validatePort(p int) error {
	if p < 1 || p > 65535 {
		return fmt.Errorf("端口必须在 1-65535 之间")
	}
	return nil
}

// listenAddrString 拼出 net.Listen 用的地址串。
func (s WebSettings) listenAddrString() string {
	return net.JoinHostPort(s.ListenAddr, strconv.Itoa(s.Port))
}

// normalizeDomain 规范化并校验域名（RFC 1123）。剔除协议前缀、端口、路径，统一转为小写。
func normalizeDomain(d string) (string, error) {
	d = strings.TrimSpace(d)
	if d == "" {
		return "", nil
	}
	if idx := strings.Index(d, "://"); idx != -1 {
		d = d[idx+3:]
	}
	if idx := strings.IndexAny(d, "/?#"); idx != -1 {
		d = d[:idx]
	}
	if h, _, err := net.SplitHostPort(d); err == nil {
		d = h
	}
	d = strings.TrimPrefix(d, "[")
	d = strings.TrimSuffix(d, "]")
	for strings.HasSuffix(d, ".") {
		d = strings.TrimSuffix(d, ".")
	}
	d = strings.ToLower(strings.TrimSpace(d))
	if d == "" {
		return "", nil
	}

	if ip := net.ParseIP(d); ip != nil {
		return ip.String(), nil
	}

	if len(d) > 253 {
		return "", fmt.Errorf("域名长度不能超过 253 个字符")
	}

	labels := strings.Split(d, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return "", fmt.Errorf("域名分段长度必须在 1 到 63 之间: %q", label)
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return "", fmt.Errorf("域名分段不能以连字符 '-' 开头或结尾: %q", label)
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				return "", fmt.Errorf("域名包含非法字符 %q (只允许字母、数字、连字符): %q", string(c), label)
			}
		}
	}
	return d, nil
}

// inspectCertificate 解析并校验 X509 证书与私钥配对，返回证书详情。
func inspectCertificate(certPath, keyPath string) (*CertDetails, error) {
	certPath = strings.TrimSpace(certPath)
	keyPath = strings.TrimSpace(keyPath)
	if certPath == "" {
		return nil, fmt.Errorf("证书文件路径不能为空")
	}
	if keyPath == "" {
		return nil, fmt.Errorf("私钥文件路径不能为空")
	}

	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("读取证书文件失败: %w", err)
	}
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	tlsCert, err := tls.X509KeyPair(certBytes, keyBytes)
	if err != nil {
		return nil, fmt.Errorf("证书与私钥校验失败: %w", err)
	}

	if len(tlsCert.Certificate) == 0 {
		return nil, fmt.Errorf("证书中未包含有效的 X509 数据")
	}

	leaf, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("解析 X509 证书失败: %w", err)
	}

	issuer := leaf.Issuer.CommonName
	if issuer == "" {
		if len(leaf.Issuer.Organization) > 0 {
			issuer = strings.Join(leaf.Issuer.Organization, ", ")
		} else {
			issuer = leaf.Issuer.String()
		}
	}

	sans := append([]string{}, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		sans = append(sans, ip.String())
	}

	now := time.Now()
	rem := time.Until(leaf.NotAfter)
	daysLeft := 0
	if rem > 0 {
		daysLeft = int(math.Ceil(rem.Hours() / 24))
	}
	isExpired := now.After(leaf.NotAfter) || now.Before(leaf.NotBefore)

	return &CertDetails{
		SubjectCN: leaf.Subject.CommonName,
		Issuer:    issuer,
		SANs:      sans,
		NotBefore: leaf.NotBefore,
		NotAfter:  leaf.NotAfter,
		DaysLeft:  daysLeft,
		IsExpired: isExpired,
		CertPath:  certPath,
		KeyPath:   keyPath,
	}, nil
}
