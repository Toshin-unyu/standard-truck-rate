package service

import (
	"fmt"
)

// 赤帽運賃定数（税込）
const (
	// 距離制運賃
	AkabouDistanceBaseFare    = 5500 // 基本料金（20km以内）
	AkabouDistanceRate21to50  = 242  // 21-50km: 円/km
	AkabouDistanceRate51to100 = 187  // 51-100km: 円/km
	AkabouDistanceRate101to150 = 154 // 101-150km: 円/km
	AkabouDistanceRate151plus = 132  // 151km以上: 円/km

	// 時間制運賃
	AkabouTimeBaseFare     = 6050 // 基本料金（2時間・20km以内）
	AkabouTimeBaseMinutes  = 120  // 基本時間（分）
	AkabouTimeOvertimeRate = 1375 // 超過料金（30分ごと）
	AkabouTimeOvertimeUnit = 30   // 超過単位（分）

	// 割増率
	AkabouNightSurchargeRate   = 1.3 // 深夜割増: +30%
	AkabouHolidaySurchargeRate = 1.2 // 休日割増: +20%

	// 地区割増
	AkabouAreaSurcharge = 440 // 東京23区・大阪市内

	// 付帯料金
	AkabouWorkFreeMinutes = 30   // 作業料金無料時間（分）
	AkabouWorkUnitMinutes = 15   // 作業料金課金単位（分）
	AkabouWorkFeePerUnit  = 550  // 作業料金（円/15分）
	AkabouWaitFreeMinutes = 30   // 待機時間無料時間（分）
	AkabouWaitUnitMinutes = 30   // 待機時間課金単位（分）
	AkabouWaitFeePerUnit  = 1100 // 待機時間料（円/30分）
)

// 地区割増対象エリア
var akabouSurchargeAreas = map[string]bool{
	"東京23区": true,
	"大阪市内": true,
}

// AkabouFareService 赤帽運賃計算サービス
type AkabouFareService struct{}

// NewAkabouFareService 新しいAkabouFareServiceを作成
func NewAkabouFareService() *AkabouFareService {
	return &AkabouFareService{}
}

// AkabouDistanceFareResult 赤帽距離制運賃計算結果
type AkabouDistanceFareResult struct {
	DistanceKm       int     // 距離 (km)
	BaseFare         int     // 基本料金（円）
	DistanceCharge   int     // 距離加算（円）
	AreaSurcharge    int     // 地区割増（円）
	NightSurcharge   int     // 深夜割増額（円）
	HolidaySurcharge int     // 休日割増額（円）
	TotalFare        int     // 合計運賃（円）
	IsNight          bool    // 深夜適用
	IsHoliday        bool    // 休日適用
	Area             string  // 地区
	NightRate        float64 // 深夜割増率
	HolidayRate      float64 // 休日割増率
}

// AkabouTimeFareResult 赤帽時間制運賃計算結果
type AkabouTimeFareResult struct {
	DurationMin      int     // 作業時間（分）
	BaseFare         int     // 基本料金（円）
	OvertimeCharge   int     // 超過料金（円）
	OvertimeMin      int     // 超過時間（分）
	AreaSurcharge    int     // 地区割増（円）
	NightSurcharge   int     // 深夜割増額（円）
	HolidaySurcharge int     // 休日割増額（円）
	TotalFare        int     // 合計運賃（円）
	IsNight          bool    // 深夜適用
	IsHoliday        bool    // 休日適用
	Area             string  // 地区
	NightRate        float64 // 深夜割増率
	HolidayRate      float64 // 休日割増率
}

