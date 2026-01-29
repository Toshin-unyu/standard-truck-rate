package repository

import (
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// === AkabouDistanceFare テスト ===

func TestAkabouFareRepository_CreateDistanceFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	maxKm := 20
	baseFare := 5000
	perKmRate := 200
	fare := &model.AkabouDistanceFare{
		MinKm:     10,
		MaxKm:     &maxKm,
		BaseFare:  &baseFare,
		PerKmRate: &perKmRate,
	}

	id, err := repo.CreateDistanceFare(fare)
	if err != nil {
		t.Fatalf("CreateDistanceFare() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateDistanceFare() returned invalid id = %d", id)
	}
}

func TestAkabouFareRepository_GetDistanceFareByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	maxKm := 20
	baseFare := 5000
	perKmRate := 200
	fare := &model.AkabouDistanceFare{
		MinKm:     10,
		MaxKm:     &maxKm,
		BaseFare:  &baseFare,
		PerKmRate: &perKmRate,
	}
	id, _ := repo.CreateDistanceFare(fare)

	got, err := repo.GetDistanceFareByID(id)
	if err != nil {
		t.Fatalf("GetDistanceFareByID() error = %v", err)
	}
	if got.MinKm != 10 || *got.MaxKm != 20 {
		t.Errorf("GetDistanceFareByID() = %+v", got)
	}
}

func TestAkabouFareRepository_GetAllDistanceFares(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	maxKm := 20
	baseFare := 5000
	perKmRate := 200
	repo.CreateDistanceFare(&model.AkabouDistanceFare{MinKm: 0, MaxKm: &maxKm, BaseFare: &baseFare, PerKmRate: &perKmRate})
	repo.CreateDistanceFare(&model.AkabouDistanceFare{MinKm: 20, MaxKm: nil, BaseFare: nil, PerKmRate: &perKmRate})

	got, err := repo.GetAllDistanceFares()
	if err != nil {
		t.Fatalf("GetAllDistanceFares() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllDistanceFares() returned %d items, want 2", len(got))
	}
}

func TestAkabouFareRepository_UpdateDistanceFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	maxKm := 20
	baseFare := 5000
	perKmRate := 200
	fare := &model.AkabouDistanceFare{MinKm: 10, MaxKm: &maxKm, BaseFare: &baseFare, PerKmRate: &perKmRate}
	id, _ := repo.CreateDistanceFare(fare)

	newBaseFare := 5500
	fare.ID = id
	fare.BaseFare = &newBaseFare
	err := repo.UpdateDistanceFare(fare)
	if err != nil {
		t.Fatalf("UpdateDistanceFare() error = %v", err)
	}

	got, _ := repo.GetDistanceFareByID(id)
	if *got.BaseFare != 5500 {
		t.Errorf("UpdateDistanceFare() BaseFare = %d, want 5500", *got.BaseFare)
	}
}

func TestAkabouFareRepository_DeleteDistanceFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	maxKm := 20
	baseFare := 5000
	perKmRate := 200
	fare := &model.AkabouDistanceFare{MinKm: 10, MaxKm: &maxKm, BaseFare: &baseFare, PerKmRate: &perKmRate}
	id, _ := repo.CreateDistanceFare(fare)

	err := repo.DeleteDistanceFare(id)
	if err != nil {
		t.Fatalf("DeleteDistanceFare() error = %v", err)
	}

	_, err = repo.GetDistanceFareByID(id)
	if err == nil {
		t.Error("DeleteDistanceFare() 削除後もデータが取得できる")
	}
}

// === AkabouTimeFare テスト ===

func TestAkabouFareRepository_CreateTimeFare(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	fare := &model.AkabouTimeFare{
		BaseHours:    2,
		BaseKm:       20,
		BaseFare:     8000,
		OvertimeRate: 2000,
	}

	id, err := repo.CreateTimeFare(fare)
	if err != nil {
		t.Fatalf("CreateTimeFare() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateTimeFare() returned invalid id = %d", id)
	}
}

func TestAkabouFareRepository_GetTimeFareByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	fare := &model.AkabouTimeFare{BaseHours: 2, BaseKm: 20, BaseFare: 8000, OvertimeRate: 2000}
	id, _ := repo.CreateTimeFare(fare)

	got, err := repo.GetTimeFareByID(id)
	if err != nil {
		t.Fatalf("GetTimeFareByID() error = %v", err)
	}
	if got.BaseHours != 2 || got.BaseFare != 8000 {
		t.Errorf("GetTimeFareByID() = %+v", got)
	}
}

func TestAkabouFareRepository_GetAllTimeFares(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	repo.CreateTimeFare(&model.AkabouTimeFare{BaseHours: 2, BaseKm: 20, BaseFare: 8000, OvertimeRate: 2000})
	repo.CreateTimeFare(&model.AkabouTimeFare{BaseHours: 4, BaseKm: 40, BaseFare: 15000, OvertimeRate: 2500})

	got, err := repo.GetAllTimeFares()
	if err != nil {
		t.Fatalf("GetAllTimeFares() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllTimeFares() returned %d items, want 2", len(got))
	}
}

