package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEvaluateQuality(t *testing.T) {
	// 1. 机房/IDC
	dc := EvaluateQuality("1.1.1.1", "Cloudflare, Inc.", "Cloudflare", "AS13335 Cloudflare", true, false, false)
	if dc.Type != "datacenter" || dc.Score != 60 || dc.TypeLabel != "🏢 机房/IDC" {
		t.Errorf("datacenter evaluation mismatch: %+v", dc)
	}

	// 2. 普通家宽
	res := EvaluateQuality("2.2.2.2", "Local Telecom", "Local Tel", "AS9999 Local", false, false, false)
	if res.Type != "residential" || res.Score != 95 || res.TypeLabel != "🏠 家庭宽带" {
		t.Errorf("residential evaluation mismatch: %+v", res)
	}

	// 3. 知名优质家宽 (如 NTT OCN, 加5分，最高100)
	ntt := EvaluateQuality("3.3.3.3", "NTT Communications", "OCN", "AS4713 NTT Communications", false, false, false)
	if ntt.Type != "residential" || ntt.Score != 100 {
		t.Errorf("recognized residential ISP bonus mismatch: %+v", ntt)
	}

	// 4. 移动蜂窝
	mob := EvaluateQuality("4.4.4.4", "Generic Mobile", "Mobile Network", "AS8888 Mob", false, false, true)
	if mob.Type != "mobile" || mob.Score != 88 || mob.TypeLabel != "📱 移动蜂窝" {
		t.Errorf("mobile evaluation mismatch: %+v", mob)
	}

	// 5. 代理扣分
	proxy := EvaluateQuality("5.5.5.5", "Some ISP", "Some Org", "AS1234 Some", false, true, false)
	if proxy.Score != 70 { // 95 - 25 = 70
		t.Errorf("proxy penalty mismatch: %+v", proxy)
	}

	// 6. 分数边界限制 (0~100)
	proxyDC := EvaluateQuality("6.6.6.6", "Bad Hosting", "Bad", "AS555", true, true, false)
	if proxyDC.Score != 35 { // 60 - 25 = 35
		t.Errorf("datacenter proxy score mismatch: %+v", proxyDC)
	}
}

func TestQualityCacheSaveLoad(t *testing.T) {
	dir := t.TempDir()
	qc := NewQualityCache(dir)

	info := QualityInfo{
		IP:        "192.0.2.1",
		Type:      "residential",
		TypeLabel: "🏠 家庭宽带",
		Score:     98,
		ISP:       "Test ISP",
		Org:       "Test Org",
		ASN:       "AS12345",
		Hosting:   false,
		Proxy:     false,
		Mobile:    false,
		UpdatedAt: time.Now(),
	}

	qc.Set(info)
	if err := qc.Save(); err != nil {
		t.Fatalf("Save cache failed: %v", err)
	}

	// 新建一个 cache 对象读取保存的文件
	qc2 := NewQualityCache(dir)
	got, ok := qc2.Get("192.0.2.1")
	if !ok {
		t.Fatalf("Failed to retrieve cached IP")
	}
	if got.Score != 98 || got.ISP != "Test ISP" {
		t.Errorf("Loaded info mismatch: %+v", got)
	}

	// 模拟过期
	qc2.mu.Lock()
	got.UpdatedAt = time.Now().Add(-25 * time.Hour)
	qc2.items["192.0.2.1"] = got
	qc2.mu.Unlock()

	if _, ok := qc2.Get("192.0.2.1"); ok {
		t.Errorf("Expected expired cache entry to be missed")
	}
}

func TestBatchEvaluateMock(t *testing.T) {
	// 创建 mock ip-api batch 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqItems []ipAPIBatchItem
		if err := json.NewDecoder(r.Body).Decode(&reqItems); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := make([]ipAPIBatchRespItem, len(reqItems))
		for i, item := range reqItems {
			if item.Query == "198.51.100.1" {
				resp[i] = ipAPIBatchRespItem{
					Status:      "success",
					CountryCode: "JP",
					City:        "Tokyo",
					ISP:         "NTT Communications",
					Org:         "OCN",
					AS:          "AS4713 NTT",
					Mobile:      false,
					Proxy:       false,
					Hosting:     false,
					Query:       item.Query,
				}
			} else {
				resp[i] = ipAPIBatchRespItem{
					Status:      "success",
					CountryCode: "US",
					City:        "Los Angeles",
					ISP:         "DataCamp Limited",
					Org:         "DataCamp",
					AS:          "AS60068 Datacamp",
					Mobile:      false,
					Proxy:       false,
					Hosting:     true,
					Query:       item.Query,
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	dir := t.TempDir()
	qc := NewQualityCache(dir)
	qc.apiURL = mockServer.URL
	qc.client = mockServer.Client()

	ips := []string{"198.51.100.1", "203.0.113.2"}
	results := qc.BatchEvaluate(ips)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	q1, ok1 := results["198.51.100.1"]
	if !ok1 || q1.Type != "residential" || q1.Score != 100 {
		t.Errorf("Result 1 mismatch: %+v", q1)
	}

	q2, ok2 := results["203.0.113.2"]
	if !ok2 || q2.Type != "datacenter" || q2.Score != 60 {
		t.Errorf("Result 2 mismatch: %+v", q2)
	}

	// 确认存入缓存
	cached1, ok := qc.Get("198.51.100.1")
	if !ok || cached1.Score != 100 {
		t.Errorf("Cache retrieval mismatch: %+v", cached1)
	}
}

func TestFallbackQualityAndErrorHandling(t *testing.T) {
	// 1. 验证 fallbackQuality 是诚实的未知评级（不可误标为家宽或虚高分）
	fb := fallbackQuality("1.2.3.4")
	if fb.Type != "unknown" || fb.Score != 50 || fb.TypeLabel != "❓ 待检测" {
		t.Errorf("fallbackQuality should be unknown/50, got: %+v", fb)
	}

	dir := t.TempDir()
	qc := NewQualityCache(dir)

	// 2. qc.Get 针对 unknown 应当视为 miss（不锁死24h，以便恢复后重试）
	qc.Set(fb)
	if _, ok := qc.Get("1.2.3.4"); ok {
		t.Errorf("qc.Get for unknown type should return miss/false")
	}

	// 3. 模拟外部接口 500 报错，验证返回 fallback 且不污染持久化文件
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Rate limited or server error", http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	qc.apiURL = errorServer.URL
	qc.client = errorServer.Client()

	results := qc.BatchEvaluate([]string{"5.6.7.8"})
	q, ok := results["5.6.7.8"]
	if !ok || q.Type != "unknown" || q.Score != 50 {
		t.Errorf("Failed query should return fallback quality, got: %+v", q)
	}

	// 确认不被持久化为缓存
	if _, ok := qc.Get("5.6.7.8"); ok {
		t.Errorf("Transient error should not be cached in qc")
	}
}
