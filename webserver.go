package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// webServer 管理 HTTP / HTTPS 监听，支持在运行时切换端口/监听地址/TLS 证书而不重启进程。
// 切换端口或协议会新起 net.Listener，旧的优雅关闭；证书更新则使用原子指针 0ms 无感生效。
type webServer struct {
	handler     http.Handler
	mu          sync.Mutex
	ln          net.Listener
	srv         *http.Server
	addr        string
	isTLS       bool
	certPath    string
	keyPath     string
	currentCert atomic.Pointer[tls.Certificate] // 动态原子指针，实现 0 毫秒无感证书刷新
}

func newWebServer(h http.Handler) *webServer {
	return &webServer{handler: h}
}

// IsTLS 返回当前 Web 服务是否正在以 HTTPS (TLS) 协议提供服务。
func (s *webServer) IsTLS() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isTLS
}

// serve 用当前 WebSettings 起第一个监听并阻塞。返回时说明监听彻底退出。
func (s *webServer) serve() error {
	cfg := getWebSettings()
	cfgToRun := cfg
	if cfg.IsTLSEnabled() {
		// 校验启动时证书是否存在且可用，若损坏或路径失效则安全降级为 HTTP，防止失联
		if _, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile); err != nil {
			log.Printf("⚠️ TLS 证书加载失败 (%v)，安全降级为 HTTP 明文模式启动，防止失联！磁盘配置已保留，可修复证书后重启或在设置中重新测试生效。", err)
			cfgToRun.SSLMode = "none"
		}
	}
	if err := s.reload(cfgToRun); err != nil {
		return err
	}
	// 主 goroutine 就地阻塞，等监听被 reload 或退出替换。
	// 这里靠一个永不返回的 select 挂住：真正的 Serve 在 reload 里各自的 goroutine 跑。
	select {}
}

// reload 切换到新的监听地址或协议：
// 若为同端口协议切换（HTTP <-> HTTPS），先关旧 listener 再重绑；
// 若为不同地址，先探测能否绑上，绑得上再关旧的、启新的。
func (s *webServer) reload(cfg WebSettings) error {
	addr := cfg.listenAddrString()
	isTLS := cfg.IsTLSEnabled()

	var newCert *tls.Certificate
	if isTLS {
		c, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return fmt.Errorf("加载证书失败：%w", err)
		}
		newCert = &c
	}

	s.mu.Lock()
	sameAddr := (s.ln != nil && s.addr == addr)
	oldSrv := s.srv
	oldLn := s.ln
	oldWasTLS := s.isTLS
	s.mu.Unlock()

	var ln net.Listener
	var err error

	if sameAddr {
		// 同端口切换（如 HTTP <-> HTTPS 协议切换）：
		// 必须先关闭旧 listener 释放端口
		if oldLn != nil {
			_ = oldLn.Close()
		}
		// 重试绑定，给操作系统回收 socket 预留时间
		for i := 0; i < 20; i++ {
			ln, err = net.Listen("tcp", addr)
			if err == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if err != nil {
			// 尝试恢复旧监听
			if rln, rerr := net.Listen("tcp", addr); rerr == nil && oldSrv != nil {
				s.mu.Lock()
				s.ln = rln
				s.mu.Unlock()
				go func() {
					if oldWasTLS {
						_ = oldSrv.ServeTLS(rln, "", "")
					} else {
						_ = oldSrv.Serve(rln)
					}
				}()
			}
			return fmt.Errorf("同端口重新监听 %s 失败：%w", addr, err)
		}
	} else {
		// 不同端口/地址：先探测新端口能否绑上，确保失败时不影响旧服务
		ln, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("无法监听 %s：%w", addr, err)
		}
	}

	var srv *http.Server
	if isTLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				if c := s.currentCert.Load(); c != nil {
					return c, nil
				}
				return nil, fmt.Errorf("未配置有效 TLS 证书")
			},
		}
		srv = &http.Server{
			Handler:   s.handler,
			TLSConfig: tlsConfig,
		}
		s.currentCert.Store(newCert)
	} else {
		srv = &http.Server{Handler: s.handler}
		s.currentCert.Store(nil)
	}

	s.mu.Lock()
	s.srv = srv
	s.ln = ln
	s.addr = addr
	s.isTLS = isTLS
	s.certPath = cfg.CertFile
	s.keyPath = cfg.KeyFile
	s.mu.Unlock()

	go func() {
		var serveErr error
		if isTLS {
			serveErr = srv.ServeTLS(ln, "", "")
		} else {
			serveErr = srv.Serve(ln)
		}
		if serveErr != nil && serveErr != http.ErrServerClosed {
			log.Printf("Web 监听 %s 退出: %v", addr, serveErr)
		}
	}()

	// 关掉旧监听。给正在处理的请求一点收尾时间，
	// 尤其是触发这次 reload 的那个请求本身要先把响应写完。
	if oldSrv != nil {
		go func() {
			if !sameAddr {
				time.Sleep(1 * time.Second)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = oldSrv.Shutdown(ctx)
			if !sameAddr && oldLn != nil {
				_ = oldLn.Close()
			}
		}()
	}

	proto := "HTTP"
	if isTLS {
		proto = "HTTPS"
	}
	log.Printf("管理界面监听已切换到 %s (%s)", addr, proto)
	return nil
}