// CalculateDistanceFare 距離制運賃を計算
func (s *AkabouFareService) CalculateDistanceFare(
	distanceKm int,
	isNight, isHoliday bool,
	area string,
) (*AkabouDistanceFareResult, error) {
	// 入力値検証
	if distanceKm < 1 {
		return nil, fmt.Errorf("無効な距離: %d（1km以上を指定）", distanceKm)
	}

	// 基本料金
	baseFare := AkabouDistanceBaseFare

	// 距離加算を計算
	distanceCharge := calculateDistanceCharge(distanceKm)

	// 地区割増
	areaSurcharge := 0
	if akabouSurchargeAreas[area] {
		areaSurcharge = AkabouAreaSurcharge
	}

	// 小計（割増前）
	subtotal := baseFare + distanceCharge + areaSurcharge

	// 割増計算
	nightRate := 1.0
	holidayRate := 1.0
	nightSurcharge := 0
	holidaySurcharge := 0
	totalFare := subtotal

	// 深夜割増（+30%）
	if isNight {
		nightRate = AkabouNightSurchargeRate
		nightSurcharge = int(float64(subtotal)*(AkabouNightSurchargeRate-1.0))
		totalFare = int(float64(subtotal) * AkabouNightSurchargeRate)
	}

	// 休日割増（+20%）- 深夜割増後に適用
	if isHoliday {
		holidayRate = AkabouHolidaySurchargeRate
		holidaySurcharge = int(float64(totalFare) * (AkabouHolidaySurchargeRate - 1.0))
		totalFare = int(float64(totalFare) * AkabouHolidaySurchargeRate)
	}

	return &AkabouDistanceFareResult{
		DistanceKm:       distanceKm,
		BaseFare:         baseFare,
		DistanceCharge:   distanceCharge,
		AreaSurcharge:    areaSurcharge,
		NightSurcharge:   nightSurcharge,
		HolidaySurcharge: holidaySurcharge,
		TotalFare:        totalFare,
		IsNight:          isNight,
		IsHoliday:        isHoliday,
		Area:             area,
		NightRate:        nightRate,
		HolidayRate:      holidayRate,
	}, nil
}

// calculateDistanceCharge 距離加算を計算
func calculateDistanceCharge(distanceKm int) int {
	if distanceKm <= 20 {
		return 0
	}

	charge := 0

	// 21-50km区間
	if distanceKm > 20 {
		km := min(distanceKm, 50) - 20
		charge += km * AkabouDistanceRate21to50
	}

	// 51-100km区間
	if distanceKm > 50 {
		km := min(distanceKm, 100) - 50
		charge += km * AkabouDistanceRate51to100
	}

	// 101-150km区間
	if distanceKm > 100 {
		km := min(distanceKm, 150) - 100
		charge += km * AkabouDistanceRate101to150
	}

	// 151km以上区間
	if distanceKm > 150 {
		km := distanceKm - 150
		charge += km * AkabouDistanceRate151plus
	}

	return charge
}

// CalculateTimeFare 時間制運賃を計算
func (s *AkabouFareService) CalculateTimeFare(
	durationMin int,
	isNight, isHoliday bool,
	area string,
) (*AkabouTimeFareResult, error) {
	// 入力値検証
	if durationMin < 1 {
		return nil, fmt.Errorf("無効な時間: %d（1分以上を指定）", durationMin)
	}

	// 基本料金
	baseFare := AkabouTimeBaseFare

	// 超過時間を計算（30分単位で切り上げ）
	overtimeMin := 0
	overtimeCharge := 0
	if durationMin > AkabouTimeBaseMinutes {
		overtimeMin = durationMin - AkabouTimeBaseMinutes
		// 30分単位で切り上げ
		overtimeUnits := (overtimeMin + AkabouTimeOvertimeUnit - 1) / AkabouTimeOvertimeUnit
		overtimeCharge = overtimeUnits * AkabouTimeOvertimeRate
	}

	// 地区割増
	areaSurcharge := 0
	if akabouSurchargeAreas[area] {
		areaSurcharge = AkabouAreaSurcharge
	}

	// 小計（割増前）
	subtotal := baseFare + overtimeCharge + areaSurcharge

	// 割増計算
	nightRate := 1.0
	holidayRate := 1.0
	nightSurcharge := 0
	holidaySurcharge := 0
	totalFare := subtotal

	// 深夜割増（+30%）
	if isNight {
		nightRate = AkabouNightSurchargeRate
		nightSurcharge = int(float64(subtotal) * (AkabouNightSurchargeRate - 1.0))
		totalFare = int(float64(subtotal) * AkabouNightSurchargeRate)
	}

	// 休日割増（+20%）- 深夜割増後に適用
	if isHoliday {
		holidayRate = AkabouHolidaySurchargeRate
		holidaySurcharge = int(float64(totalFare) * (AkabouHolidaySurchargeRate - 1.0))
		totalFare = int(float64(totalFare) * AkabouHolidaySurchargeRate)
	}

	return &AkabouTimeFareResult{
		DurationMin:      durationMin,
		BaseFare:         baseFare,
		OvertimeCharge:   overtimeCharge,
		OvertimeMin:      overtimeMin,
		AreaSurcharge:    areaSurcharge,
		NightSurcharge:   nightSurcharge,
		HolidaySurcharge: holidaySurcharge,
		TotalFare:        totalFare,
		IsNight:          isNight,
		IsHoliday:        isHoliday,
		Area:             area,
		NightRate:        nightRate,
		HolidayRate:      holidayRate,
	}, nil
}

