package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Auth 给管理界面加一层登录。
// 口令存在工作目录下，首次启动自动生成，避免公网上裸奔。
type Auth struct {
	dir      string
	password string
	mu       sync.RWMutex
	sessions map[string]time.Time
	// 按来源 IP 记录登录失败，挡低速凭据喷洒
	fails map[string]*loginFails
}

// loginFails 跟踪单个来源 IP 的连续失败。
type loginFails struct {
	count   int
	last    time.Time
	blocked time.Time
}

const sessionTTL = 12 * time.Hour

// 登录失败节流：同一 IP 连续错 loginMaxFails 次后，锁 loginBlockFor。
// 阈值给得宽松，正常用户偶尔输错不受影响；成功登录会清零。
const (
	loginMaxFails  = 8
	loginBlockFor  = 2 * time.Minute
	loginFailReset = 10 * time.Minute
)

// NewAuth 载入或生成访问口令。返回口令是否为本次新建。
func NewAuth(dir string) (*Auth, bool, error) {
	path := filepath.Join(dir, "password")
	created := false

	blob, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		pw, gerr := randomToken(9)
		if gerr != nil {
			return nil, false, gerr
		}
		if werr := os.WriteFile(path, []byte(pw+"\n"), 0600); werr != nil {
			return nil, false, fmt.Errorf("写口令文件失败: %w", werr)
		}
		blob = []byte(pw)
		created = true
	} else if err != nil {
		return nil, false, err
	}

	return &Auth{
		dir:      dir,
		password: strings.TrimSpace(string(blob)),
		sessions: map[string]time.Time{},
		fails:    map[string]*loginFails{},
	}, created, nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// check 比对口令，用恒定时间比较避免时序泄漏。
func (a *Auth) check(pw string) bool {
	a.mu.RLock()
	cur := a.password
	a.mu.RUnlock()
	want := sha256.Sum256([]byte(cur))
	got := sha256.Sum256([]byte(pw))
	return subtle.ConstantTimeCompare(want[:], got[:]) == 1
}

// SetPassword 改访问口令并落盘。空口令拒绝，避免误改成无密码裸奔。
// 改完不动已有会话：当前登录的浏览器不会被踢，新登录才用新口令。
func (a *Auth) SetPassword(pw string) error {
	pw = strings.TrimSpace(pw)
	if pw == "" {
		return fmt.Errorf("口令不能为空")
	}
	if len(pw) < 4 {
		return fmt.Errorf("口令至少 4 位")
	}
	path := filepath.Join(a.dir, "password")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(pw+"\n"), 0600); err != nil {
		return fmt.Errorf("写口令文件失败: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("保存口令失败: %w", err)
	}
	a.mu.Lock()
	a.password = pw
	a.mu.Unlock()
	return nil
}

// Password 返回当前访问口令。
func (a *Auth) Password() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.password
}

// issue 发一个会话 token。
func (a *Auth) issue() (string, error) {
	tok, err := randomToken(16)
	if err != nil {
		return "", err
	}
	a.mu.Lock()
	a.sessions[tok] = time.Now().Add(sessionTTL)
	// 顺手清掉过期会话
	for k, exp := range a.sessions {
		if time.Now().After(exp) {
			delete(a.sessions, k)
		}
	}
	a.mu.Unlock()
	return tok, nil
}

func (a *Auth) valid(tok string) bool {
	a.mu.RLock()
	exp, ok := a.sessions[tok]
	a.mu.RUnlock()
	return ok && time.Now().Before(exp)
}

const sessionCookie = "fanout_session"

// Wrap 保护一个 handler，未登录时 API 返回 401、页面跳登录。
func (a *Auth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")
		if path == "/login" {
			a.handleLogin(w, r)
			return
		}
		if path == "/logout" || path == "/api/logout" {
			a.handleLogout(w, r)
			return
		}
		// 客户端订阅免 cookie 访问（凭 token 或密码鉴权）
		if path == "/sub" || path == "/api/sub" {
			tok := r.URL.Query().Get("token")
			if tok == "" {
				tok = r.URL.Query().Get("pwd")
			}
			if tok != "" && (a.check(tok) || a.valid(tok)) {
				next.ServeHTTP(w, r)
				return
			}
		}
		if c, err := r.Cookie(sessionCookie); err == nil && a.valid(c.Value) {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "未登录"})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(loginHTML))
	})
}

// blocked 判断某来源 IP 是否处于登录冷却期。
func (a *Auth) blocked(ip string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	f, ok := a.fails[ip]
	return ok && time.Now().Before(f.blocked)
}

