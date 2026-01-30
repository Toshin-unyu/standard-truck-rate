package repository

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

func TestHighwayICRepository_Create(t *testing.T) {
	db := setupMainTestDB(t)
	defer db.Close()

	repo := NewHighwayICRepository(db)

	ic := &model.HighwayIC{
		Code:      "1010001",
		Name:      "東京",
		Yomi:      "とうきょう",
		Type:      model.ICTypeIC,
		RoadNo:    "1010",
		RoadName:  "【E1】東名高速道路",
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ic)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 作成されたデータを確認
	got, err := repo.GetByCode("1010001")
	if err != nil {
		t.Fatalf("GetByCode failed: %v", err)
	}

	if got.Name != "東京" {
		t.Errorf("Name: 期待=東京, 実際=%s", got.Name)
	}
	if got.Yomi != "とうきょう" {
		t.Errorf("Yomi: 期待=とうきょう, 実際=%s", got.Yomi)
	}
}

func TestHighwayICRepository_BulkCreate(t *testing.T) {
	db := setupMainTestDB(t)
	defer db.Close()

	repo := NewHighwayICRepository(db)

	ics := []*model.HighwayIC{
		{Code: "1010001", Name: "東京", Yomi: "とうきょう", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
		{Code: "1010002", Name: "用賀", Yomi: "ようが", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
		{Code: "1010003", Name: "川崎", Yomi: "かわさき", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
	}

	err := repo.BulkCreate(ics)
	if err != nil {
		t.Fatalf("BulkCreate failed: %v", err)
	}

	// 件数確認
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("Count: 期待=3, 実際=%d", count)
	}
}

func TestHighwayICRepository_SearchByName(t *testing.T) {
	db := setupMainTestDB(t)
	defer db.Close()

	repo := NewHighwayICRepository(db)

	// テストデータ作成
	ics := []*model.HighwayIC{
		{Code: "1010001", Name: "東京", Yomi: "とうきょう", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
		{Code: "1010002", Name: "東京外環", Yomi: "とうきょうがいかん", Type: 1, RoadNo: "1020", RoadName: "【C3】東京外環自動車道"},
		{Code: "1010003", Name: "川崎", Yomi: "かわさき", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
	}
	repo.BulkCreate(ics)

	// 名前で検索
	results, err := repo.SearchByName("東京")
	if err != nil {
		t.Fatalf("SearchByName failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("検索結果: 期待=2件, 実際=%d件", len(results))
	}
}

func TestHighwayICRepository_SearchByYomi(t *testing.T) {
	db := setupMainTestDB(t)
	defer db.Close()

	repo := NewHighwayICRepository(db)

	// テストデータ作成
	ics := []*model.HighwayIC{
		{Code: "1010001", Name: "東京", Yomi: "とうきょう", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
		{Code: "1010002", Name: "用賀", Yomi: "ようが", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
	}
	repo.BulkCreate(ics)

	// 読みで検索
	results, err := repo.SearchByYomi("とう")
	if err != nil {
		t.Fatalf("SearchByYomi failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("検索結果: 期待=1件, 実際=%d件", len(results))
	}
}

func TestHighwayICRepository_GetAll(t *testing.T) {
	db := setupMainTestDB(t)
	defer db.Close()

	repo := NewHighwayICRepository(db)

	// テストデータ作成
	ics := []*model.HighwayIC{
		{Code: "1010001", Name: "東京", Yomi: "とうきょう", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
		{Code: "1010002", Name: "用賀", Yomi: "ようが", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
	}
	repo.BulkCreate(ics)

	// 全件取得
	results, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("件数: 期待=2, 実際=%d", len(results))
	}
}

func TestHighwayICRepository_DeleteAll(t *testing.T) {
	db := setupMainTestDB(t)
	defer db.Close()

	repo := NewHighwayICRepository(db)

	// テストデータ作成
	ics := []*model.HighwayIC{
		{Code: "1010001", Name: "東京", Yomi: "とうきょう", Type: 1, RoadNo: "1010", RoadName: "【E1】東名高速道路"},
	}
	repo.BulkCreate(ics)

	// 全削除
	err := repo.DeleteAll()
	if err != nil {
		t.Fatalf("DeleteAll failed: %v", err)
	}

	count, _ := repo.Count()
	if count != 0 {
		t.Errorf("Count: 期待=0, 実際=%d", count)
	}
}

// ヘルパー関数
func setupMainTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_str.db")
	db, err := database.InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	return db
}
