package service

import (
	"errors"
	"sync"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// ErrApiLimitExceeded API使用量制限超過エラー
var ErrApiLimitExceeded = errors.New("API使用量制限を超過しました")

// ApiUsageStore API使用量ストアインターフェース
type ApiUsageStore interface {
	GetOrCreateCurrent() (*model.ApiUsage, error)
	IncrementCount(yearMonth string) error
}

// ApiUsageService API使用量管理サービス
type ApiUsageService struct {
	store ApiUsageStore
}

// NewApiUsageService 新しいAPI使用量管理サービスを作成
func NewApiUsageService(store ApiUsageStore) *ApiUsageService {
	return &ApiUsageService{
		store: store,
	}
}

// CheckLimit 使用量が制限内かチェック
func (s *ApiUsageService) CheckLimit() error {
	usage, err := s.store.GetOrCreateCurrent()
	if err != nil {
		return err
	}

	if usage.RequestCount >= usage.LimitCount {
		return ErrApiLimitExceeded
	}

	return nil
}

// IncrementAndCheck 使用量をインクリメントし、制限をチェック
func (s *ApiUsageService) IncrementAndCheck() error {
	// まず制限をチェック
	usage, err := s.store.GetOrCreateCurrent()
	if err != nil {
		return err
	}

	if usage.RequestCount >= usage.LimitCount {
		return ErrApiLimitExceeded
	}

	// インクリメント
	return s.store.IncrementCount(usage.YearMonth)
}

// UsageStats 使用量統計
type UsageStats struct {
	YearMonth    string  `json:"year_month"`
	RequestCount int     `json:"request_count"`
	LimitCount   int     `json:"limit_count"`
	Remaining    int     `json:"remaining"`
	UsagePercent float64 `json:"usage_percent"`
	Level        string  `json:"level"` // "ok", "warning", "critical"
}

// GetStats 現在の使用量統計を取得
func (s *ApiUsageService) GetStats() (*UsageStats, error) {
	usage, err := s.store.GetOrCreateCurrent()
	if err != nil {
		return nil, err
	}

	stats := &UsageStats{
		YearMonth:    usage.YearMonth,
		RequestCount: usage.RequestCount,
		LimitCount:   usage.LimitCount,
		Remaining:    usage.LimitCount - usage.RequestCount,
		UsagePercent: usage.UsagePercent(),
		Level:        "ok",
	}

	if usage.IsCritical() {
		stats.Level = "critical"
	} else if usage.IsWarning() {
		stats.Level = "warning"
	}

	return stats, nil
}

// CacheStats キャッシュ統計
type CacheStats struct {
	mu     sync.RWMutex
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
}

// NewCacheStats 新しいキャッシュ統計を作成
func NewCacheStats() *CacheStats {
	return &CacheStats{}
}

// RecordHit キャッシュヒットを記録
func (c *CacheStats) RecordHit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Hits++
}

// RecordMiss キャッシュミスを記録
func (c *CacheStats) RecordMiss() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Misses++
}

// Total トータルリクエスト数
func (c *CacheStats) Total() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Hits + c.Misses
}

// HitRate ヒット率（%）
func (c *CacheStats) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.Hits + c.Misses
	if total == 0 {
		return 0
	}
	return float64(c.Hits) / float64(total) * 100
}

// Reset 統計をリセット
func (c *CacheStats) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Hits = 0
	c.Misses = 0
}

// GetSnapshot スナップショットを取得（スレッドセーフ）
func (c *CacheStats) GetSnapshot() CacheStatsSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStatsSnapshot{
		Hits:    c.Hits,
		Misses:  c.Misses,
		Total:   c.Hits + c.Misses,
		HitRate: c.HitRate(),
	}
}

// CacheStatsSnapshot キャッシュ統計のスナップショット
type CacheStatsSnapshot struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	Total   int64   `json:"total"`
	HitRate float64 `json:"hit_rate"`
}