// applyWebSettings 校验、落盘并应用配置。
// 若仅修改域名或证书热重载（地址与协议未变），则零中断热更新证书指针；
// 若地址改变或 HTTP <-> HTTPS 协议切换，则平滑切换监听器。
func (s *webServer) applyWebSettings(next WebSettings) error {
	if err := validatePort(next.Port); err != nil {
		return err
	}
	norm, err := normalizeListenAddr(next.ListenAddr)
	if err != nil {
		return err
	}
	next.ListenAddr = norm

	normDom, err := normalizeDomain(next.Domain)
	if err != nil {
		return err
	}
	next.Domain = normDom

	nextTLS := next.IsTLSEnabled()
	var nextCert *tls.Certificate
	if nextTLS {
		c, err := tls.LoadX509KeyPair(next.CertFile, next.KeyFile)
		if err != nil {
			return fmt.Errorf("证书或私钥无效：%w", err)
		}
		nextCert = &c
	}

	nextAddr := next.listenAddrString()

	s.mu.Lock()
	srvAddr := s.addr
	srvTLS := s.isTLS
	s.mu.Unlock()

	// 判定是否需要重新绑定网络监听：
	// 如果监听器当前运行的地址和协议与新配置一致，则无需重绑监听器
	addrChanged := (srvAddr == "" || srvAddr != nextAddr)
	protoChanged := (srvTLS != nextTLS)

	if !addrChanged && !protoChanged {
		// 情况 A：仅修改域名、SSLMode 或者更新已有证书文件内容
		// 直接热更新证书指针，无需重启或关闭 TCP Listener，零网络中断！
		if nextTLS && nextCert != nil {
			s.currentCert.Store(nextCert)
			s.mu.Lock()
			s.certPath = next.CertFile
			s.keyPath = next.KeyFile
			s.mu.Unlock()
			log.Printf("TLS 证书已动态热重载 (0ms 无感更新)")
		}

		webSettingsMu.Lock()
		webSettingsCur = next
		webSettingsMu.Unlock()
		if err := saveWebSettings(); err != nil {
			log.Printf("保存 Web 设置失败: %v", err)
			return err
		}
		return nil
	}

	// 情况 B：端口改变或 HTTP <-> HTTPS 协议切换
	if err := s.reload(next); err != nil {
		return err
	}

	webSettingsMu.Lock()
	webSettingsCur = next
	webSettingsMu.Unlock()
	if err := saveWebSettings(); err != nil {
		log.Printf("保存 Web 设置失败: %v", err)
		return err
	}
	return nil
}
