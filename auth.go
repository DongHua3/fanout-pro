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
<title>fanout - 登录</title>
<style>
:root{
  --bg:#0e1116; --card:#161b22; --line:#262c36; --text:#dde3ec;
  --dim:#8b95a5; --accent:#4a9eda; --bad:#c25450;
}
*{box-sizing:border-box}
body{margin:0;height:100vh;display:flex;flex-direction:column;gap:18px;
  align-items:center;justify-content:center;
  background:radial-gradient(circle at 50% 30%, #151a23 0%, var(--bg) 100%);
  color:var(--text);font:13px/1.5 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}
form{background:var(--card);border:1px solid var(--line);border-radius:8px;
  padding:26px 24px;width:320px;box-shadow:0 12px 36px rgba(0,0,0,.45)}
.brand{display:flex;align-items:center;gap:8px;margin-bottom:6px}
h1{font-size:15px;font-weight:700;margin:0;letter-spacing:.3px;color:#fff}
.tag{font-size:10px;font-weight:700;padding:1px 5px;background:rgba(74,158,218,.15);
  color:var(--accent);border:1px solid rgba(74,158,218,.35);border-radius:4px}
.subtitle{font-size:11px;color:var(--dim);margin-bottom:18px}
label{display:block;color:var(--dim);font-size:11px;margin-bottom:6px}
input{width:100%;box-sizing:border-box;background:#0a0d11;border:1px solid var(--line);
  color:var(--text);border-radius:4px;padding:8px 10px;font:inherit;transition:border-color .15s}
input:focus{outline:none;border-color:var(--accent)}
button{width:100%;margin-top:16px;background:var(--accent);border:0;color:#080b0f;
  font:inherit;font-weight:700;border-radius:4px;padding:9px;cursor:pointer;transition:opacity .15s}
button:hover{opacity:.92}
button:active{transform:translateY(1px)}
.err{color:var(--bad);font-size:11px;margin-top:10px;min-height:16px;text-align:center}
.links{display:flex;gap:16px}
.links a{color:var(--dim);text-decoration:none;font-size:11px;transition:color .15s}
.links a:hover{color:var(--accent)}
</style>
</head>
<body>
<form id="f">
  <div class="brand">
    <h1>fanout</h1>
    <span class="tag">PRO</span>
  </div>
  <div class="subtitle">个人专属出口网关与智能分流控制台</div>
  <label for="pw">访问口令</label>
  <input type="password" id="pw" autofocus autocomplete="current-password" placeholder="请输入管理口令">
  <button type="submit">进入控制台</button>
  <div class="err" id="err"></div>
</form>
<div class="links">
  <a href="https://github.com/DongHua3/fanout-pro" target="_blank" rel="noopener">fanout-pro</a>
</div>
<script>
document.getElementById('f').onsubmit = async e => {
  e.preventDefault();
  const pw = document.getElementById('pw').value;
  if(!pw){ document.getElementById('err').textContent = '请输入口令'; return; }
  const btn = document.querySelector('button');
  btn.disabled = true;
  btn.textContent = '验证中…';
  try{
    const body = new URLSearchParams({password: pw});
    const r = await fetch('login', {method:'POST', body});
    if(r.ok){ location.replace('./'); return; }
    const d = await r.json().catch(()=>({}));
    document.getElementById('err').textContent = d.error || '登录失败';
  }catch(e){
    document.getElementById('err').textContent = '请求异常: ' + e.message;
  }finally{
    btn.disabled = false;
    btn.textContent = '进入控制台';
  }
};
</script>
</body>
</html>`