// AkabouAdditionalFeesResult 赤帽付帯料金計算結果
type AkabouAdditionalFeesResult struct {
	WorkMinutes    int // 作業時間（分）
	WaitingMinutes int // 待機時間（分）
	WorkFee        int // 作業料金（円）
	WaitingFee     int // 待機時間料（円）
	TotalFee       int // 付帯料金合計（円）
}

// CalculateAdditionalFees 付帯料金を計算
func (s *AkabouFareService) CalculateAdditionalFees(workMinutes, waitingMinutes int) *AkabouAdditionalFeesResult {
	result := &AkabouAdditionalFeesResult{
		WorkMinutes:    workMinutes,
		WaitingMinutes: waitingMinutes,
	}

	// 作業料金: 30分まで無料、超過15分ごとに550円（切り上げ）
	if workMinutes > AkabouWorkFreeMinutes {
		excessMin := workMinutes - AkabouWorkFreeMinutes
		units := (excessMin + AkabouWorkUnitMinutes - 1) / AkabouWorkUnitMinutes
		result.WorkFee = units * AkabouWorkFeePerUnit
	}

	// 待機時間料: 30分まで無料、超過30分ごとに1,100円（切り上げ）
	if waitingMinutes > AkabouWaitFreeMinutes {
		excessMin := waitingMinutes - AkabouWaitFreeMinutes
		units := (excessMin + AkabouWaitUnitMinutes - 1) / AkabouWaitUnitMinutes
		result.WaitingFee = units * AkabouWaitFeePerUnit
	}

	result.TotalFee = result.WorkFee + result.WaitingFee
	return result
}

// Breakdown 計算根拠を文字列で返す（距離制）
func (r *AkabouDistanceFareResult) Breakdown() string {
	result := fmt.Sprintf("【赤帽運賃・距離制】\n")
	result += fmt.Sprintf("  距離: %dkm\n", r.DistanceKm)
	result += fmt.Sprintf("  基本料金: %d円\n", r.BaseFare)

	if r.DistanceCharge > 0 {
		result += fmt.Sprintf("  距離加算: +%d円\n", r.DistanceCharge)
	}

	if r.AreaSurcharge > 0 {
		result += fmt.Sprintf("  地区割増（%s）: +%d円\n", r.Area, r.AreaSurcharge)
	}

	if r.IsNight {
		result += fmt.Sprintf("  深夜割増: +%d円（%.0f%%増）\n", r.NightSurcharge, (r.NightRate-1.0)*100)
	}

	if r.IsHoliday {
		result += fmt.Sprintf("  休日割増: +%d円（%.0f%%増）\n", r.HolidaySurcharge, (r.HolidayRate-1.0)*100)
	}

	result += fmt.Sprintf("  合計運賃: %d円\n", r.TotalFare)

	return result
}

// Breakdown 計算根拠を文字列で返す（時間制）
func (r *AkabouTimeFareResult) Breakdown() string {
	result := fmt.Sprintf("【赤帽運賃・時間制】\n")
	result += fmt.Sprintf("  作業時間: %d分（%d時間%d分）\n", r.DurationMin, r.DurationMin/60, r.DurationMin%60)
	result += fmt.Sprintf("  基本料金: %d円（2時間まで）\n", r.BaseFare)

	if r.OvertimeCharge > 0 {
		result += fmt.Sprintf("  超過料金: +%d円（%d分超過）\n", r.OvertimeCharge, r.OvertimeMin)
	}

	if r.AreaSurcharge > 0 {
		result += fmt.Sprintf("  地区割増（%s）: +%d円\n", r.Area, r.AreaSurcharge)
	}

	if r.IsNight {
		result += fmt.Sprintf("  深夜割増: +%d円（%.0f%%増）\n", r.NightSurcharge, (r.NightRate-1.0)*100)
	}

	if r.IsHoliday {
		result += fmt.Sprintf("  休日割増: +%d円（%.0f%%増）\n", r.HolidaySurcharge, (r.HolidayRate-1.0)*100)
	}

	result += fmt.Sprintf("  合計運賃: %d円\n", r.TotalFare)

	return result
}
