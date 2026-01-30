package main

import (
	"database/sql"
	"log"
	"path/filepath"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
)

func main() {
	dataDir := "data"
	dbPath := filepath.Join(dataDir, "str.db")

	db, err := database.InitMainDB(dbPath)
	if err != nil {
		log.Fatalf("DB接続エラー: %v", err)
	}
	defer db.Close()

	// JTA時間制運賃投入
	if err := seedJtaTimeFares(db); err != nil {
		log.Fatalf("JTA時間制運賃投入エラー: %v", err)
	}
	log.Println("JTA時間制運賃投入完了")

	// 赤帽運賃投入
	if err := seedAkabouFares(db); err != nil {
		log.Fatalf("赤帽運賃投入エラー: %v", err)
	}
	log.Println("赤帽運賃投入完了")

	log.Println("全マスタデータ投入完了")
}

// seedJtaTimeFares JTA時間制運賃を投入する
func seedJtaTimeFares(db *sql.DB) error {
	repo := repository.NewJtaTimeFareRepository(db)

	// 地域別・車格別・制度別の基礎額データ（令和6年3月告示）
	// region_code: 1=北海道, 2=東北, 3=関東, 4=北陸信越, 5=中部, 6=近畿, 7=中国, 8=四国, 9=九州, 10=沖縄
	// vehicle_code: 1=小型車(2t), 2=中型車(4t), 3=大型車(10t), 4=トレーラー(20t)

	// 8時間制基礎額 [region][vehicle] = fare_yen
	// 基礎走行キロ: 小型車100km、その他130km
	baseFares8h := map[int]map[int]int{
		1:  {1: 33250, 2: 39840, 3: 53240, 4: 68890}, // 北海道
		2:  {1: 33160, 2: 39880, 3: 52610, 4: 68410}, // 東北
		3:  {1: 39380, 2: 46640, 3: 60090, 4: 76840}, // 関東
		4:  {1: 34630, 2: 41160, 3: 54400, 4: 70020}, // 北陸信越
		5:  {1: 36390, 2: 43220, 3: 56490, 4: 73120}, // 中部
		6:  {1: 37640, 2: 43190, 3: 57690, 4: 73970}, // 近畿
		7:  {1: 34740, 2: 41760, 3: 55200, 4: 70780}, // 中国
		8:  {1: 33140, 2: 40640, 3: 53870, 4: 69470}, // 四国
		9:  {1: 33770, 2: 40740, 3: 53860, 4: 69790}, // 九州
		10: {1: 31310, 2: 37550, 3: 50420, 4: 66950}, // 沖縄
	}

	// 4時間制基礎額 [region][vehicle] = fare_yen
	// 基礎走行キロ: 小型車50km、その他60km
	baseFares4h := map[int]map[int]int{
		1:  {1: 19950, 2: 23900, 3: 31950, 4: 41330}, // 北海道
		2:  {1: 19900, 2: 23930, 3: 31570, 4: 41050}, // 東北
		3:  {1: 23630, 2: 27980, 3: 36050, 4: 46100}, // 関東
		4:  {1: 20780, 2: 24700, 3: 32640, 4: 42010}, // 北陸信越
		5:  {1: 21830, 2: 25940, 3: 33890, 4: 43870}, // 中部
		6:  {1: 22580, 2: 25060, 3: 34610, 4: 44380}, // 近畿
		7:  {1: 19880, 2: 25060, 3: 33120, 4: 42450}, // 中国
		8:  {1: 19880, 2: 24380, 3: 32320, 4: 41680}, // 四国
		9:  {1: 20260, 2: 24440, 3: 32320, 4: 41820}, // 九州
		10: {1: 18790, 2: 22530, 3: 30250, 4: 39830}, // 沖縄
	}

	// 距離超過加算額 [region][vehicle] = fare_yen (10kmごと)
	distanceSurcharges := map[int]map[int]int{
		1:  {1: 350, 2: 410, 3: 630, 4: 920}, // 北海道
		2:  {1: 340, 2: 410, 3: 630, 4: 920}, // 東北
		3:  {1: 410, 2: 410, 3: 630, 4: 920}, // 関東
		4:  {1: 350, 2: 410, 3: 630, 4: 920}, // 北陸信越
		5:  {1: 340, 2: 410, 3: 630, 4: 920}, // 中部
		6:  {1: 410, 2: 410, 3: 630, 4: 920}, // 近畿
		7:  {1: 340, 2: 410, 3: 630, 4: 920}, // 中国
		8:  {1: 340, 2: 410, 3: 630, 4: 920}, // 四国
		9:  {1: 340, 2: 410, 3: 630, 4: 920}, // 九州
		10: {1: 340, 2: 410, 3: 630, 4: 920}, // 沖縄
	}

	// 時間超過加算額 [region][vehicle] = fare_yen (1時間ごと)
	timeSurcharges := map[int]map[int]int{
		1:  {1: 2790, 2: 2930, 3: 3150, 4: 3700}, // 北海道
		2:  {1: 2780, 2: 2910, 3: 3130, 4: 3680}, // 東北
		3:  {1: 3710, 2: 3890, 3: 4180, 4: 4920}, // 関東
		4:  {1: 3010, 2: 3110, 3: 3380, 4: 3970}, // 北陸信越
		5:  {1: 3130, 2: 3190, 3: 3490, 4: 4150}, // 中部
		6:  {1: 3430, 2: 3090, 3: 3470, 4: 4050}, // 近畿
		7:  {1: 3060, 2: 3210, 3: 3450, 4: 4060}, // 中国
		8:  {1: 2890, 2: 3030, 3: 3260, 4: 3830}, // 四国
		9:  {1: 2940, 2: 3090, 3: 3320, 4: 3900}, // 九州
		10: {1: 2550, 2: 2680, 3: 2920, 4: 3380}, // 沖縄
	}

	// 基礎額投入
	for regionCode := 1; regionCode <= 10; regionCode++ {
		for vehicleCode := 1; vehicleCode <= 4; vehicleCode++ {
			// 8時間制
			baseKm8h := 130
			if vehicleCode == 1 {
				baseKm8h = 100
			}
			fare8h := &model.JtaTimeBaseFare{
				RegionCode:  regionCode,
				VehicleCode: vehicleCode,
				Hours:       8,
				BaseKm:      baseKm8h,
				FareYen:     baseFares8h[regionCode][vehicleCode],
			}
			if _, err := repo.CreateBaseFare(fare8h); err != nil {
				return err
			}

			// 4時間制
			baseKm4h := 60
			if vehicleCode == 1 {
				baseKm4h = 50
			}
			fare4h := &model.JtaTimeBaseFare{
				RegionCode:  regionCode,
				VehicleCode: vehicleCode,
				Hours:       4,
				BaseKm:      baseKm4h,
				FareYen:     baseFares4h[regionCode][vehicleCode],
			}
			if _, err := repo.CreateBaseFare(fare4h); err != nil {
				return err
			}
		}
	}

	// 加算額投入
	for regionCode := 1; regionCode <= 10; regionCode++ {
		for vehicleCode := 1; vehicleCode <= 4; vehicleCode++ {
			// 距離超過
			distSurcharge := &model.JtaTimeSurcharge{
				RegionCode:    regionCode,
				VehicleCode:   vehicleCode,
				SurchargeType: "distance",
				FareYen:       distanceSurcharges[regionCode][vehicleCode],
			}
			if _, err := repo.CreateSurcharge(distSurcharge); err != nil {
				return err
			}

			// 時間超過
			timeSurcharge := &model.JtaTimeSurcharge{
				RegionCode:    regionCode,
				VehicleCode:   vehicleCode,
				SurchargeType: "time",
				FareYen:       timeSurcharges[regionCode][vehicleCode],
			}
			if _, err := repo.CreateSurcharge(timeSurcharge); err != nil {
				return err
			}
		}
	}

	return nil
}

