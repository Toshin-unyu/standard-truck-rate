package repository

import (
	"testing"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

func TestApiUsageRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	usage := &model.ApiUsage{
		YearMonth:    "2026-01",
		RequestCount: 100,
		LimitCount:   9000,
	}

	err := repo.Create(usage)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestApiUsageRepository_GetByYearMonth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	// テストデータ作成
	usage := &model.ApiUsage{
		YearMonth:    "2026-01",
		RequestCount: 100,
		LimitCount:   9000,
	}
	repo.Create(usage)

	// 取得テスト
	got, err := repo.GetByYearMonth("2026-01")
	if err != nil {
		t.Fatalf("GetByYearMonth() error = %v", err)
	}
	if got.RequestCount != 100 || got.LimitCount != 9000 {
		t.Errorf("GetByYearMonth() = %+v", got)
	}
}

func TestApiUsageRepository_GetByYearMonth_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	_, err := repo.GetByYearMonth("9999-12")
	if err == nil {
		t.Error("GetByYearMonth() 存在しない年月でエラーが返らない")
	}
}

func TestApiUsageRepository_GetCurrent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	// 現在の年月のデータを作成
	currentYearMonth := time.Now().Format("2006-01")
	usage := &model.ApiUsage{
		YearMonth:    currentYearMonth,
		RequestCount: 50,
		LimitCount:   9000,
	}
	repo.Create(usage)

	// 取得テスト
	got, err := repo.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent() error = %v", err)
	}
	if got.YearMonth != currentYearMonth {
		t.Errorf("GetCurrent() YearMonth = %s, want %s", got.YearMonth, currentYearMonth)
	}
}

func TestApiUsageRepository_GetOrCreateCurrent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	// データがない状態で呼び出し
	got, err := repo.GetOrCreateCurrent()
	if err != nil {
		t.Fatalf("GetOrCreateCurrent() error = %v", err)
	}

	currentYearMonth := time.Now().Format("2006-01")
	if got.YearMonth != currentYearMonth {
		t.Errorf("GetOrCreateCurrent() YearMonth = %s, want %s", got.YearMonth, currentYearMonth)
	}
	if got.RequestCount != 0 {
		t.Errorf("GetOrCreateCurrent() RequestCount = %d, want 0", got.RequestCount)
	}
	if got.LimitCount != 9000 {
		t.Errorf("GetOrCreateCurrent() LimitCount = %d, want 9000", got.LimitCount)
	}
}

func TestApiUsageRepository_IncrementCount(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	// テストデータ作成
	usage := &model.ApiUsage{
		YearMonth:    "2026-01",
		RequestCount: 100,
		LimitCount:   9000,
	}
	repo.Create(usage)

	// インクリメント
	err := repo.IncrementCount("2026-01")
	if err != nil {
		t.Fatalf("IncrementCount() error = %v", err)
	}

	// 確認
	got, _ := repo.GetByYearMonth("2026-01")
	if got.RequestCount != 101 {
		t.Errorf("IncrementCount() RequestCount = %d, want 101", got.RequestCount)
	}
}

func TestApiUsageRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	// テストデータ作成
	usage := &model.ApiUsage{
		YearMonth:    "2026-01",
		RequestCount: 100,
		LimitCount:   9000,
	}
	repo.Create(usage)

	// 更新
	usage.RequestCount = 500
	usage.LimitCount = 10000
	err := repo.Update(usage)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 確認
	got, _ := repo.GetByYearMonth("2026-01")
	if got.RequestCount != 500 || got.LimitCount != 10000 {
		t.Errorf("Update() = %+v", got)
	}
}

func TestApiUsageRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewApiUsageRepository(db.MainDB())

	// テストデータ作成
	repo.Create(&model.ApiUsage{YearMonth: "2026-01", RequestCount: 100, LimitCount: 9000})
	repo.Create(&model.ApiUsage{YearMonth: "2026-02", RequestCount: 200, LimitCount: 9000})

	got, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAll() returned %d items, want 2", len(got))
	}
}