// recordFail 记一次失败，达到阈值就进入冷却。
func (a *Auth) recordFail(ip string) {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	f, ok := a.fails[ip]
	// 距上次失败太久就重新计数，避免长期累积误伤
	if !ok || (f.blocked.IsZero() && now.Sub(f.last) > loginFailReset) {
		f = &loginFails{}
		a.fails[ip] = f
	}
	f.count++
	f.last = now
	if f.count >= loginMaxFails {
		f.blocked = now.Add(loginBlockFor)
		f.count = 0
	}
	// 顺手清掉早已过期的记录，别让 map 无限增长
	for k, v := range a.fails {
		if now.Sub(v.last) > loginFailReset && now.After(v.blocked) {
			delete(a.fails, k)
		}
	}
}

// clearFails 登录成功后清掉该 IP 的失败记录。
func (a *Auth) clearFails(ip string) {
	a.mu.Lock()
	delete(a.fails, ip)
	a.mu.Unlock()
}

// clientIP 从 RemoteAddr 取来源 IP。服务直接监听公网端口、不在反代后，
// 所以不采信 X-Forwarded-For 之类可伪造的头。
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func redirectRelative(w http.ResponseWriter, target string, code int) {
	w.Header().Set("Location", target)
	w.WriteHeader(code)
}

func (a *Auth) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		if c, err := r.Cookie(sessionCookie); err == nil && a.valid(c.Value) {
			redirectRelative(w, "./", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(loginHTML))
		return
	}
	ip := clientIP(r)
	if a.blocked(ip) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "登录失败次数过多，请稍后再试"})
		return
	}
	if !a.check(r.FormValue("password")) {
		a.recordFail(ip)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "口令不对"})
		return
	}
	a.clearFails(ip)
	tok, err := a.issue()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
	writeJSON(w, http.StatusOK, map[string]string{"ok": "已登录"})
}

func (a *Auth) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
	path := strings.TrimSuffix(r.URL.Path, "/")
	if r.Method == http.MethodPost || path == "/api/logout" || strings.Contains(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, http.StatusOK, map[string]string{"ok": "已退出登录"})
		return
	}
	redirectRelative(w, "login", http.StatusFound)
}

const loginHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>fanout PRO - 登录控制台</title>
<script>
(function(){
  try{
    var t = localStorage.getItem('fanout_theme');
    if(!t){
      t = window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
    }
    document.documentElement.className = t === 'light' ? 'light' : 'dark';
  }catch(e){}
})();
</script>
<style>
:root, html.dark {
  --bg-canvas: #0b0f19;
  --bg-card: rgba(17, 24, 39, 0.9);
  --bg-surface: #172033;
  --bg-input: #0e1524;
  --border-card: rgba(255, 255, 255, 0.08);
  --border-focus: #0284c7;
  --text-main: #f1f5f9;
  --text-muted: #94a3b8;
  --accent-blue: #0284c7;
  --status-danger-bg: rgba(239, 68, 68, 0.12);
  --status-danger-border: rgba(239, 68, 68, 0.32);
  --status-danger-text: #fca5a5;
  --card-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255, 255, 255, 0.04);
  --radius-lg: 14px;
  --radius-md: 8px;
  --radius-sm: 6px;
}

html.light {
  --bg-canvas: #f8fafc;
  --bg-card: rgba(255, 255, 255, 0.95);
  --bg-surface: #f1f5f9;
  --bg-input: #ffffff;
  --border-card: rgba(15, 23, 42, 0.1);
  --border-focus: #0284c7;
  --text-main: #0f172a;
  --text-muted: #64748b;
  --accent-blue: #0284c7;
  --status-danger-bg: rgba(239, 68, 68, 0.08);
  --status-danger-border: rgba(239, 68, 68, 0.25);
  --status-danger-text: #dc2626;
  --card-shadow: 0 20px 40px -15px rgba(0, 0, 0, 0.08), 0 0 0 1px rgba(0, 0, 0, 0.05);
  --radius-lg: 14px;
  --radius-md: 8px;
  --radius-sm: 6px;
}

* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: var(--bg-canvas);
  background-image:
    radial-gradient(circle at 50% 20%, rgba(2, 132, 199, 0.16) 0%, transparent 60%),
    radial-gradient(circle at 80% 80%, rgba(79, 70, 229, 0.08) 0%, transparent 50%);
  color: var(--text-main);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  padding: 20px;
  overflow-x: hidden;
  transition: background-color 0.2s ease, color 0.2s ease;
}
html.light body {
  background-color: #f8fafc;
  background-image:
    radial-gradient(circle at 50% 20%, rgba(2, 132, 199, 0.08) 0%, transparent 60%),
    radial-gradient(circle at 80% 80%, rgba(14, 165, 233, 0.05) 0%, transparent 50%);
}

.login-card {
  background: var(--bg-card);
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  border: 1px solid var(--border-card);
  box-shadow: var(--card-shadow);
  border-radius: var(--radius-lg);
  padding: 38px 32px;
  width: 100%;
  max-width: 380px;
  animation: cardFadeUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}
