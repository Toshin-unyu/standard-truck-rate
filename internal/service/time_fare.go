package service

import (
	"fmt"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// TimeFareGetter 時間制運賃取得インターフェース（テスト用にモック可能）
type TimeFareGetter interface {
	GetBaseFare(regionCode, vehicleCode, hours int) (*model.JtaTimeBaseFare, error)
	GetSurcharge(regionCode, vehicleCode int, surchargeType string) (*model.JtaTimeSurcharge, error)
}

// TimeFareService 時間制運賃計算サービス
type TimeFareService struct {
	fareGetter TimeFareGetter
}

// NewTimeFareService 新しいTimeFareServiceを作成
func NewTimeFareService(fareGetter TimeFareGetter) *TimeFareService {
	return &TimeFareService{
		fareGetter: fareGetter,
	}
}

// TimeFareResult 時間制運賃計算結果
type TimeFareResult struct {
	// 入力パラメータ
	RegionCode     int // 運輸局コード
	VehicleCode    int // 車格コード
	DistanceKm     int // 走行距離 (km)
	DrivingMinutes int // 走行時間（分）
	LoadingMinutes int // 荷役時間（分）
	TotalMinutes   int // 総作業時間（分）

	// 適用制度
	AppliedHours int // 適用時間制（4 or 8）
	BaseKm       int // 基礎走行キロ

	// 超過計算
	ExcessKm      int // 超過距離 (km)
	ExcessMinutes int // 超過時間（分）
	ExcessHours   int // 超過時間（時間、切り上げ）

	// 運賃計算結果
	BaseFare           int // 基礎額（円）
	DistanceSurcharge  int // 距離超過加算額（円）
	TimeSurcharge      int // 時間超過加算額（円）
	SubTotal           int // 小計（割増前）
	NightSurcharge     int // 深夜割増額（円）
	HolidaySurcharge   int // 休日割増額（円）
	TotalFare          int // 合計運賃（円）

	// 割増率
	NightRate   float64 // 深夜割増率（1.0 or 1.3）
	HolidayRate float64 // 休日割増率（1.0 or 1.2）

	// フラグ
	IsNight         bool // 深夜適用
	IsHoliday       bool // 休日適用
	UseSimpleBaseKm bool // シンプル版基礎走行キロ使用
}

// DetermineHoursSystem 総作業時間から適用時間制を判定
// 4時間以内 → 4時間制、それ以外 → 8時間制
func DetermineHoursSystem(totalMinutes int) int {
	if totalMinutes <= 240 { // 4時間 = 240分
		return 4
	}
	return 8
}

// シンプル版の基礎走行キロ（Issue記載値）
const (
	SimpleBaseKm4Hours = 30 // 4時間制: 30km固定
	SimpleBaseKm8Hours = 50 // 8時間制: 50km固定
)

// Calculate 時間制運賃を計算する
// useSimpleBaseKm: trueの場合、シンプル版の基礎走行キロ（30km/50km固定）を使用
//
//	falseの場合、トラ協PDF版の基礎走行キロ（車格別）を使用
func (s *TimeFareService) Calculate(
	regionCode, vehicleCode, distanceKm int,
	drivingMinutes, loadingMinutes int,
	isNight, isHoliday bool,
	useSimpleBaseKm bool,
) (*TimeFareResult, error) {
	// 入力値検証
	if err := s.validateInput(regionCode, vehicleCode, distanceKm, drivingMinutes); err != nil {
		return nil, err
	}

	// 総作業時間を計算
	totalMinutes := drivingMinutes + loadingMinutes

	// 適用時間制を判定
	appliedHours := DetermineHoursSystem(totalMinutes)

	// 基礎額を取得
	baseFare, err := s.fareGetter.GetBaseFare(regionCode, vehicleCode, appliedHours)
	if err != nil {
		return nil, fmt.Errorf("基礎額取得エラー: %w", err)
	}

	// 距離超過加算額を取得
	distSurcharge, err := s.fareGetter.GetSurcharge(regionCode, vehicleCode, "distance")
	if err != nil {
		return nil, fmt.Errorf("距離超過加算額取得エラー: %w", err)
	}

	// 時間超過加算額を取得
	timeSurcharge, err := s.fareGetter.GetSurcharge(regionCode, vehicleCode, "time")
	if err != nil {
		return nil, fmt.Errorf("時間超過加算額取得エラー: %w", err)
	}

	// 基礎走行キロを決定
	var baseKm int
	if useSimpleBaseKm {
		// シンプル版: 車格に関係なく固定値
		if appliedHours == 4 {
			baseKm = SimpleBaseKm4Hours
		} else {
			baseKm = SimpleBaseKm8Hours
		}
	} else {
		// トラ協PDF版: DBから取得した車格別の値
		baseKm = baseFare.BaseKm
	}

	// 超過距離を計算（10km単位）
	excessKm := 0
	if distanceKm > baseKm {
		excessKm = distanceKm - baseKm
	}
	distanceSurchargeAmount := (excessKm / 10) * distSurcharge.FareYen

	// 超過時間を計算（1時間単位、切り上げ）
	baseMinutes := appliedHours * 60
	excessMinutes := 0
	if totalMinutes > baseMinutes {
		excessMinutes = totalMinutes - baseMinutes
	}
	excessHours := (excessMinutes + 59) / 60 // 切り上げ
	timeSurchargeAmount := excessHours * timeSurcharge.FareYen

	// 小計（割増前）
	subTotal := baseFare.FareYen + distanceSurchargeAmount + timeSurchargeAmount

	// 割増計算
	nightRate := 1.0
	holidayRate := 1.0
	nightSurchargeAmount := 0
	holidaySurchargeAmount := 0
	totalFare := subTotal

	// 深夜割増（3割増）
	if isNight {
		nightRate = NightSurchargeRate
		nightSurchargeAmount = int(float64(subTotal) * (NightSurchargeRate - 1.0))
		totalFare += nightSurchargeAmount
	}

	// 休日割増（2割増）- 深夜割増後の金額に適用
	if isHoliday {
		holidayRate = HolidaySurchargeRate
		holidaySurchargeAmount = int(float64(totalFare) * (HolidaySurchargeRate - 1.0))
		totalFare += holidaySurchargeAmount
	}

	return &TimeFareResult{
		RegionCode:        regionCode,
		VehicleCode:       vehicleCode,
		DistanceKm:        distanceKm,
		DrivingMinutes:    drivingMinutes,
		LoadingMinutes:    loadingMinutes,
		TotalMinutes:      totalMinutes,
		AppliedHours:      appliedHours,
		BaseKm:            baseKm,
		ExcessKm:          excessKm,
		ExcessMinutes:     excessMinutes,
		ExcessHours:       excessHours,
		BaseFare:          baseFare.FareYen,
		DistanceSurcharge: distanceSurchargeAmount,
		TimeSurcharge:     timeSurchargeAmount,
		SubTotal:          subTotal,
		NightSurcharge:    nightSurchargeAmount,
		HolidaySurcharge:  holidaySurchargeAmount,
		TotalFare:         totalFare,
		NightRate:         nightRate,
		HolidayRate:       holidayRate,
		IsNight:           isNight,
		IsHoliday:         isHoliday,
		UseSimpleBaseKm:   useSimpleBaseKm,
	}, nil
}

// validateInput 入力値を検証
func (s *TimeFareService) validateInput(regionCode, vehicleCode, distanceKm, drivingMinutes int) error {
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

	// 走行時間: 1以上
	if drivingMinutes < 1 {
		return fmt.Errorf("無効な走行時間: %d（1分以上を指定）", drivingMinutes)
	}

	return nil
}

// Breakdown 計算根拠を文字列で返す
func (r *TimeFareResult) Breakdown() string {
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
	result += fmt.Sprintf("  適用制度: %d時間制\n", r.AppliedHours)

	baseKmType := "トラ協PDF版"
	if r.UseSimpleBaseKm {
		baseKmType = "シンプル版"
	}
	result += fmt.Sprintf("  走行距離: %dkm（基礎走行キロ: %dkm [%s]）\n", r.DistanceKm, r.BaseKm, baseKmType)
	result += fmt.Sprintf("  走行時間: %d時間%d分\n", r.DrivingMinutes/60, r.DrivingMinutes%60)
	result += fmt.Sprintf("  荷役時間: %d時間%d分\n", r.LoadingMinutes/60, r.LoadingMinutes%60)
	result += fmt.Sprintf("  総作業時間: %d時間%d分\n", r.TotalMinutes/60, r.TotalMinutes%60)
	result += fmt.Sprintf("  基礎額: %d円\n", r.BaseFare)

	if r.ExcessKm > 0 {
		result += fmt.Sprintf("  距離超過: +%dkm → +%d円\n", r.ExcessKm, r.DistanceSurcharge)
	}
	if r.ExcessMinutes > 0 {
		result += fmt.Sprintf("  時間超過: +%d分（%d時間） → +%d円\n", r.ExcessMinutes, r.ExcessHours, r.TimeSurcharge)
	}

	result += fmt.Sprintf("  小計（割増前）: %d円\n", r.SubTotal)

	if r.IsNight {
		result += fmt.Sprintf("  深夜割増: +%d円（%.0f%%増）\n", r.NightSurcharge, (r.NightRate-1.0)*100)
	}
	if r.IsHoliday {
		result += fmt.Sprintf("  休日割増: +%d円（%.0f%%増）\n", r.HolidaySurcharge, (r.HolidayRate-1.0)*100)
	}

	result += fmt.Sprintf("  合計運賃: %d円\n", r.TotalFare)

	return result
}
