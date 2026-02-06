package service

import (
	"fmt"
	"sort"
)

// FareCalculatorService 統合運賃計算サービス
// 距離制・時間制・赤帽の3運賃を一括計算する
type FareCalculatorService struct {
	distanceFare *DistanceFareService
	timeFare     *TimeFareService
	akabouFare   *AkabouFareService
}

// NewFareCalculatorService 新しいFareCalculatorServiceを作成
func NewFareCalculatorService(
	distanceFare *DistanceFareService,
	timeFare *TimeFareService,
	akabouFare *AkabouFareService,
) *FareCalculatorService {
	return &FareCalculatorService{
		distanceFare: distanceFare,
		timeFare:     timeFare,
		akabouFare:   akabouFare,
	}
}

// FareCalculationRequest 運賃計算リクエスト
type FareCalculationRequest struct {
	// 共通パラメータ
	RegionCode  int  // 運輸局コード（1-10）
	VehicleCode int  // 車格コード（1-4）
	DistanceKm  int  // 距離（km）
	IsNight     bool // 深夜割増
	IsHoliday   bool // 休日割増

	// 距離（表示用）
	DistanceKmRaw float64 // 元距離（km、小数点付き）- Google Maps API取得値

	// 時間制用パラメータ
	DrivingMinutes  int  // 走行時間（分）- Google Maps API取得値
	LoadingMinutes  int  // 荷役時間（分）- デフォルト60分
	UseSimpleBaseKm bool // シンプル版基礎走行キロ使用（false=トラ協PDF版）

	// 赤帽用パラメータ
	Area           string // 地区（東京23区、大阪市内など）
	WorkMinutes    int    // 作業時間（分）- 赤帽付帯料金用
	WaitingMinutes int    // 待機時間（分）- 赤帽付帯料金用
}

// FareRanking 運賃ランキング
type FareRanking struct {
	Rank int    // 順位（1が最安）
	Type string // 運賃タイプ
	Fare int    // 運賃額（円）
}

// FareComparisonResult 運賃比較結果
type FareComparisonResult struct {
	// 車格コード（0=軽貨物, 1-4=トラック）
	VehicleCode int

	// 共通情報（計算根拠表示用）
	DistanceKmRaw  float64 // 元距離（km、小数点付き）
	DrivingMinutes int     // 走行時間（分）
	LoadingMinutes int     // 荷役時間（分）

	// 各運賃の計算結果
	DistanceFareResult   *DistanceFareResult        // 距離制運賃（トラック用）
	TimeFareResult       *TimeFareResult            // 時間制運賃（トラック用）
	AkabouDistanceResult *AkabouDistanceFareResult  // 赤帽運賃（距離制、軽貨物用）
	AkabouTimeResult     *AkabouTimeFareResult      // 赤帽運賃（時間制、軽貨物用）
	AdditionalFees       *AkabouAdditionalFeesResult // 赤帽付帯料金（軽貨物用）

	// 比較・ランキング
	Rankings     []FareRanking // 金額順ランキング
	CheapestType string        // 最安運賃タイプ
	CheapestFare int           // 最安運賃額（円）
}

// VehicleCodeLight 軽貨物/赤帽の車格コード
const VehicleCodeLight = 0

