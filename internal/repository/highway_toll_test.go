package repository

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

func TestHighwayTollRepository_Create(t *testing.T) {
	db := setupCacheTestDB(t)
	defer db.Close()

	repo := NewHighwayTollRepository(db)

	toll := &model.HighwayToll{
		OriginIC:    "東京",
		DestIC:      "名古屋",
		CarType:     model.CarTypeLarge,
		NormalToll:  8350,
		EtcToll:     5840,
		Etc2Toll:    5840,
		DistanceKm:  325.5,
		DurationMin: 210,
		CreatedAt:   time.Now(),
	}

	err := repo.Create(toll)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 作成されたデータを確認
	got, err := repo.Get("東京", "名古屋", model.CarTypeLarge)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.NormalToll != 8350 {
		t.Errorf("NormalToll: 期待=8350, 実際=%d", got.NormalToll)
	}
	if got.EtcToll != 5840 {
		t.Errorf("EtcToll: 期待=5840, 実際=%d", got.EtcToll)
	}
}

func TestHighwayTollRepository_Upsert(t *testing.T) {
	db := setupCacheTestDB(t)
	defer db.Close()

	repo := NewHighwayTollRepository(db)

	toll := &model.HighwayToll{
		OriginIC:    "東京",
		DestIC:      "名古屋",
		CarType:     model.CarTypeLarge,
		NormalToll:  8350,
		EtcToll:     5840,
		Etc2Toll:    5840,
		DistanceKm:  325.5,
		DurationMin: 210,
	}

	// 1回目：新規作成
	err := repo.Upsert(toll)
	if err != nil {
		t.Fatalf("Upsert(1回目) failed: %v", err)
	}

	// 2回目：更新
	toll.NormalToll = 9000
	toll.EtcToll = 6300
	err = repo.Upsert(toll)
	if err != nil {
		t.Fatalf("Upsert(2回目) failed: %v", err)
	}

	// 更新されたデータを確認
	got, err := repo.Get("東京", "名古屋", model.CarTypeLarge)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if got.NormalToll != 9000 {
		t.Errorf("NormalToll: 期待=9000, 実際=%d", got.NormalToll)
	}
	if got.EtcToll != 6300 {
		t.Errorf("EtcToll: 期待=6300, 実際=%d", got.EtcToll)
	}
}

func TestHighwayTollRepository_Exists(t *testing.T) {
	db := setupCacheTestDB(t)
	defer db.Close()

	repo := NewHighwayTollRepository(db)

	// データがない場合
	if repo.Exists("東京", "名古屋", model.CarTypeLarge) {
		t.Error("存在しないはずのデータが存在すると判定された")
	}

	// データを作成
	toll := &model.HighwayToll{
		OriginIC:    "東京",
		DestIC:      "名古屋",
		CarType:     model.CarTypeLarge,
		NormalToll:  8350,
		EtcToll:     5840,
		Etc2Toll:    5840,
		DistanceKm:  325.5,
		DurationMin: 210,
	}
	repo.Create(toll)

	// データがある場合
	if !repo.Exists("東京", "名古屋", model.CarTypeLarge) {
		t.Error("存在するはずのデータが存在しないと判定された")
	}
}

func TestHighwayTollRepository_Delete(t *testing.T) {
	db := setupCacheTestDB(t)
	defer db.Close()

	repo := NewHighwayTollRepository(db)

	// データを作成
	toll := &model.HighwayToll{
		OriginIC:    "東京",
		DestIC:      "名古屋",
		CarType:     model.CarTypeLarge,
		NormalToll:  8350,
		EtcToll:     5840,
		Etc2Toll:    5840,
		DistanceKm:  325.5,
		DurationMin: 210,
	}
	repo.Create(toll)

	// 削除
	err := repo.Delete("東京", "名古屋", model.CarTypeLarge)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 存在しないことを確認
	if repo.Exists("東京", "名古屋", model.CarTypeLarge) {
		t.Error("削除後もデータが存在している")
	}
}

// ヘルパー関数
func setupCacheTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_cache.db")
	db, err := database.InitCacheDB(dbPath)
	if err != nil {
		t.Fatalf("InitCacheDB failed: %v", err)
	}
	return db
}
