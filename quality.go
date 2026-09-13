package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// QualityInfo 是一个 IP 的网络属性与质量画像。
type QualityInfo struct {
	IP        string    `json:"ip"`
	Type      string    `json:"type"`       // "residential" | "datacenter" | "mobile"
	TypeLabel string    `json:"type_label"` // "🏠 家庭宽带" | "🏢 机房/IDC" | "📱 移动蜂窝"
	Score     int       `json:"score"`      // 纯净度/质量得分 0~100
	ISP       string    `json:"isp"`
	Org       string    `json:"org"`
	ASN       string    `json:"asn"`
	Hosting   bool      `json:"hosting"`
	Proxy     bool      `json:"proxy"`
	Mobile    bool      `json:"mobile"`
	UpdatedAt time.Time `json:"updated_at"`
}

const (
	defaultIPAPIBatchURL = "http://ip-api.com/batch"
	qualityCacheFile     = "quality_cache.json"
	qualityCacheTTL      = 24 * time.Hour
)

// EvaluateQuality 计算 IP 质量画像与纯净度得分。
func EvaluateQuality(ip, isp, org, asn string, hosting, proxy, mobile bool) QualityInfo {
	qType := "residential"
	typeLabel := "🏠 家庭宽带"
	score := 95

	if hosting {
		qType = "datacenter"
		typeLabel = "🏢 机房/IDC"
		score = 60
	} else if mobile {
		qType = "mobile"
		typeLabel = "📱 移动蜂窝"
		score = 88
	}

	// 被识别为已知代理/VPN，扣 25 分
	if proxy {
		score -= 25
	}

	// 知名家宽运营商加 5 分
	lowerInfo := strings.ToLower(isp + " " + org + " " + asn)
	residentialISPs := []string{
		"ocn", "ntt", "kddi", "softbank", "so-net", "biglobe", "optage", "plala", "j:com",
		"comcast", "charter", "spectrum", "at&t", "verizon", "centurylink", "cox",
		"chunghwa", "hinet", "so-net taiwan", "taiwan mobile", "kbro",
		"kt", "korea telecom", "sk broadband", "lg uplus", "hanaro",
		"telstra", "optus", "bell canada", "rogers", "shaw", "bt", "virgin media", "vodafone",
		"deutsche telekom", "orange", "free sas", "telefonica", "telecom italia",
	}
	for _, kw := range residentialISPs {
		if strings.Contains(lowerInfo, kw) {
			if !hosting {
				score += 5
			}
			break
		}
	}

	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	return QualityInfo{
		IP:        ip,
		Type:      qType,
		TypeLabel: typeLabel,
		Score:     score,
		ISP:       isp,
		Org:       org,
		ASN:       asn,
		Hosting:   hosting,
		Proxy:     proxy,
		Mobile:    mobile,
		UpdatedAt: time.Now(),
	}
}

// QualityCache 本地缓存与持久化管理器。
type QualityCache struct {
	mu      sync.RWMutex
	evalMu  sync.Mutex
	saveMu  sync.Mutex
	items   map[string]QualityInfo
	path    string
	apiURL  string
	client  *http.Client
}

// NewQualityCache 创建质量缓存管理器。
func NewQualityCache(workDir string) *QualityCache {
	filePath := qualityCacheFile
	if workDir != "" {
		filePath = filepath.Join(workDir, qualityCacheFile)
	}
	qc := &QualityCache{
		items:  make(map[string]QualityInfo),
		path:   filePath,
		apiURL: defaultIPAPIBatchURL,
		client: &http.Client{Timeout: 12 * time.Second},
	}
	qc.Load()
	return qc
}

// Load 从磁盘读取持久化缓存。
func (qc *QualityCache) Load() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	blob, err := os.ReadFile(qc.path)
	if err != nil {
		return
	}
	var loaded map[string]QualityInfo
	if err := json.Unmarshal(blob, &loaded); err == nil && loaded != nil {
		qc.items = loaded
	}
}

// Save 将缓存写入磁盘。
func (qc *QualityCache) Save() error {
	qc.saveMu.Lock()
	defer qc.saveMu.Unlock()
	qc.mu.RLock()
	blob, err := json.MarshalIndent(qc.items, "", "  ")
	qc.mu.RUnlock()
	if err != nil {
		return err
	}
	dir := filepath.Dir(qc.path)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0700)
	}
	return os.WriteFile(qc.path, blob, 0644)
}

// Get 读取指定 IP 的画像（TTL 24h）。
func (qc *QualityCache) Get(ip string) (QualityInfo, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	info, ok := qc.items[ip]
	if !ok {
		return QualityInfo{}, false
	}
	if info.Type == "unknown" || time.Since(info.UpdatedAt) > qualityCacheTTL {
		return QualityInfo{}, false
	}
	return info, true
}

// Set 写入指定 IP 的画像。
func (qc *QualityCache) Set(info QualityInfo) {
	if info.UpdatedAt.IsZero() {
		info.UpdatedAt = time.Now()
	}
	qc.mu.Lock()
	qc.items[info.IP] = info
	qc.mu.Unlock()
}

