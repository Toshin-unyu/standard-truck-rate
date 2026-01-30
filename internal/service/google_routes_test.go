package service

import (
	"errors"
	"testing"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// モック用のRouteCacheRepository
type mockRouteCacheRepository struct {
	cache     map[string]*model.RouteCache
	getCalled bool
	upsertErr error
}

func newMockRouteCacheRepository() *mockRouteCacheRepository {
	return &mockRouteCacheRepository{
		cache: make(map[string]*model.RouteCache),
	}
}

func (m *mockRouteCacheRepository) Get(origin, dest string) (*model.RouteCache, error) {
	m.getCalled = true
	key := origin + "|" + dest
	if c, ok := m.cache[key]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (m *mockRouteCacheRepository) Upsert(cache *model.RouteCache) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	key := cache.Origin + "|" + cache.Dest
	m.cache[key] = cache
	return nil
}

func (m *mockRouteCacheRepository) setCache(origin, dest string, distanceKm float64, durationMin int, createdAt time.Time) {
	key := origin + "|" + dest
	m.cache[key] = &model.RouteCache{
		Origin:      origin,
		Dest:        dest,
		DistanceKm:  distanceKm,
		DurationMin: durationMin,
		CreatedAt:   createdAt,
	}
}

// TestMockRoutesClient_GetRoute モッククライアントのテスト
func TestMockRoutesClient_GetRoute(t *testing.T) {
	client := NewMockRoutesClient()

	route, err := client.GetRoute("東京都千代田区", "大阪府大阪市")
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	if route.DistanceKm <= 0 {
		t.Errorf("距離が0以下です: %f", route.DistanceKm)
	}
	if route.DurationMin <= 0 {
		t.Errorf("所要時間が0以下です: %d", route.DurationMin)
	}
	if route.Origin != "東京都千代田区" {
		t.Errorf("出発地が一致しません: %s", route.Origin)
	}
	if route.Dest != "大阪府大阪市" {
		t.Errorf("目的地が一致しません: %s", route.Dest)
	}
}

// TestMockRoutesClient_GetRoute_EmptyAddress 空アドレスのテスト
func TestMockRoutesClient_GetRoute_EmptyAddress(t *testing.T) {
	client := NewMockRoutesClient()

	_, err := client.GetRoute("", "大阪府大阪市")
	if err == nil {
		t.Error("空の出発地でエラーが発生しませんでした")
	}

	_, err = client.GetRoute("東京都千代田区", "")
	if err == nil {
		t.Error("空の目的地でエラーが発生しませんでした")
	}
}

// TestCachedRouteService_GetRoute_CacheHit キャッシュヒット時のテスト
func TestCachedRouteService_GetRoute_CacheHit(t *testing.T) {
	mockRepo := newMockRouteCacheRepository()
	mockClient := NewMockRoutesClient()

	// キャッシュを事前に設定（有効期限内）
	mockRepo.setCache("東京都千代田区", "大阪府大阪市", 500.0, 360, time.Now())

	service := NewCachedRouteService(mockClient, mockRepo, 30*24*time.Hour)

	route, err := service.GetRoute("東京都千代田区", "大阪府大阪市")
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	// キャッシュの値が返されることを確認
	if route.DistanceKm != 500.0 {
		t.Errorf("キャッシュの距離が返されませんでした: %f", route.DistanceKm)
	}
	if route.DurationMin != 360 {
		t.Errorf("キャッシュの所要時間が返されませんでした: %d", route.DurationMin)
	}
}

// TestCachedRouteService_GetRoute_CacheMiss キャッシュミス時のテスト
func TestCachedRouteService_GetRoute_CacheMiss(t *testing.T) {
	mockRepo := newMockRouteCacheRepository()
	mockClient := NewMockRoutesClient()

	service := NewCachedRouteService(mockClient, mockRepo, 30*24*time.Hour)

	route, err := service.GetRoute("東京都千代田区", "大阪府大阪市")
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	// APIから取得した値が返されることを確認
	if route.DistanceKm <= 0 {
		t.Errorf("距離が0以下です: %f", route.DistanceKm)
	}

	// キャッシュに保存されていることを確認
	cached, err := mockRepo.Get("東京都千代田区", "大阪府大阪市")
	if err != nil {
		t.Errorf("キャッシュに保存されていません: %v", err)
	}
	if cached.DistanceKm != route.DistanceKm {
		t.Errorf("キャッシュの値が一致しません")
	}
}

// TestCachedRouteService_GetRoute_CacheExpired キャッシュ期限切れのテスト
func TestCachedRouteService_GetRoute_CacheExpired(t *testing.T) {
	mockRepo := newMockRouteCacheRepository()
	mockClient := NewMockRoutesClient()

	// 期限切れのキャッシュを設定（31日前）
	expiredTime := time.Now().Add(-31 * 24 * time.Hour)
	mockRepo.setCache("東京都千代田区", "大阪府大阪市", 500.0, 360, expiredTime)

	service := NewCachedRouteService(mockClient, mockRepo, 30*24*time.Hour)

	route, err := service.GetRoute("東京都千代田区", "大阪府大阪市")
	if err != nil {
		t.Fatalf("エラーが発生しました: %v", err)
	}

	// 期限切れのため、APIから新しい値が取得されることを確認
	// モッククライアントは固定値を返すので、キャッシュの500.0とは異なるはず
	// （モッククライアントの実装次第だが、少なくともキャッシュが更新されるべき）
	cached, _ := mockRepo.Get("東京都千代田区", "大阪府大阪市")
	if cached.CreatedAt.Before(expiredTime) {
		t.Errorf("キャッシュが更新されていません")
	}

	if route.DistanceKm <= 0 {
		t.Errorf("距離が0以下です: %f", route.DistanceKm)
	}
}

// TestGoogleRoutesClient_GetRoute_NoAPIKey APIキーなしのテスト
func TestGoogleRoutesClient_GetRoute_NoAPIKey(t *testing.T) {
	client := NewGoogleRoutesClient("")

	_, err := client.GetRoute("東京都千代田区", "大阪府大阪市")
	if err == nil {
		t.Error("APIキーなしでエラーが発生しませんでした")
	}
}

// TestRouteResult_Validation ルート結果の検証テスト
func TestRouteResult_Validation(t *testing.T) {
	tests := []struct {
		name        string
		origin      string
		dest        string
		expectError bool
	}{
		{"正常ケース", "東京都千代田区", "大阪府大阪市", false},
		{"同一地点", "東京都千代田区", "東京都千代田区", true},
		{"出発地が空", "", "大阪府大阪市", true},
		{"目的地が空", "東京都千代田区", "", true},
		{"両方空", "", "", true},
	}

	client := NewMockRoutesClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetRoute(tt.origin, tt.dest)
			if tt.expectError && err == nil {
				t.Errorf("エラーが期待されましたが発生しませんでした")
			}
			if !tt.expectError && err != nil {
				t.Errorf("エラーが発生しました: %v", err)
			}
		})
	}
}