@keyframes cardFadeUp {
  from { opacity: 0; transform: translateY(12px) scale(0.98); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20%, 60% { transform: translateX(-6px); }
  40%, 80% { transform: translateX(6px); }
}
.shake { animation: shake 0.38s ease-in-out; }

.brand-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 8px;
}
.logo-mark {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: linear-gradient(135deg, #0284c7, #4f46e5);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: 0 0 14px rgba(2, 132, 199, 0.4);
}
.logo-mark svg {
  width: 18px;
  height: 18px;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
  fill: none;
}
.brand-title {
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: var(--text-main);
  display: flex;
  align-items: center;
  gap: 8px;
}
.pro-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 7px;
  border-radius: 9999px;
  background: rgba(2, 132, 199, 0.15);
  border: 1px solid rgba(2, 132, 199, 0.35);
  color: #38bdf8;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}
.brand-subtitle {
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 26px;
}

.field {
  margin-bottom: 18px;
}
.field label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 8px;
}
.input-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.input-icon {
  position: absolute;
  left: 12px;
  color: var(--text-muted);
  display: flex;
  align-items: center;
  pointer-events: none;
}
.input-icon svg, .eye-toggle svg {
  width: 16px;
  height: 16px;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
  fill: none;
}
.input-wrap input {
  width: 100%;
  background: var(--bg-input);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-md);
  padding: 11px 40px 11px 38px;
  color: var(--text-main);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  outline: none;
  transition: all 0.15s ease;
}
.input-wrap input:focus {
  border-color: var(--border-focus);
  box-shadow: 0 0 0 3px rgba(2, 132, 199, 0.22);
  background: var(--bg-input);
}
/* 浏览器自动填充与密码管理器的背景强制匹配 */
.input-wrap input:-webkit-autofill,
.input-wrap input:-webkit-autofill:hover, 
.input-wrap input:-webkit-autofill:focus,
.input-wrap input:-webkit-autofill:active {
  -webkit-text-fill-color: var(--text-main) !important;
  -webkit-box-shadow: 0 0 0px 1000px var(--bg-input) inset !important;
  box-shadow: 0 0 0px 1000px var(--bg-input) inset !important;
  transition: background-color 5000s ease-in-out 0s;
}
.eye-toggle {
  position: absolute;
  right: 10px;
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  border-radius: 4px;
  transition: color 0.15s;
}
.eye-toggle:hover {
  color: var(--text-main);
}