type ipAPIBatchItem struct {
	Query  string `json:"query"`
	Fields string `json:"fields"`
}

type ipAPIBatchRespItem struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
	AS          string `json:"as"`
	Mobile      bool   `json:"mobile"`
	Proxy       bool   `json:"proxy"`
	Hosting     bool   `json:"hosting"`
	Query       string `json:"query"`
}

// BatchEvaluate 批量评估 IP 质量画像。已缓存的直接复用，未缓存的切成 100 个/批发起 POST。
func (qc *QualityCache) BatchEvaluate(ips []string) map[string]QualityInfo {
	qc.evalMu.Lock()
	defer qc.evalMu.Unlock()

	result := make(map[string]QualityInfo, len(ips))
	needed := make([]string, 0, len(ips))

	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		if q, ok := qc.Get(ip); ok {
			result[ip] = q
		} else {
			needed = append(needed, ip)
		}
	}

	if len(needed) == 0 {
		return result
	}

	// 针对未命中缓存的 IP 去重
	uniqueNeeded := make([]string, 0, len(needed))
	seen := make(map[string]bool, len(needed))
	for _, ip := range needed {
		if !seen[ip] {
			seen[ip] = true
			uniqueNeeded = append(uniqueNeeded, ip)
		}
	}

	updatedAny := false
	// ip-api.com/batch 单次支持最多 100 个 IP
	chunkSize := 100
	for i := 0; i < len(uniqueNeeded); i += chunkSize {
		end := i + chunkSize
		if end > len(uniqueNeeded) {
			end = len(uniqueNeeded)
		}
		chunk := uniqueNeeded[i:end]

		reqItems := make([]ipAPIBatchItem, len(chunk))
		for j, ip := range chunk {
			reqItems[j] = ipAPIBatchItem{
				Query:  ip,
				Fields: "status,message,countryCode,city,isp,org,as,mobile,proxy,hosting,query",
			}
		}

		bodyBytes, err := json.Marshal(reqItems)
		if err != nil {
			log.Printf("构建 IP 质量评估请求失败: %v", err)
			for _, ip := range chunk {
				result[ip] = fallbackQuality(ip)
			}
			continue
		}

		req, err := http.NewRequest(http.MethodPost, qc.apiURL, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("创建 IP 质量请求失败: %v", err)
			for _, ip := range chunk {
				result[ip] = fallbackQuality(ip)
			}
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := qc.client.Do(req)
		if err != nil {
			log.Printf("IP 批量质量评估请求失败: %v", err)
			// 出错时不污染持久缓存，仅临时返回兜底评级
			for _, ip := range chunk {
				result[ip] = fallbackQuality(ip)
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			log.Printf("IP 质量评估响应异常 (HTTP %d): %s", resp.StatusCode, string(respBody))
			for _, ip := range chunk {
				result[ip] = fallbackQuality(ip)
			}
			continue
		}

		var respItems []ipAPIBatchRespItem
		if err := json.Unmarshal(respBody, &respItems); err != nil {
			log.Printf("解析 IP 质量响应失败: %v", err)
			for _, ip := range chunk {
				result[ip] = fallbackQuality(ip)
			}
			continue
		}

		for _, item := range respItems {
			ip := item.Query
			var q QualityInfo
			if item.Status == "success" {
				q = EvaluateQuality(ip, item.ISP, item.Org, item.AS, item.Hosting, item.Proxy, item.Mobile)
				qc.Set(q)
				updatedAny = true
			} else {
				q = fallbackQuality(ip)
			}
			result[ip] = q
		}

		// 确保当前 chunk 中所有未在响应中的 IP 均有返回值
		for _, ip := range chunk {
			if _, ok := result[ip]; !ok {
				result[ip] = fallbackQuality(ip)
			}
		}
	}

	if updatedAny {
		if err := qc.Save(); err != nil {
			log.Printf("保存 IP 质量缓存失败: %v", err)
		}
	}

	return result
}

// fallbackQuality 当外部接口不可达时的保守兜底评价
func fallbackQuality(ip string) QualityInfo {
	return QualityInfo{
		IP:        ip,
		Type:      "unknown",
		TypeLabel: "❓ 待检测",
		Score:     50,
		ISP:       "未知运营商",
		Org:       "",
		ASN:       "",
		Hosting:   false,
		Proxy:     false,
		Mobile:    false,
		UpdatedAt: time.Now(),
	}
}

// globalQualityCache 单例
var (
	globalQualityCache   *QualityCache
	globalQualityCacheMu sync.Mutex
)

// GetQualityCache 获取全局质量缓存单例
func GetQualityCache(workDir string) *QualityCache {
	globalQualityCacheMu.Lock()
	defer globalQualityCacheMu.Unlock()
	if globalQualityCache == nil {
		globalQualityCache = NewQualityCache(workDir)
	}
	return globalQualityCache
}
