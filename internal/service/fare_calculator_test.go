package service

import (
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// MockTimeFareGetter テスト用モック
type MockTimeFareGetter struct{}

func (m *MockTimeFareGetter) GetBaseFare(regionCode, vehicleCode, hours int) (*model.JtaTimeBaseFare, error) {
	// 関東・大型車・8時間制の場合
	if regionCode == 3 && vehicleCode == 3 && hours == 8 {
		return &model.JtaTimeBaseFare{
			RegionCode:  3,
			VehicleCode: 3,
			Hours:       8,
			BaseKm:      130,
			FareYen:     60090,
		}, nil
	}
	// 関東・大型車・4時間制の場合
	if regionCode == 3 && vehicleCode == 3 && hours == 4 {
		return &model.JtaTimeBaseFare{
			RegionCode:  3,
			VehicleCode: 3,
			Hours:       4,
			BaseKm:      60,
			FareYen:     36050,
		}, nil
	}
	return &model.JtaTimeBaseFare{
		RegionCode:  regionCode,
		VehicleCode: vehicleCode,
		Hours:       hours,
		BaseKm:      100,
		FareYen:     40000,
	}, nil
}

func (m *MockTimeFareGetter) GetSurcharge(regionCode, vehicleCode int, surchargeType string) (*model.JtaTimeSurcharge, error) {
	if surchargeType == "distance" {
		return &model.JtaTimeSurcharge{
			RegionCode:    regionCode,
			VehicleCode:   vehicleCode,
			SurchargeType: "distance",
			FareYen:       630, // 大型車の10kmあたり加算額
		}, nil
	}
	return &model.JtaTimeSurcharge{
		RegionCode:    regionCode,
		VehicleCode:   vehicleCode,
		SurchargeType: "time",
		FareYen:       4180, // 大型車の1時間あたり加算額
	}, nil
}

// MockFareGetter 距離制運賃用モック
type MockFareGetter struct{}

func (m *MockFareGetter) GetDistanceFareYen(regionCode, vehicleCode, distanceKm int) (int, error) {
	// 関東・大型車・100kmの場合
	if regionCode == 3 && vehicleCode == 3 && distanceKm == 100 {
		return 35000, nil
	}
	// 関東・大型車・550kmの場合（requirements.mdの例）
	if regionCode == 3 && vehicleCode == 3 && distanceKm == 550 {
		return 182480, nil
	}
	return 50000, nil
}

// TestFareCalculatorService_CalculateAll_Truck トラック（2t以上）の運賃一括計算テスト
func TestFareCalculatorService_CalculateAll_Truck(t *testing.T) {
	// サービスを作成
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	// テストケース: 関東・大型車・100km・走行2時間・荷役1時間
	req := &FareCalculationRequest{
		RegionCode:      3,  // 関東
		VehicleCode:     3,  // 大型車（2t以上なのでトラ協運賃のみ）
		DistanceKm:      100,
		DrivingMinutes:  120, // 2時間
		LoadingMinutes:  60,  // 1時間
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false, // トラ協PDF版
		Area:            "",
	}

	result, err := calculator.CalculateAll(req)
	if err != nil {
		t.Fatalf("CalculateAll failed: %v", err)
	}

	// 距離制運賃が計算されていること
	if result.DistanceFareResult == nil {
		t.Error("DistanceFareResult should not be nil for truck")
	} else {
		if result.DistanceFareResult.TotalFare <= 0 {
			t.Error("DistanceFare TotalFare should be > 0")
		}
	}

	// 時間制運賃が計算されていること
	if result.TimeFareResult == nil {
		t.Error("TimeFareResult should not be nil for truck")
	} else {
		if result.TimeFareResult.TotalFare <= 0 {
			t.Error("TimeFare TotalFare should be > 0")
		}
	}

	// トラックの場合、赤帽運賃はnilであること
	if result.AkabouDistanceResult != nil {
		t.Error("AkabouDistanceResult should be nil for truck")
	}
	if result.AkabouTimeResult != nil {
		t.Error("AkabouTimeResult should be nil for truck")
	}

	// ランキングが2件（距離制・時間制のみ）であること
	if len(result.Rankings) != 2 {
		t.Errorf("Rankings should have 2 items for truck, got %d", len(result.Rankings))
	}

	// 最安値が設定されていること
	if result.CheapestType == "" {
		t.Error("CheapestType should not be empty")
	}
	if result.CheapestFare <= 0 {
		t.Error("CheapestFare should be > 0")
	}
}

// TestFareCalculatorService_CalculateAll_Light 軽貨物の運賃一括計算テスト
func TestFareCalculatorService_CalculateAll_Light(t *testing.T) {
	// サービスを作成
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	// テストケース: 軽貨物・100km・走行2時間・荷役1時間
	req := &FareCalculationRequest{
		RegionCode:      3,  // 関東
		VehicleCode:     0,  // 軽貨物（赤帽運賃のみ）
		DistanceKm:      100,
		DrivingMinutes:  120, // 2時間
		LoadingMinutes:  60,  // 1時間
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	result, err := calculator.CalculateAll(req)
	if err != nil {
		t.Fatalf("CalculateAll failed: %v", err)
	}

	// 軽貨物の場合、トラ協運賃はnilであること
	if result.DistanceFareResult != nil {
		t.Error("DistanceFareResult should be nil for light vehicle")
	}
	if result.TimeFareResult != nil {
		t.Error("TimeFareResult should be nil for light vehicle")
	}

	// 赤帽運賃（距離制）が計算されていること
	if result.AkabouDistanceResult == nil {
		t.Error("AkabouDistanceResult should not be nil for light vehicle")
	} else {
		if result.AkabouDistanceResult.TotalFare <= 0 {
			t.Error("AkabouDistanceFare TotalFare should be > 0")
		}
	}

	// 赤帽運賃（時間制）が計算されていること
	if result.AkabouTimeResult == nil {
		t.Error("AkabouTimeResult should not be nil for light vehicle")
	} else {
		if result.AkabouTimeResult.TotalFare <= 0 {
			t.Error("AkabouTimeFare TotalFare should be > 0")
		}
	}

	// ランキングが2件（赤帽距離制・時間制のみ）であること
	if len(result.Rankings) != 2 {
		t.Errorf("Rankings should have 2 items for light vehicle, got %d", len(result.Rankings))
	}

	// 最安値が設定されていること
	if result.CheapestType == "" {
		t.Error("CheapestType should not be empty")
	}
	if result.CheapestFare <= 0 {
		t.Error("CheapestFare should be > 0")
	}
}

// TestFareCalculatorService_Rankings ランキング順序テスト
func TestFareCalculatorService_Rankings(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	req := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     3,
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	result, err := calculator.CalculateAll(req)
	if err != nil {
		t.Fatalf("CalculateAll failed: %v", err)
	}

	// ランキングが金額順（昇順）にソートされていること
	for i := 1; i < len(result.Rankings); i++ {
		if result.Rankings[i-1].Fare > result.Rankings[i].Fare {
			t.Errorf("Rankings should be sorted by fare ascending: %d > %d",
				result.Rankings[i-1].Fare, result.Rankings[i].Fare)
		}
	}

	// 最安値がランキングの先頭と一致すること
	if len(result.Rankings) > 0 {
		if result.CheapestFare != result.Rankings[0].Fare {
			t.Errorf("CheapestFare should match first ranking: %d != %d",
				result.CheapestFare, result.Rankings[0].Fare)
		}
		if result.CheapestType != result.Rankings[0].Type {
			t.Errorf("CheapestType should match first ranking: %s != %s",
				result.CheapestType, result.Rankings[0].Type)
		}
	}
}

// TestFareCalculatorService_WithSurcharges_Truck トラックの割増適用テスト
func TestFareCalculatorService_WithSurcharges_Truck(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	// 割増なし
	reqNoSurcharge := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     3, // 大型車
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	// 深夜・休日割増あり
	reqWithSurcharge := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     3,
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         true,
		IsHoliday:       true,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	resultNo, err := calculator.CalculateAll(reqNoSurcharge)
	if err != nil {
		t.Fatalf("CalculateAll (no surcharge) failed: %v", err)
	}

	resultWith, err := calculator.CalculateAll(reqWithSurcharge)
	if err != nil {
		t.Fatalf("CalculateAll (with surcharge) failed: %v", err)
	}

	// 割増ありの方が高いこと（トラ協運賃のみ）
	if resultWith.DistanceFareResult.TotalFare <= resultNo.DistanceFareResult.TotalFare {
		t.Error("Distance fare with surcharge should be higher")
	}
	if resultWith.TimeFareResult.TotalFare <= resultNo.TimeFareResult.TotalFare {
		t.Error("Time fare with surcharge should be higher")
	}
}

// TestFareCalculatorService_WithSurcharges_Light 軽貨物の割増適用テスト
func TestFareCalculatorService_WithSurcharges_Light(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	// 割増なし
	reqNoSurcharge := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     0, // 軽貨物
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	// 深夜・休日割増あり
	reqWithSurcharge := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     0,
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         true,
		IsHoliday:       true,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	resultNo, err := calculator.CalculateAll(reqNoSurcharge)
	if err != nil {
		t.Fatalf("CalculateAll (no surcharge) failed: %v", err)
	}

	resultWith, err := calculator.CalculateAll(reqWithSurcharge)
	if err != nil {
		t.Fatalf("CalculateAll (with surcharge) failed: %v", err)
	}

	// 割増ありの方が高いこと（赤帽運賃のみ）
	if resultWith.AkabouDistanceResult.TotalFare <= resultNo.AkabouDistanceResult.TotalFare {
		t.Error("Akabou distance fare with surcharge should be higher")
	}
	if resultWith.AkabouTimeResult.TotalFare <= resultNo.AkabouTimeResult.TotalFare {
		t.Error("Akabou time fare with surcharge should be higher")
	}
}

// TestFareCalculatorService_Breakdown_Truck トラックの計算根拠テスト
func TestFareCalculatorService_Breakdown_Truck(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	req := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     3, // 大型車
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	result, err := calculator.CalculateAll(req)
	if err != nil {
		t.Fatalf("CalculateAll failed: %v", err)
	}

	// 計算根拠が生成されること
	breakdown := result.Breakdown()
	if breakdown == "" {
		t.Error("Breakdown should not be empty")
	}

	// トラ協運賃の情報が含まれていること
	if !containsString(breakdown, "距離制") {
		t.Error("Breakdown should contain 距離制")
	}
	if !containsString(breakdown, "時間制") {
		t.Error("Breakdown should contain 時間制")
	}
	// トラックの場合、赤帽は含まれないこと
	if containsString(breakdown, "赤帽") {
		t.Error("Breakdown should not contain 赤帽 for truck")
	}
}

// TestFareCalculatorService_Breakdown_Light 軽貨物の計算根拠テスト
func TestFareCalculatorService_Breakdown_Light(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	req := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     0, // 軽貨物
		DistanceKm:      100,
		DrivingMinutes:  120,
		LoadingMinutes:  60,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	result, err := calculator.CalculateAll(req)
	if err != nil {
		t.Fatalf("CalculateAll failed: %v", err)
	}

	// 計算根拠が生成されること
	breakdown := result.Breakdown()
	if breakdown == "" {
		t.Error("Breakdown should not be empty")
	}

	// 赤帽運賃の情報が含まれていること
	if !containsString(breakdown, "赤帽") {
		t.Error("Breakdown should contain 赤帽 for light vehicle")
	}
}

// TestFareCalculatorService_AreaSurcharge 地区割増テスト（軽貨物）
func TestFareCalculatorService_AreaSurcharge(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	// 地区割増なし
	reqNoArea := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     0, // 軽貨物（地区割増は赤帽のみ適用）
		DistanceKm:      50,
		DrivingMinutes:  60,
		LoadingMinutes:  30,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "",
	}

	// 東京23区（地区割増あり）
	reqTokyo := &FareCalculationRequest{
		RegionCode:      3,
		VehicleCode:     0, // 軽貨物
		DistanceKm:      50,
		DrivingMinutes:  60,
		LoadingMinutes:  30,
		IsNight:         false,
		IsHoliday:       false,
		UseSimpleBaseKm: false,
		Area:            "東京23区",
	}

	resultNo, err := calculator.CalculateAll(reqNoArea)
	if err != nil {
		t.Fatalf("CalculateAll (no area) failed: %v", err)
	}

	resultTokyo, err := calculator.CalculateAll(reqTokyo)
	if err != nil {
		t.Fatalf("CalculateAll (Tokyo) failed: %v", err)
	}

	// 東京23区の方が赤帽運賃が高いこと（地区割増440円）
	if resultTokyo.AkabouDistanceResult.TotalFare <= resultNo.AkabouDistanceResult.TotalFare {
		t.Error("Akabou fare in Tokyo should be higher due to area surcharge")
	}
	if resultTokyo.AkabouDistanceResult.AreaSurcharge != 440 {
		t.Errorf("Area surcharge should be 440, got %d", resultTokyo.AkabouDistanceResult.AreaSurcharge)
	}
}

// TestFareCalculatorService_DistanceKmRaw 小数点距離と荷役時間の伝播テスト
func TestFareCalculatorService_DistanceKmRaw(t *testing.T) {
	distanceFareService := NewDistanceFareService(&MockFareGetter{})
	timeFareService := NewTimeFareService(&MockTimeFareGetter{})
	akabouFareService := NewAkabouFareService()

	calculator := NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)

	// トラック: DistanceKmRaw と LoadingMinutes が結果に伝播されること
	req := &FareCalculationRequest{
		RegionCode:     3,
		VehicleCode:    3,
		DistanceKm:     505,
		DistanceKmRaw:  504.6,
		DrivingMinutes: 360,
		LoadingMinutes: 60,
		IsNight:        false,
		IsHoliday:      false,
	}

	result, err := calculator.CalculateAll(req)
	if err != nil {
		t.Fatalf("CalculateAll failed: %v", err)
	}

	if result.DistanceKmRaw != 504.6 {
		t.Errorf("DistanceKmRaw should be 504.6, got %f", result.DistanceKmRaw)
	}
	if result.LoadingMinutes != 60 {
		t.Errorf("LoadingMinutes should be 60, got %d", result.LoadingMinutes)
	}

	// 軽貨物でも同様に伝播されること
	reqLight := &FareCalculationRequest{
		RegionCode:     3,
		VehicleCode:    0,
		DistanceKm:     100,
		DistanceKmRaw:  99.8,
		DrivingMinutes: 120,
		LoadingMinutes: 45,
		IsNight:        false,
		IsHoliday:      false,
	}

	resultLight, err := calculator.CalculateAll(reqLight)
	if err != nil {
		t.Fatalf("CalculateAll (light) failed: %v", err)
	}

	if resultLight.DistanceKmRaw != 99.8 {
		t.Errorf("DistanceKmRaw should be 99.8, got %f", resultLight.DistanceKmRaw)
	}
	if resultLight.LoadingMinutes != 45 {
		t.Errorf("LoadingMinutes should be 45, got %d", resultLight.LoadingMinutes)
	}
}

// containsString 文字列に部分文字列が含まれるか
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