.btn-submit {
  width: 100%;
  margin-top: 6px;
  background: linear-gradient(135deg, #0284c7, #0369a1);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  border-radius: var(--radius-md);
  padding: 11px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  box-shadow: 0 4px 14px rgba(2, 132, 199, 0.35);
  transition: all 0.15s ease;
}
.btn-submit:hover:not(:disabled) {
  background: linear-gradient(135deg, #0369a1, #0284c7);
  box-shadow: 0 6px 18px rgba(2, 132, 199, 0.45);
  transform: translateY(-1px);
}
.btn-submit:active:not(:disabled) {
  transform: translateY(0);
}
.btn-submit:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  transform: none;
}
.spin {
  width: 16px;
  height: 16px;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  fill: none;
  animation: spin 0.8s linear infinite;
}
@keyframes spin { 100% { transform: rotate(360deg); } }

.err-box {
  margin-top: 14px;
  padding: 9px 12px;
  border-radius: var(--radius-sm);
  background: var(--status-danger-bg);
  border: 1px solid var(--status-danger-border);
  color: var(--status-danger-text);
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.err-box svg {
  width: 15px;
  height: 15px;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
  fill: none;
  flex-shrink: 0;
}

.login-footer {
  margin-top: 24px;
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-muted);
  font-size: 11px;
}
.login-footer svg {
  width: 13px;
  height: 13px;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
  fill: none;
}

.theme-toggle-corner {
  position: fixed;
  top: 20px;
  right: 20px;
  background: var(--bg-card);
  border: 1px solid var(--border-card);
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  width: 38px;
  height: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: var(--card-shadow);
  transition: all 0.15s ease;
  z-index: 100;
}
.theme-toggle-corner:hover {
  color: var(--text-main);
  border-color: var(--border-focus);
  transform: translateY(-1px);
}
.theme-toggle-corner svg {
  width: 18px;
  height: 18px;
}
</style>
</head>
<body>
<button type="button" class="theme-toggle-corner" id="themeToggleBtn" onclick="toggleTheme()" title="切换浅色/深色模式">
  <span id="themeToggleIcon"></span>
</button>

<div class="login-card" id="loginCard">
  <div class="brand-header">
    <div class="logo-mark">
      <svg viewBox="0 0 24 24">
        <circle cx="6" cy="12" r="3"/>
        <circle cx="18" cy="6" r="3"/>
        <circle cx="18" cy="18" r="3"/>
        <path d="M9 12h3a3 3 0 0 0 3-3V6m-3 6a3 3 0 0 1 3 3v3"/>
      </svg>
    </div>
    <div class="brand-title">
      fanout
      <span class="pro-badge">PRO</span>
    </div>
  </div>
  <p class="brand-subtitle">个人专属出口网关与智能分流控制台</p>

  <form id="f">
    <div class="field">
      <label for="pw">访问口令</label>
      <div class="input-wrap">
        <span class="input-icon">
          <svg viewBox="0 0 24 24"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
        </span>
        <input type="password" id="pw" autofocus autocomplete="current-password" placeholder="请输入管理口令">
        <button type="button" class="eye-toggle" id="eyeToggle" title="显示/隐藏口令" tabindex="-1">
          <svg id="eyeIcon" viewBox="0 0 24 24"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/></svg>
        </button>
      </div>
    </div>

    <button type="submit" id="submitBtn" class="btn-submit">
      <span id="btnText">进入控制台</span>
      <svg id="btnSpinner" class="spin" style="display:none" viewBox="0 0 24 24"><path d="M21 12a9 9 0 1 1-6.2-8.5"/></svg>
    </button>

    <div class="err-box" id="errBox" style="display:none">
      <svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
      <span id="errText"></span>
    </div>
  </form>
</div>

<div class="login-footer">
  <svg viewBox="0 0 24 24"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
  <span>Fanout PRO · 端到端安全加密控制台</span>
</div>

<script>
const pwInput = document.getElementById('pw');
const eyeBtn = document.getElementById('eyeToggle');
const eyeIcon = document.getElementById('eyeIcon');
const card = document.getElementById('loginCard');
const errBox = document.getElementById('errBox');
const errText = document.getElementById('errText');
const submitBtn = document.getElementById('submitBtn');
const btnText = document.getElementById('btnText');
const btnSpinner = document.getElementById('btnSpinner');

const EYE_OPEN = '<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"/><circle cx="12" cy="12" r="3"/>';
const EYE_CLOSE = '<path d="M9.88 9.88a3 3 0 1 0 4.24 4.24"/><path d="M10.73 5.08A10.43 10.43 0 0 1 12 5c7 0 10 7 10 7a13.16 13.16 0 0 1-1.67 2.68"/><path d="M6.61 6.61A13.526 13.526 0 0 0 2 12s3 7 10 7a9.74 9.74 0 0 0 5.39-1.61"/><line x1="2" x2="22" y1="2" y2="22"/>';

eyeBtn.onclick = () => {
  const isPw = pwInput.type === 'password';
  pwInput.type = isPw ? 'text' : 'password';
  eyeIcon.innerHTML = isPw ? EYE_CLOSE : EYE_OPEN;
};

function showError(msg){
  errText.textContent = msg;
  errBox.style.display = 'flex';
  card.classList.remove('shake');
  void card.offsetWidth; // 触发 reflow 重置动画
  card.classList.add('shake');
}

document.getElementById('f').onsubmit = async e => {
  e.preventDefault();
  const pw = pwInput.value.trim();
  if(!pw){
    showError('请输入访问管理口令');
    pwInput.focus();
    return;
  }
  errBox.style.display = 'none';
  submitBtn.disabled = true;
  btnText.textContent = '验证中…';
  btnSpinner.style.display = 'inline-block';

  try{
    const body = new URLSearchParams({password: pw});
    const r = await fetch('login', {method:'POST', body});
    if(r.ok){
      btnText.textContent = '已验证，正在进入…';
      location.replace('./');
      return;
    }
    const d = await r.json().catch(()=>({}));
    showError(d.error || '访问口令不正确');
    pwInput.select();
  }catch(err){
    showError('网络连接异常: ' + err.message);
  }finally{
    submitBtn.disabled = false;
    btnText.textContent = '进入控制台';
    btnSpinner.style.display = 'none';
  }
};

function updateThemeIcon(isLight){
  const iconSpan = document.getElementById('themeToggleIcon');
  if(!iconSpan) return;
  if(isLight){
    iconSpan.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/></svg>';
  } else {
    iconSpan.innerHTML = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg>';
  }
}
function toggleTheme(){
  const isLight = document.documentElement.classList.contains('light');
  const target = isLight ? 'dark' : 'light';
  document.documentElement.className = target;
  try{ localStorage.setItem('fanout_theme', target); }catch(e){}
  updateThemeIcon(target === 'light');
}
updateThemeIcon(document.documentElement.classList.contains('light'));
if(window.matchMedia){
  window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', e => {
    if(!localStorage.getItem('fanout_theme')){
      const target = e.matches ? 'light' : 'dark';
      document.documentElement.className = target;
      updateThemeIcon(target === 'light');
    }
  });
}
</script>
</body>
</html>`

