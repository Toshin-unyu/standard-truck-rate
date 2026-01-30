package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := database.InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("テストDB初期化失敗: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestSeedJtaTimeBaseFares(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// データ投入実行
	if err := seedJtaTimeFares(db); err != nil {
		t.Fatalf("JTA時間制運賃投入失敗: %v", err)
	}

	repo := repository.NewJtaTimeFareRepository(db)

	// 基礎額の件数確認（10地域 × 4車格 × 2制度 = 80件）
	baseFares, err := repo.GetAllBaseFares()
	if err != nil {
		t.Fatalf("基礎額取得失敗: %v", err)
	}
	if len(baseFares) != 80 {
		t.Errorf("基礎額件数: got %d, want 80", len(baseFares))
	}

	// 関東・大型車・8時間制の基礎額確認
	var kantoLarge8h *int
	for _, f := range baseFares {
		if f.RegionCode == 3 && f.VehicleCode == 3 && f.Hours == 8 {
			kantoLarge8h = &f.FareYen
			break
		}
	}
	if kantoLarge8h == nil {
		t.Error("関東・大型車・8時間制のデータが見つからない")
	} else if *kantoLarge8h != 60090 {
		t.Errorf("関東・大型車・8時間制基礎額: got %d, want 60090", *kantoLarge8h)
	}

	// 関東・小型車・4時間制の基礎額確認
	var kantoSmall4h *int
	for _, f := range baseFares {
		if f.RegionCode == 3 && f.VehicleCode == 1 && f.Hours == 4 {
			kantoSmall4h = &f.FareYen
			break
		}
	}
	if kantoSmall4h == nil {
		t.Error("関東・小型車・4時間制のデータが見つからない")
	} else if *kantoSmall4h != 23630 {
		t.Errorf("関東・小型車・4時間制基礎額: got %d, want 23630", *kantoSmall4h)
	}

	// 基礎走行キロの確認（8時間制・小型車は100km）
	for _, f := range baseFares {
		if f.Hours == 8 && f.VehicleCode == 1 && f.BaseKm != 100 {
			t.Errorf("8時間制・小型車の基礎走行キロ: got %d, want 100", f.BaseKm)
		}
		if f.Hours == 8 && f.VehicleCode != 1 && f.BaseKm != 130 {
			t.Errorf("8時間制・小型車以外の基礎走行キロ: got %d, want 130", f.BaseKm)
		}
		if f.Hours == 4 && f.VehicleCode == 1 && f.BaseKm != 50 {
			t.Errorf("4時間制・小型車の基礎走行キロ: got %d, want 50", f.BaseKm)
		}
		if f.Hours == 4 && f.VehicleCode != 1 && f.BaseKm != 60 {
			t.Errorf("4時間制・小型車以外の基礎走行キロ: got %d, want 60", f.BaseKm)
		}
	}
}

func TestSeedJtaTimeSurcharges(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// データ投入実行
	if err := seedJtaTimeFares(db); err != nil {
		t.Fatalf("JTA時間制運賃投入失敗: %v", err)
	}

	repo := repository.NewJtaTimeFareRepository(db)

	// 加算額の件数確認（10地域 × 4車格 × 2種別 = 80件）
	surcharges, err := repo.GetAllSurcharges()
	if err != nil {
		t.Fatalf("加算額取得失敗: %v", err)
	}
	if len(surcharges) != 80 {
		t.Errorf("加算額件数: got %d, want 80", len(surcharges))
	}

	// 関東・大型車・距離超過加算額確認（630円/10km）
	var kantoLargeDistance *int
	for _, s := range surcharges {
		if s.RegionCode == 3 && s.VehicleCode == 3 && s.SurchargeType == "distance" {
			kantoLargeDistance = &s.FareYen
			break
		}
	}
	if kantoLargeDistance == nil {
		t.Error("関東・大型車・距離超過のデータが見つからない")
	} else if *kantoLargeDistance != 630 {
		t.Errorf("関東・大型車・距離超過加算額: got %d, want 630", *kantoLargeDistance)
	}

	// 関東・大型車・時間超過加算額確認（4,180円/時間）
	var kantoLargeTime *int
	for _, s := range surcharges {
		if s.RegionCode == 3 && s.VehicleCode == 3 && s.SurchargeType == "time" {
			kantoLargeTime = &s.FareYen
			break
		}
	}
	if kantoLargeTime == nil {
		t.Error("関東・大型車・時間超過のデータが見つからない")
	} else if *kantoLargeTime != 4180 {
		t.Errorf("関東・大型車・時間超過加算額: got %d, want 4180", *kantoLargeTime)
	}
}

func TestSeedAkabouFares(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// データ投入実行
	if err := seedAkabouFares(db); err != nil {
		t.Fatalf("赤帽運賃投入失敗: %v", err)
	}

	repo := repository.NewAkabouFareRepository(db)

	// 距離制運賃の件数確認（5件）
	distanceFares, err := repo.GetAllDistanceFares()
	if err != nil {
		t.Fatalf("距離制運賃取得失敗: %v", err)
	}
	if len(distanceFares) != 5 {
		t.Errorf("距離制運賃件数: got %d, want 5", len(distanceFares))
	}

	// 20km以下の基本料金確認（5,500円）
	var base20km *int
	for _, f := range distanceFares {
		if f.MinKm == 0 && f.MaxKm != nil && *f.MaxKm == 20 {
			base20km = f.BaseFare
			break
		}
	}
	if base20km == nil {
		t.Error("20km以下の距離制運賃が見つからない")
	} else if *base20km != 5500 {
		t.Errorf("20km以下基本料金: got %d, want 5500", *base20km)
	}

	// 時間制運賃の件数確認（1件）
	timeFares, err := repo.GetAllTimeFares()
	if err != nil {
		t.Fatalf("時間制運賃取得失敗: %v", err)
	}
	if len(timeFares) != 1 {
		t.Errorf("時間制運賃件数: got %d, want 1", len(timeFares))
	}

	// 基本料金確認（6,050円）
	if len(timeFares) > 0 && timeFares[0].BaseFare != 6050 {
		t.Errorf("時間制基本料金: got %d, want 6050", timeFares[0].BaseFare)
	}

	// 割増料金の件数確認（2件: 休日・深夜）
	surcharges, err := repo.GetAllSurcharges()
	if err != nil {
		t.Fatalf("割増料金取得失敗: %v", err)
	}
	if len(surcharges) != 2 {
		t.Errorf("割増料金件数: got %d, want 2", len(surcharges))
	}

	// 地区割増の件数確認（2件: 東京23区・大阪市内）
	areaSurcharges, err := repo.GetAllAreaSurcharges()
	if err != nil {
		t.Fatalf("地区割増取得失敗: %v", err)
	}
	if len(areaSurcharges) != 2 {
		t.Errorf("地区割増件数: got %d, want 2", len(areaSurcharges))
	}

	// 付帯料金の件数確認（2件: 作業料・待機料）
	additionalFees, err := repo.GetAllAdditionalFees()
	if err != nil {
		t.Fatalf("付帯料金取得失敗: %v", err)
	}
	if len(additionalFees) != 2 {
		t.Errorf("付帯料金件数: got %d, want 2", len(additionalFees))
	}
}
