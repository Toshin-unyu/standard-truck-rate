package repository

import (
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

func TestRouteCacheRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	cache := &model.RouteCache{
		Origin:      "東京都新宿区",
		Dest:        "神奈川県横浜市",
		DistanceKm:  35.5,
		DurationMin: 45,
	}

	err := repo.Create(cache)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestRouteCacheRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	// テストデータ作成
	cache := &model.RouteCache{
		Origin:      "東京都新宿区",
		Dest:        "神奈川県横浜市",
		DistanceKm:  35.5,
		DurationMin: 45,
	}
	repo.Create(cache)

	// 取得テスト
	got, err := repo.Get("東京都新宿区", "神奈川県横浜市")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.DistanceKm != 35.5 || got.DurationMin != 45 {
		t.Errorf("Get() = %+v", got)
	}
}

func TestRouteCacheRepository_Get_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	_, err := repo.Get("存在しない場所", "存在しない場所")
	if err == nil {
		t.Error("Get() 存在しないキーでエラーが返らない")
	}
}

func TestRouteCacheRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	// テストデータ作成
	repo.Create(&model.RouteCache{Origin: "東京", Dest: "横浜", DistanceKm: 30, DurationMin: 40})
	repo.Create(&model.RouteCache{Origin: "東京", Dest: "大阪", DistanceKm: 500, DurationMin: 360})

	got, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAll() returned %d items, want 2", len(got))
	}
}

func TestRouteCacheRepository_Upsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	// 初回作成
	cache := &model.RouteCache{
		Origin:      "東京",
		Dest:        "横浜",
		DistanceKm:  30,
		DurationMin: 40,
	}
	err := repo.Upsert(cache)
	if err != nil {
		t.Fatalf("Upsert() 初回 error = %v", err)
	}

	// 同じキーで更新
	cache.DistanceKm = 35
	cache.DurationMin = 50
	err = repo.Upsert(cache)
	if err != nil {
		t.Fatalf("Upsert() 更新 error = %v", err)
	}

	// 確認
	got, _ := repo.Get("東京", "横浜")
	if got.DistanceKm != 35 || got.DurationMin != 50 {
		t.Errorf("Upsert() 更新後 = %+v, want DistanceKm=35, DurationMin=50", got)
	}
}

func TestRouteCacheRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	// テストデータ作成
	cache := &model.RouteCache{Origin: "東京", Dest: "横浜", DistanceKm: 30, DurationMin: 40}
	repo.Create(cache)

	// 削除
	err := repo.Delete("東京", "横浜")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 確認
	_, err = repo.Get("東京", "横浜")
	if err == nil {
		t.Error("Delete() 削除後もデータが取得できる")
	}
}

func TestRouteCacheRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRouteCacheRepository(db.CacheDB())

	// 存在しない場合
	if repo.Exists("東京", "横浜") {
		t.Error("Exists() 存在しないのにtrueを返した")
	}

	// データ作成後
	repo.Create(&model.RouteCache{Origin: "東京", Dest: "横浜", DistanceKm: 30, DurationMin: 40})

	if !repo.Exists("東京", "横浜") {
		t.Error("Exists() 存在するのにfalseを返した")
	}
}