// === AkabouSurcharge テスト ===

func TestAkabouFareRepository_CreateSurcharge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	desc := "休日割増"
	surcharge := &model.AkabouSurcharge{
		SurchargeType: "holiday",
		RatePercent:   20,
		Description:   &desc,
	}

	id, err := repo.CreateSurcharge(surcharge)
	if err != nil {
		t.Fatalf("CreateSurcharge() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateSurcharge() returned invalid id = %d", id)
	}
}

func TestAkabouFareRepository_GetSurchargeByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	desc := "休日割増"
	surcharge := &model.AkabouSurcharge{SurchargeType: "holiday", RatePercent: 20, Description: &desc}
	id, _ := repo.CreateSurcharge(surcharge)

	got, err := repo.GetSurchargeByID(id)
	if err != nil {
		t.Fatalf("GetSurchargeByID() error = %v", err)
	}
	if got.SurchargeType != "holiday" || got.RatePercent != 20 {
		t.Errorf("GetSurchargeByID() = %+v", got)
	}
}

func TestAkabouFareRepository_GetAllSurcharges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	repo.CreateSurcharge(&model.AkabouSurcharge{SurchargeType: "holiday", RatePercent: 20, Description: nil})
	repo.CreateSurcharge(&model.AkabouSurcharge{SurchargeType: "night", RatePercent: 30, Description: nil})

	got, err := repo.GetAllSurcharges()
	if err != nil {
		t.Fatalf("GetAllSurcharges() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllSurcharges() returned %d items, want 2", len(got))
	}
}

// === AkabouAreaSurcharge テスト ===

func TestAkabouFareRepository_CreateAreaSurcharge(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	area := &model.AkabouAreaSurcharge{
		AreaName:        "東京都心",
		SurchargeAmount: 1000,
	}

	id, err := repo.CreateAreaSurcharge(area)
	if err != nil {
		t.Fatalf("CreateAreaSurcharge() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateAreaSurcharge() returned invalid id = %d", id)
	}
}

func TestAkabouFareRepository_GetAreaSurchargeByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	area := &model.AkabouAreaSurcharge{AreaName: "東京都心", SurchargeAmount: 1000}
	id, _ := repo.CreateAreaSurcharge(area)

	got, err := repo.GetAreaSurchargeByID(id)
	if err != nil {
		t.Fatalf("GetAreaSurchargeByID() error = %v", err)
	}
	if got.AreaName != "東京都心" || got.SurchargeAmount != 1000 {
		t.Errorf("GetAreaSurchargeByID() = %+v", got)
	}
}

func TestAkabouFareRepository_GetAllAreaSurcharges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	repo.CreateAreaSurcharge(&model.AkabouAreaSurcharge{AreaName: "東京都心", SurchargeAmount: 1000})
	repo.CreateAreaSurcharge(&model.AkabouAreaSurcharge{AreaName: "横浜", SurchargeAmount: 500})

	got, err := repo.GetAllAreaSurcharges()
	if err != nil {
		t.Fatalf("GetAllAreaSurcharges() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllAreaSurcharges() returned %d items, want 2", len(got))
	}
}

// === AkabouAdditionalFee テスト ===

func TestAkabouFareRepository_CreateAdditionalFee(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	fee := &model.AkabouAdditionalFee{
		FeeType:     "work",
		FreeMinutes: 30,
		UnitMinutes: 15,
		FeeAmount:   500,
	}

	id, err := repo.CreateAdditionalFee(fee)
	if err != nil {
		t.Fatalf("CreateAdditionalFee() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("CreateAdditionalFee() returned invalid id = %d", id)
	}
}

func TestAkabouFareRepository_GetAdditionalFeeByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	fee := &model.AkabouAdditionalFee{FeeType: "work", FreeMinutes: 30, UnitMinutes: 15, FeeAmount: 500}
	id, _ := repo.CreateAdditionalFee(fee)

	got, err := repo.GetAdditionalFeeByID(id)
	if err != nil {
		t.Fatalf("GetAdditionalFeeByID() error = %v", err)
	}
	if got.FeeType != "work" || got.FeeAmount != 500 {
		t.Errorf("GetAdditionalFeeByID() = %+v", got)
	}
}

func TestAkabouFareRepository_GetAllAdditionalFees(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewAkabouFareRepository(db.MainDB())

	repo.CreateAdditionalFee(&model.AkabouAdditionalFee{FeeType: "work", FreeMinutes: 30, UnitMinutes: 15, FeeAmount: 500})
	repo.CreateAdditionalFee(&model.AkabouAdditionalFee{FeeType: "waiting", FreeMinutes: 60, UnitMinutes: 30, FeeAmount: 1000})

	got, err := repo.GetAllAdditionalFees()
	if err != nil {
		t.Fatalf("GetAllAdditionalFees() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllAdditionalFees() returned %d items, want 2", len(got))
	}
}
