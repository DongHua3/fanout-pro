package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAPINodesSortingAndFiltering(t *testing.T) {
	mgr := &Manager{
		workDir: t.TempDir(),
		nodes: []Node{
			{HostName: "jp-res", IP: "1.1.1.1", Country: "Japan", CountryCode: "JP", SpeedMbps: 50, Ping: 30},
			{HostName: "us-dc", IP: "2.2.2.2", Country: "United States", CountryCode: "US", SpeedMbps: 200, Ping: 150},
			{HostName: "kr-fast", IP: "3.3.3.3", Country: "Korea", CountryCode: "KR", SpeedMbps: 500, Ping: 40},
		},
	}

	qc := GetQualityCache(mgr.workDir)
	qc.Set(QualityInfo{
		IP:        "1.1.1.1",
		Type:      "residential",
		TypeLabel: "🏠 家庭宽带",
		Score:     95,
		ISP:       "NTT Communications",
	})
	qc.Set(QualityInfo{
		IP:        "2.2.2.2",
		Type:      "datacenter",
		TypeLabel: "🏢 机房/IDC",
		Score:     60,
		ISP:       "DigitalOcean LLC",
	})
	qc.Set(QualityInfo{
		IP:        "3.3.3.3",
		Type:      "residential",
		TypeLabel: "🏠 家庭宽带",
		Score:     98,
		ISP:       "KT Corp",
	})

	handler := apiNodes(mgr)

	// 1. 默认按照 Quality 降序排序: kr-fast (98) -> jp-res (95) -> us-dc (60)
	req := httptest.NewRequest(http.MethodGet, "/api/nodes?sort=quality", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("apiNodes returned %d", w.Code)
	}
	var resp struct {
		Nodes []NodeWithQuality `json:"nodes"`
		Total int               `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if len(resp.Nodes) != 3 {
		t.Fatalf("Expected 3 nodes, got %d", len(resp.Nodes))
	}
	if resp.Nodes[0].HostName != "kr-fast" || resp.Nodes[1].HostName != "jp-res" || resp.Nodes[2].HostName != "us-dc" {
		t.Errorf("Unexpected quality sort order: got %v, %v, %v",
			resp.Nodes[0].HostName, resp.Nodes[1].HostName, resp.Nodes[2].HostName)
	}

	// 2. 按照 Speed 降序排序: kr-fast (500) -> us-dc (200) -> jp-res (50)
	req = httptest.NewRequest(http.MethodGet, "/api/nodes?sort=speed", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	var respSpeed struct {
		Nodes []NodeWithQuality `json:"nodes"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respSpeed)
	if respSpeed.Nodes[0].HostName != "kr-fast" || respSpeed.Nodes[1].HostName != "us-dc" || respSpeed.Nodes[2].HostName != "jp-res" {
		t.Errorf("Unexpected speed sort order: got %v, %v, %v",
			respSpeed.Nodes[0].HostName, respSpeed.Nodes[1].HostName, respSpeed.Nodes[2].HostName)
	}

	// 3. 按照 Region 过滤: JP -> 只返回 jp-res
	req = httptest.NewRequest(http.MethodGet, "/api/nodes?region=JP", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	var respRegion struct {
		Nodes []NodeWithQuality `json:"nodes"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respRegion)
	if len(respRegion.Nodes) != 1 || respRegion.Nodes[0].HostName != "jp-res" {
		t.Errorf("Region filter failed, got %v", respRegion.Nodes)
	}

	// 4. 按照 Type 过滤: residential -> 返回 2 个住宅节点 (kr-fast, jp-res)
	req = httptest.NewRequest(http.MethodGet, "/api/nodes?type=residential", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	var respType struct {
		Nodes []NodeWithQuality `json:"nodes"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respType)
	if len(respType.Nodes) != 2 {
		t.Errorf("Type filter failed, expected 2 nodes, got %d", len(respType.Nodes))
	}

	// 5. 关键词搜索
	req = httptest.NewRequest(http.MethodGet, "/api/nodes?kw=digitalocean", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	var respKW struct {
		Nodes []NodeWithQuality `json:"nodes"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &respKW)
	if len(respKW.Nodes) != 1 || respKW.Nodes[0].HostName != "us-dc" {
		t.Errorf("Keyword search failed, got %v", respKW.Nodes)
	}
}

func TestAPINodesStartValidation(t *testing.T) {
	mgr := &Manager{
		workDir: t.TempDir(),
		nodes: []Node{
			{HostName: "test-node-01", IP: "4.4.4.4", Country: "Japan", CountryCode: "JP"},
		},
	}

	handler := apiNodesStart(mgr)

	// 1. 缺少 host 参数
	req := httptest.NewRequest(http.MethodPost, "/api/nodes/start", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing host, got %d", w.Code)
	}

	// 2. 节点不存在
	req = httptest.NewRequest(http.MethodPost, "/api/nodes/start?host=non-existent", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("Expected 409 for non-existent node, got %d", w.Code)
	}
}

func TestAPIXUIBindDirectUnbind(t *testing.T) {
	testDir := t.TempDir()
	mgr := &Manager{
		workDir: testDir,
	}

	handler := apiXUIBind(mgr)

	// 缺少 tag/id/port 参数
	req := httptest.NewRequest(http.MethodPost, "/api/xui/bind", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing tag/port, got %d", w.Code)
	}

	// 参数通过 Query 传 port 且 host 为 direct
	q := url.Values{}
	q.Set("port", "443")
	q.Set("host", "direct")
	req = httptest.NewRequest(http.MethodPost, "/api/xui/bind?"+q.Encode(), nil)
	w = httptest.NewRecorder()
	handler(w, req)
	if w.Code != http.StatusBadGateway {
		t.Logf("apiXUIBind returned %d (body: %s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "error") {
		t.Errorf("Expected error message in body, got %s", w.Body.String())
	}
}

func TestAPISubscription(t *testing.T) {
	testDir := t.TempDir()
	mgr := &Manager{
		workDir: testDir,
	}
	handler := apiSubscription(mgr)
	req := httptest.NewRequest(http.MethodGet, "/sub", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	t.Logf("apiSubscription status: %d", w.Code)
}
