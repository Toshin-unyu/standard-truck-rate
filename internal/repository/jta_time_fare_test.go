package repository

import (
	"path/filepath"
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	tmpDir := t.TempDir()
	db, err := NewDB(
		filepath.Join(tmpDir, "main.db"),
		filepath.Join(tmpDir, "cache.db"),
	)
	if err != nil {
		t.Fatalf("テストDB作成失敗: %v", err)
	}
	return db
}

// === JtaTimeBaseFare テスト ===

func TestJtaTimeFareRepository_CreateBaseFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	fare := &model.JtaTimeBaseFare{
		RegionCode:  3, // 関東
		VehicleCode: 2, // 中型車
		Hours:       4,
		BaseKm:      30,
		FareYen:     15000,
	}

	id, err := repo.CreateBaseFare(fare)
	if err != nil {
		t.Fatalf("CreateBaseFare() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateBaseFare() returned invalid id = %d", id)
	}
}

func TestJtaTimeFareRepository_GetBaseFareByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	fare := &model.JtaTimeBaseFare{
		RegionCode:  3,
		VehicleCode: 2,
		Hours:       4,
		BaseKm:      30,
		FareYen:     15000,
	}
	id, _ := repo.CreateBaseFare(fare)

	// 取得テスト
	got, err := repo.GetBaseFareByID(id)
	if err != nil {
		t.Fatalf("GetBaseFareByID() error = %v", err)
	}
	if got.RegionCode != 3 || got.VehicleCode != 2 || got.Hours != 4 {
		t.Errorf("GetBaseFareByID() = %+v, want RegionCode=3, VehicleCode=2, Hours=4", got)
	}
}

func TestJtaTimeFareRepository_GetBaseFareByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	_, err := repo.GetBaseFareByID(99999)
	if err == nil {
		t.Error("GetBaseFareByID() 存在しないIDでエラーが返らない")
	}
}

func TestJtaTimeFareRepository_GetAllBaseFares(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	fares := []*model.JtaTimeBaseFare{
		{RegionCode: 1, VehicleCode: 1, Hours: 4, BaseKm: 20, FareYen: 10000},
		{RegionCode: 2, VehicleCode: 2, Hours: 8, BaseKm: 40, FareYen: 20000},
	}
	for _, f := range fares {
		repo.CreateBaseFare(f)
	}

	got, err := repo.GetAllBaseFares()
	if err != nil {
		t.Fatalf("GetAllBaseFares() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllBaseFares() returned %d items, want 2", len(got))
	}
}

func TestJtaTimeFareRepository_UpdateBaseFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	fare := &model.JtaTimeBaseFare{
		RegionCode:  3,
		VehicleCode: 2,
		Hours:       4,
		BaseKm:      30,
		FareYen:     15000,
	}
	id, _ := repo.CreateBaseFare(fare)

	// 更新
	fare.ID = id
	fare.FareYen = 16000
	err := repo.UpdateBaseFare(fare)
	if err != nil {
		t.Fatalf("UpdateBaseFare() error = %v", err)
	}

	// 確認
	got, _ := repo.GetBaseFareByID(id)
	if got.FareYen != 16000 {
		t.Errorf("UpdateBaseFare() FareYen = %d, want 16000", got.FareYen)
	}
}

func TestJtaTimeFareRepository_DeleteBaseFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	fare := &model.JtaTimeBaseFare{
		RegionCode:  3,
		VehicleCode: 2,
		Hours:       4,
		BaseKm:      30,
		FareYen:     15000,
	}
	id, _ := repo.CreateBaseFare(fare)

	// 削除
	err := repo.DeleteBaseFare(id)
	if err != nil {
		t.Fatalf("DeleteBaseFare() error = %v", err)
	}

	// 確認
	_, err = repo.GetBaseFareByID(id)
	if err == nil {
		t.Error("DeleteBaseFare() 削除後もデータが取得できる")
	}
}

// === JtaTimeSurcharge テスト ===

func TestJtaTimeFareRepository_CreateSurcharge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	surcharge := &model.JtaTimeSurcharge{
		RegionCode:    3,
		VehicleCode:   2,
		SurchargeType: "distance",
		FareYen:       500,
	}

	id, err := repo.CreateSurcharge(surcharge)
	if err != nil {
		t.Fatalf("CreateSurcharge() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateSurcharge() returned invalid id = %d", id)
	}
}

func TestJtaTimeFareRepository_GetSurchargeByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	surcharge := &model.JtaTimeSurcharge{
		RegionCode:    3,
		VehicleCode:   2,
		SurchargeType: "distance",
		FareYen:       500,
	}
	id, _ := repo.CreateSurcharge(surcharge)

	// 取得テスト
	got, err := repo.GetSurchargeByID(id)
	if err != nil {
		t.Fatalf("GetSurchargeByID() error = %v", err)
	}
	if got.SurchargeType != "distance" || got.FareYen != 500 {
		t.Errorf("GetSurchargeByID() = %+v", got)
	}
}

func TestJtaTimeFareRepository_GetAllSurcharges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	surcharges := []*model.JtaTimeSurcharge{
		{RegionCode: 1, VehicleCode: 1, SurchargeType: "distance", FareYen: 400},
		{RegionCode: 1, VehicleCode: 1, SurchargeType: "time", FareYen: 600},
	}
	for _, s := range surcharges {
		repo.CreateSurcharge(s)
	}

	got, err := repo.GetAllSurcharges()
	if err != nil {
		t.Fatalf("GetAllSurcharges() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllSurcharges() returned %d items, want 2", len(got))
	}
}

func TestJtaTimeFareRepository_UpdateSurcharge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	surcharge := &model.JtaTimeSurcharge{
		RegionCode:    3,
		VehicleCode:   2,
		SurchargeType: "distance",
		FareYen:       500,
	}
	id, _ := repo.CreateSurcharge(surcharge)

	// 更新
	surcharge.ID = id
	surcharge.FareYen = 550
	err := repo.UpdateSurcharge(surcharge)
	if err != nil {
		t.Fatalf("UpdateSurcharge() error = %v", err)
	}

	// 確認
	got, _ := repo.GetSurchargeByID(id)
	if got.FareYen != 550 {
		t.Errorf("UpdateSurcharge() FareYen = %d, want 550", got.FareYen)
	}
}

func TestJtaTimeFareRepository_DeleteSurcharge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewJtaTimeFareRepository(db.MainDB())

	// テストデータ作成
	surcharge := &model.JtaTimeSurcharge{
		RegionCode:    3,
		VehicleCode:   2,
		SurchargeType: "distance",
		FareYen:       500,
	}
	id, _ := repo.CreateSurcharge(surcharge)

	// 削除
	err := repo.DeleteSurcharge(id)
	if err != nil {
		t.Fatalf("DeleteSurcharge() error = %v", err)
	}

	// 確認
	_, err = repo.GetSurchargeByID(id)
	if err == nil {
		t.Error("DeleteSurcharge() 削除後もデータが取得できる")
	}
}