// CalculateAll 運賃を一括計算する
// 軽貨物（VehicleCode=0）の場合は赤帽のみ、2t以上（VehicleCode=1-4）の場合はトラ協のみを計算
func (s *FareCalculatorService) CalculateAll(req *FareCalculationRequest) (*FareComparisonResult, error) {
	result := &FareComparisonResult{
		VehicleCode:    req.VehicleCode,
		DistanceKmRaw:  req.DistanceKmRaw,
		DrivingMinutes: req.DrivingMinutes,
		LoadingMinutes: req.LoadingMinutes,
	}

	// 軽貨物（赤帽）の場合
	if req.VehicleCode == VehicleCodeLight {
		// 赤帽運賃（距離制）を計算
		akabouDistanceResult, err := s.akabouFare.CalculateDistanceFare(
			req.DistanceKm,
			req.IsNight,
			req.IsHoliday,
			req.Area,
		)
		if err != nil {
			return nil, fmt.Errorf("赤帽距離制運賃計算エラー: %w", err)
		}
		result.AkabouDistanceResult = akabouDistanceResult

		// 赤帽運賃（時間制）を計算
		totalMinutes := req.DrivingMinutes + req.LoadingMinutes
		akabouTimeResult, err := s.akabouFare.CalculateTimeFare(
			totalMinutes,
			req.IsNight,
			req.IsHoliday,
			req.Area,
		)
		if err != nil {
			return nil, fmt.Errorf("赤帽時間制運賃計算エラー: %w", err)
		}
		result.AkabouTimeResult = akabouTimeResult

		// 付帯料金を計算
		additionalFees := s.akabouFare.CalculateAdditionalFees(req.WorkMinutes, req.WaitingMinutes)
		result.AdditionalFees = additionalFees

		// 付帯料金をTotalFareに加算
		if additionalFees.TotalFee > 0 {
			result.AkabouDistanceResult.TotalFare += additionalFees.TotalFee
			result.AkabouTimeResult.TotalFare += additionalFees.TotalFee
		}

		// ランキングを生成（赤帽のみ）
		result.Rankings = s.createRankingsForLight(result)
	} else {
		// 2t以上（トラ協）の場合
		// 距離制運賃を計算
		distanceResult, err := s.distanceFare.Calculate(
			req.RegionCode,
			req.VehicleCode,
			req.DistanceKm,
			req.IsNight,
			req.IsHoliday,
		)
		if err != nil {
			return nil, fmt.Errorf("距離制運賃計算エラー: %w", err)
		}
		result.DistanceFareResult = distanceResult

		// 時間制運賃を計算
		timeResult, err := s.timeFare.Calculate(
			req.RegionCode,
			req.VehicleCode,
			req.DistanceKm,
			req.DrivingMinutes,
			req.LoadingMinutes,
			req.IsNight,
			req.IsHoliday,
			req.UseSimpleBaseKm,
		)
		if err != nil {
			return nil, fmt.Errorf("時間制運賃計算エラー: %w", err)
		}
		result.TimeFareResult = timeResult

		// ランキングを生成（トラ協のみ）
		result.Rankings = s.createRankingsForTruck(result)
	}

	// 最安値を設定
	if len(result.Rankings) > 0 {
		result.CheapestType = result.Rankings[0].Type
		result.CheapestFare = result.Rankings[0].Fare
	}

	return result, nil
}

// createRankingsForLight 軽貨物用ランキングを生成（赤帽のみ）
func (s *FareCalculatorService) createRankingsForLight(result *FareComparisonResult) []FareRanking {
	rankings := []FareRanking{
		{Type: "赤帽（距離制）", Fare: result.AkabouDistanceResult.TotalFare},
		{Type: "赤帽（時間制）", Fare: result.AkabouTimeResult.TotalFare},
	}

	// 金額昇順でソート
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].Fare < rankings[j].Fare
	})

	// 順位を設定
	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	return rankings
}

// createRankingsForTruck 2t以上用ランキングを生成（トラ協のみ）
func (s *FareCalculatorService) createRankingsForTruck(result *FareComparisonResult) []FareRanking {
	rankings := []FareRanking{
		{Type: "距離制", Fare: result.DistanceFareResult.TotalFare},
		{Type: "時間制", Fare: result.TimeFareResult.TotalFare},
	}

	// 金額昇順でソート
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].Fare < rankings[j].Fare
	})

	// 順位を設定
	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	return rankings
}

// Breakdown 計算根拠を文字列で返す
func (r *FareComparisonResult) Breakdown() string {
	result := "========================================\n"
	result += "【運賃比較結果】\n"
	result += "========================================\n\n"

	// ランキング表示
	result += "【ランキング】\n"
	for _, ranking := range r.Rankings {
		marker := ""
		if ranking.Rank == 1 {
			marker = " ← 最安"
		}
		result += fmt.Sprintf("  %d位: %s %d円%s\n", ranking.Rank, ranking.Type, ranking.Fare, marker)
	}
	result += "\n"

	// 各運賃の詳細（車格に応じて表示）
	if r.VehicleCode == VehicleCodeLight {
		// 軽貨物: 赤帽のみ
		result += "----------------------------------------\n"
		result += r.AkabouDistanceResult.Breakdown()
		result += "\n----------------------------------------\n"
		result += r.AkabouTimeResult.Breakdown()
	} else {
		// トラック: トラ協のみ
		result += "----------------------------------------\n"
		result += r.DistanceFareResult.Breakdown()
		result += "\n----------------------------------------\n"
		result += r.TimeFareResult.Breakdown()
	}

	return result
}