// seedAkabouFares 赤帽運賃を投入する
func seedAkabouFares(db *sql.DB) error {
	repo := repository.NewAkabouFareRepository(db)

	// 距離制運賃（税込）
	distanceFares := []struct {
		minKm     int
		maxKm     *int
		baseFare  *int
		perKmRate *int
	}{
		{0, intPtr(20), intPtr(5500), nil},       // 20km迄: 5,500円
		{21, intPtr(50), nil, intPtr(242)},       // 21〜50km: +242円/km
		{51, intPtr(100), nil, intPtr(187)},      // 51〜100km: +187円/km
		{101, intPtr(150), nil, intPtr(154)},     // 101〜150km: +154円/km
		{151, nil, nil, intPtr(132)},             // 151km以上: +132円/km
	}
	for _, f := range distanceFares {
		fare := &model.AkabouDistanceFare{
			MinKm:     f.minKm,
			MaxKm:     f.maxKm,
			BaseFare:  f.baseFare,
			PerKmRate: f.perKmRate,
		}
		if _, err := repo.CreateDistanceFare(fare); err != nil {
			return err
		}
	}

	// 時間制運賃（税込）
	// 基本（2時間・20km迄）: 6,050円、超過30分ごと: +1,375円
	timeFare := &model.AkabouTimeFare{
		BaseHours:    2,
		BaseKm:       20,
		BaseFare:     6050,
		OvertimeRate: 1375, // 30分ごと
	}
	if _, err := repo.CreateTimeFare(timeFare); err != nil {
		return err
	}

	// 割増料金
	surcharges := []struct {
		surchargeType string
		ratePercent   int
		description   string
	}{
		{"holiday", 20, "休日（日祝）"},
		{"night", 30, "深夜・早朝（22:00〜5:00）"},
	}
	for _, s := range surcharges {
		desc := s.description
		surcharge := &model.AkabouSurcharge{
			SurchargeType: s.surchargeType,
			RatePercent:   s.ratePercent,
			Description:   &desc,
		}
		if _, err := repo.CreateSurcharge(surcharge); err != nil {
			return err
		}
	}

	// 地区割増（税込）
	areaSurcharges := []struct {
		areaName        string
		surchargeAmount int
	}{
		{"東京23区", 440},
		{"大阪市内", 440},
	}
	for _, a := range areaSurcharges {
		area := &model.AkabouAreaSurcharge{
			AreaName:        a.areaName,
			SurchargeAmount: a.surchargeAmount,
		}
		if _, err := repo.CreateAreaSurcharge(area); err != nil {
			return err
		}
	}

	// 付帯料金（税込）
	additionalFees := []struct {
		feeType     string
		freeMinutes int
		unitMinutes int
		feeAmount   int
	}{
		{"work", 30, 15, 550},     // 作業料金: 30分超過で15分ごと550円
		{"waiting", 30, 30, 1100}, // 待機時間: 30分超過で30分ごと1,100円
	}
	for _, f := range additionalFees {
		fee := &model.AkabouAdditionalFee{
			FeeType:     f.feeType,
			FreeMinutes: f.freeMinutes,
			UnitMinutes: f.unitMinutes,
			FeeAmount:   f.feeAmount,
		}
		if _, err := repo.CreateAdditionalFee(fee); err != nil {
			return err
		}
	}

	return nil
}

func intPtr(i int) *int {
	return &i
}
