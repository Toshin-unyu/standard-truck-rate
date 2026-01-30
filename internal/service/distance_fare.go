package service

import (
	"fmt"
)

// 割増率の定数
const (
	NightSurchargeRate   = 1.3 // 深夜割増: 3割増（22:00-05:00）
	HolidaySurchargeRate = 1.2 // 休日割増: 2割増（日祝日）
)

// FareGetter 運賃取得インターフェース（テスト用にモック可能）
type FareGetter interface {
	GetDistanceFareYen(regionCode, vehicleCode, distanceKm int) (int, error)
}

// JtaSupabaseClientAdapter JtaSupabaseClientをFareGetterに適合させるアダプター
type JtaSupabaseClientAdapter struct {
	client *JtaSupabaseClient
}

// NewJtaSupabaseClientAdapter アダプターを作成
func NewJtaSupabaseClientAdapter(client *JtaSupabaseClient) *JtaSupabaseClientAdapter {
	return &JtaSupabaseClientAdapter{client: client}
}

// GetDistanceFareYen 運賃を取得して金額のみ返す
func (a *JtaSupabaseClientAdapter) GetDistanceFareYen(regionCode, vehicleCode, distanceKm int) (int, error) {
	fare, err := a.client.GetDistanceFare(regionCode, vehicleCode, distanceKm)
	if err != nil {
		return 0, err
	}
	return fare.FareYen, nil
}

// DistanceFareService 距離制運賃計算サービス
type DistanceFareService struct {
	fareGetter FareGetter
}

// NewDistanceFareService 新しいDistanceFareServiceを作成
func NewDistanceFareService(fareGetter FareGetter) *DistanceFareService {
	return &DistanceFareService{
		fareGetter: fareGetter,
	}
}

// DistanceFareResult 距離制運賃計算結果
type DistanceFareResult struct {
	// 入力パラメータ
	RegionCode  int // 運輸局コード
	VehicleCode int // 車格コード
	DistanceKm  int // 入力距離 (km)
	RoundedKm   int // 丸め後距離 (km)

	// 運賃計算結果
	BaseFare          int // 基本運賃（円）
	NightSurcharge    int // 深夜割増額（円）
	HolidaySurcharge  int // 休日割増額（円）
	TotalFare         int // 合計運賃（円）

	// 割増率
	NightRate   float64 // 深夜割増率（1.0 or 1.3）
	HolidayRate float64 // 休日割増率（1.0 or 1.2）

	// フラグ
	IsNight   bool // 深夜適用
	IsHoliday bool // 休日適用
}

// Calculate 距離制運賃を計算する
func (s *DistanceFareService) Calculate(
	regionCode, vehicleCode, distanceKm int,
	isNight, isHoliday bool,
) (*DistanceFareResult, error) {
	// 入力値検証
	if err := s.validateInput(regionCode, vehicleCode, distanceKm); err != nil {
		return nil, err
	}

	// 距離を丸める
	roundedKm := RoundDistance(distanceKm, regionCode)

	// 基本運賃を取得
	baseFare, err := s.fareGetter.GetDistanceFareYen(regionCode, vehicleCode, roundedKm)
	if err != nil {
		return nil, fmt.Errorf("運賃取得エラー: %w", err)
	}

	// 割増計算
	nightRate := 1.0
	holidayRate := 1.0
	nightSurcharge := 0
	holidaySurcharge := 0
	totalFare := baseFare

	// 深夜割増（3割増）
	if isNight {
		nightRate = NightSurchargeRate
		nightSurcharge = int(float64(baseFare) * (NightSurchargeRate - 1.0))
		totalFare += nightSurcharge
	}

	// 休日割増（2割増）- 深夜割増後の金額に適用
	if isHoliday {
		holidayRate = HolidaySurchargeRate
		holidaySurcharge = int(float64(totalFare) * (HolidaySurchargeRate - 1.0))
		totalFare += holidaySurcharge
	}

	return &DistanceFareResult{
		RegionCode:       regionCode,
		VehicleCode:      vehicleCode,
		DistanceKm:       distanceKm,
		RoundedKm:        roundedKm,
		BaseFare:         baseFare,
		NightSurcharge:   nightSurcharge,
		HolidaySurcharge: holidaySurcharge,
		TotalFare:        totalFare,
		NightRate:        nightRate,
		HolidayRate:      holidayRate,
		IsNight:          isNight,
		IsHoliday:        isHoliday,
	}, nil
}

// validateInput 入力値を検証
func (s *DistanceFareService) validateInput(regionCode, vehicleCode, distanceKm int) error {
	// 運輸局コード: 1-10
	if regionCode < 1 || regionCode > 10 {
		return fmt.Errorf("無効な運輸局コード: %d（1-10の範囲で指定）", regionCode)
	}

	// 車格コード: 1-4
	if vehicleCode < 1 || vehicleCode > 4 {
		return fmt.Errorf("無効な車格コード: %d（1-4の範囲で指定）", vehicleCode)
	}

	// 距離: 1以上
	if distanceKm < 1 {
		return fmt.Errorf("無効な距離: %d（1km以上を指定）", distanceKm)
	}

	return nil
}

// Breakdown 計算根拠を文字列で返す
func (r *DistanceFareResult) Breakdown() string {
	regionNames := map[int]string{
		1: "北海道", 2: "東北", 3: "関東", 4: "北陸信越", 5: "中部",
		6: "近畿", 7: "中国", 8: "四国", 9: "九州", 10: "沖縄",
	}
	vehicleNames := map[int]string{
		1: "小型車(2t)", 2: "中型車(4t)", 3: "大型車(10t)", 4: "トレーラー(20t)",
	}

	result := fmt.Sprintf("【計算根拠】\n")
	result += fmt.Sprintf("  運輸局: %s\n", regionNames[r.RegionCode])
	result += fmt.Sprintf("  車格: %s\n", vehicleNames[r.VehicleCode])
	result += fmt.Sprintf("  経路距離: %dkm → 運賃計算距離: %dkm\n", r.DistanceKm, r.RoundedKm)
	result += fmt.Sprintf("  基本運賃: %d円\n", r.BaseFare)

	if r.IsNight {
		result += fmt.Sprintf("  深夜割増: +%d円（%.0f%%増）\n", r.NightSurcharge, (r.NightRate-1.0)*100)
	}
	if r.IsHoliday {
		result += fmt.Sprintf("  休日割増: +%d円（%.0f%%増）\n", r.HolidaySurcharge, (r.HolidayRate-1.0)*100)
	}

	result += fmt.Sprintf("  合計運賃: %d円\n", r.TotalFare)

	return result
}
