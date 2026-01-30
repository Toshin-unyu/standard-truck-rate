package service

import (
	"errors"
	"testing"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// モック用のApiUsageRepository
type mockApiUsageRepository struct {
	usage        *model.ApiUsage
	getErr       error
	incrementErr error
}

func newMockApiUsageRepository(requestCount, limitCount int) *mockApiUsageRepository {
	return &mockApiUsageRepository{
		usage: &model.ApiUsage{
			YearMonth:    time.Now().Format("2006-01"),
			RequestCount: requestCount,
			LimitCount:   limitCount,
			LastUpdated:  time.Now(),
		},
	}
}

func (m *mockApiUsageRepository) GetOrCreateCurrent() (*model.ApiUsage, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.usage, nil
}

func (m *mockApiUsageRepository) IncrementCount(yearMonth string) error {
	if m.incrementErr != nil {
		return m.incrementErr
	}
	m.usage.RequestCount++
	return nil
}

// TestApiUsageService_CheckLimit_OK 制限内のテスト
func TestApiUsageService_CheckLimit_OK(t *testing.T) {
	repo := newMockApiUsageRepository(100, 9000)
	service := NewApiUsageService(repo)

	err := service.CheckLimit()
	if err != nil {
		t.Errorf("CheckLimit() 制限内でエラーが発生: %v", err)
	}
}

// TestApiUsageService_CheckLimit_Exceeded 制限超過のテスト
func TestApiUsageService_CheckLimit_Exceeded(t *testing.T) {
	repo := newMockApiUsageRepository(9000, 9000)
	service := NewApiUsageService(repo)

	err := service.CheckLimit()
	if err == nil {
		t.Error("CheckLimit() 制限超過でエラーが発生しない")
	}
}

// TestApiUsageService_CheckLimit_NearLimit 制限ギリギリのテスト
func TestApiUsageService_CheckLimit_NearLimit(t *testing.T) {
	repo := newMockApiUsageRepository(8999, 9000)
	service := NewApiUsageService(repo)

	err := service.CheckLimit()
	if err != nil {
		t.Errorf("CheckLimit() 制限ギリギリでエラーが発生: %v", err)
	}
}

// TestApiUsageService_IncrementAndCheck_OK 正常なインクリメント
func TestApiUsageService_IncrementAndCheck_OK(t *testing.T) {
	repo := newMockApiUsageRepository(100, 9000)
	service := NewApiUsageService(repo)

	err := service.IncrementAndCheck()
	if err != nil {
		t.Errorf("IncrementAndCheck() エラーが発生: %v", err)
	}

	// カウントが増えていることを確認
	if repo.usage.RequestCount != 101 {
		t.Errorf("IncrementAndCheck() RequestCount = %d, want 101", repo.usage.RequestCount)
	}
}

// TestApiUsageService_IncrementAndCheck_LimitExceeded 制限超過時のインクリメント
func TestApiUsageService_IncrementAndCheck_LimitExceeded(t *testing.T) {
	repo := newMockApiUsageRepository(9000, 9000)
	service := NewApiUsageService(repo)

	err := service.IncrementAndCheck()
	if err == nil {
		t.Error("IncrementAndCheck() 制限超過でエラーが発生しない")
	}

	// カウントが増えていないことを確認
	if repo.usage.RequestCount != 9000 {
		t.Errorf("IncrementAndCheck() 制限超過時にカウントが増えた: %d", repo.usage.RequestCount)
	}
}

// TestApiUsageService_GetStats 統計情報取得
func TestApiUsageService_GetStats(t *testing.T) {
	repo := newMockApiUsageRepository(4500, 9000)
	service := NewApiUsageService(repo)

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() エラーが発生: %v", err)
	}

	if stats.RequestCount != 4500 {
		t.Errorf("GetStats() RequestCount = %d, want 4500", stats.RequestCount)
	}
	if stats.LimitCount != 9000 {
		t.Errorf("GetStats() LimitCount = %d, want 9000", stats.LimitCount)
	}
	if stats.UsagePercent != 50.0 {
		t.Errorf("GetStats() UsagePercent = %f, want 50.0", stats.UsagePercent)
	}
	if stats.Remaining != 4500 {
		t.Errorf("GetStats() Remaining = %d, want 4500", stats.Remaining)
	}
}

// TestApiUsageService_GetStats_Warning 警告レベル
func TestApiUsageService_GetStats_Warning(t *testing.T) {
	repo := newMockApiUsageRepository(8000, 9000) // 約89%
	service := NewApiUsageService(repo)

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() エラーが発生: %v", err)
	}

	if stats.Level != "warning" {
		t.Errorf("GetStats() Level = %s, want warning", stats.Level)
	}
}

// TestApiUsageService_GetStats_Critical 危険レベル
func TestApiUsageService_GetStats_Critical(t *testing.T) {
	repo := newMockApiUsageRepository(8600, 9000) // 約96%
	service := NewApiUsageService(repo)

	stats, err := service.GetStats()
	if err != nil {
		t.Fatalf("GetStats() エラーが発生: %v", err)
	}

	if stats.Level != "critical" {
		t.Errorf("GetStats() Level = %s, want critical", stats.Level)
	}
}

// TestApiUsageService_GetStats_RepoError リポジトリエラー
func TestApiUsageService_GetStats_RepoError(t *testing.T) {
	repo := newMockApiUsageRepository(0, 9000)
	repo.getErr = errors.New("DB error")
	service := NewApiUsageService(repo)

	_, err := service.GetStats()
	if err == nil {
		t.Error("GetStats() リポジトリエラー時にエラーが返らない")
	}
}

// TestCacheStats キャッシュ統計
func TestCacheStats_Record(t *testing.T) {
	stats := NewCacheStats()

	// ヒットを記録
	stats.RecordHit()
	stats.RecordHit()
	stats.RecordHit()

	// ミスを記録
	stats.RecordMiss()

	if stats.Hits != 3 {
		t.Errorf("CacheStats.Hits = %d, want 3", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("CacheStats.Misses = %d, want 1", stats.Misses)
	}
}

// TestCacheStats_HitRate ヒット率計算
func TestCacheStats_HitRate(t *testing.T) {
	stats := NewCacheStats()

	// ヒット率 0%（データなし）
	if stats.HitRate() != 0 {
		t.Errorf("HitRate() 初期値 = %f, want 0", stats.HitRate())
	}

	// ヒット3、ミス1 → 75%
	stats.RecordHit()
	stats.RecordHit()
	stats.RecordHit()
	stats.RecordMiss()

	rate := stats.HitRate()
	if rate != 75.0 {
		t.Errorf("HitRate() = %f, want 75.0", rate)
	}
}

// TestCacheStats_Total トータルカウント
func TestCacheStats_Total(t *testing.T) {
	stats := NewCacheStats()

	stats.RecordHit()
	stats.RecordHit()
	stats.RecordMiss()

	if stats.Total() != 3 {
		t.Errorf("Total() = %d, want 3", stats.Total())
	}
}

// TestCacheStats_Reset リセット
func TestCacheStats_Reset(t *testing.T) {
	stats := NewCacheStats()

	stats.RecordHit()
	stats.RecordMiss()
	stats.Reset()

	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("Reset() 後 Hits=%d, Misses=%d", stats.Hits, stats.Misses)
	}
}
