package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestAuth() *Auth {
	return &Auth{
		password: "secret",
		sessions: map[string]time.Time{},
		fails:    map[string]*loginFails{},
	}
}

// 连续失败达到阈值后，该 IP 应被冷却挡下。
func TestLoginThrottleBlocksAfterMaxFails(t *testing.T) {
	a := newTestAuth()
	const ip = "203.0.113.7"

	for i := 0; i < loginMaxFails; i++ {
		if a.blocked(ip) {
			t.Fatalf("第 %d 次失败前不该被封", i)
		}
		a.recordFail(ip)
	}
	if !a.blocked(ip) {
		t.Fatalf("连续 %d 次失败后应进入冷却", loginMaxFails)
	}
}

// 登录成功要清零，之前的失败不该累积到下一轮。
func TestLoginThrottleClearOnSuccess(t *testing.T) {
	a := newTestAuth()
	const ip = "203.0.113.8"

	for i := 0; i < loginMaxFails-1; i++ {
		a.recordFail(ip)
	}
	a.clearFails(ip)
	if a.blocked(ip) {
		t.Fatal("清零后不该被封")
	}
	// 清零后再错一次也不该立刻触发冷却
	a.recordFail(ip)
	if a.blocked(ip) {
		t.Fatal("清零后单次失败不该被封")
	}
}

// 不同 IP 的失败互不牵连。
func TestLoginThrottleIsolatesIPs(t *testing.T) {
	a := newTestAuth()
	for i := 0; i < loginMaxFails; i++ {
		a.recordFail("198.51.100.1")
	}
	if a.blocked("198.51.100.2") {
		t.Fatal("一个 IP 被封不该波及另一个 IP")
	}
}

// 冷却到期后应自动解封。
func TestLoginThrottleUnblocksAfterExpiry(t *testing.T) {
	a := newTestAuth()
	const ip = "203.0.113.9"
	for i := 0; i < loginMaxFails; i++ {
		a.recordFail(ip)
	}
	if !a.blocked(ip) {
		t.Fatal("应先进入冷却")
	}
	// 手动把封禁时间拨到过去，模拟冷却结束
	a.mu.Lock()
	a.fails[ip].blocked = time.Now().Add(-time.Second)
	a.mu.Unlock()
	if a.blocked(ip) {
		t.Fatal("冷却到期后应解封")
	}
}

func TestAuthLogout(t *testing.T) {
	a := newTestAuth()
	tok, err := a.issue()
	if err != nil {
		t.Fatalf("issue error: %v", err)
	}
	if !a.valid(tok) {
		t.Fatal("token should be valid initially")
	}

	// 模拟 POST /api/logout 请求
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok})

	a.handleLogout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /api/logout status = %d, want 200", rec.Code)
	}
	if a.valid(tok) {
		t.Fatal("token should be invalidated after logout")
	}

	// 模拟 GET /logout 重定向
	tok2, _ := a.issue()
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req2.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok2})
	a.handleLogout(rec2, req2)
	if rec2.Code != http.StatusFound {
		t.Fatalf("GET /logout status = %d, want 302", rec2.Code)
	}
	if a.valid(tok2) {
		t.Fatal("token2 should be invalidated after logout")
	}
}

func TestAuthLoginAuthenticatedRedirect(t *testing.T) {
	a := newTestAuth()
	tok, err := a.issue()
	if err != nil {
		t.Fatalf("issue error: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok})

	a.handleLogin(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("authenticated GET /login status = %d, want 302 redirect", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "./" {
		t.Fatalf("authenticated GET /login redirect target = %q, want ./\n", loc)
	}
}

func TestAuthWrapRouting(t *testing.T) {
	a := newTestAuth()
	innerCalled := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	wrapped := a.Wrap(inner)

	// 1. 未登录访问 /api/test 应返回 401
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated /api/test status = %d, want 401", rec1.Code)
	}
	if innerCalled {
		t.Fatal("inner handler should not be called when unauthenticated")
	}

	// 2. 登录后访问 /api/test 应放行
	tok, _ := a.issue()
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok})
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK || !innerCalled {
		t.Fatalf("authenticated /api/test status = %d, innerCalled=%v", rec2.Code, innerCalled)
	}

	// 3. 访问 /api/logout/ (带斜杠) 路由应被拦截并成功登出
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/logout/", nil)
	req3.AddCookie(&http.Cookie{Name: sessionCookie, Value: tok})
	wrapped.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("wrapped POST /api/logout/ status = %d, want 200", rec3.Code)
	}
	if a.valid(tok) {
		t.Fatal("tok should be invalidated via Wrap logout handler")
	}
}
