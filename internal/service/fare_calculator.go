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

	// 時間制用パラメータ
	DrivingMinutes  int  // 走行時間（分）- Google Maps API取得値
	LoadingMinutes  int  // 荷役時間（分）- デフォルト60分
	UseSimpleBaseKm bool // シンプル版基礎走行キロ使用（false=トラ協PDF版）

	// 赤帽用パラメータ
	Area string // 地区（東京23区、大阪市内など）
}

// FareRanking 運賃ランキング
type FareRanking struct {
	Rank int    // 順位（1が最安）
	Type string // 運賃タイプ
	Fare int    // 運賃額（円）
}

// FareComparisonResult 運賃比較結果
type FareComparisonResult struct {
	// 各運賃の計算結果
	DistanceFareResult   *DistanceFareResult       // 距離制運賃
	TimeFareResult       *TimeFareResult           // 時間制運賃
	AkabouDistanceResult *AkabouDistanceFareResult // 赤帽運賃（距離制）
	AkabouTimeResult     *AkabouTimeFareResult     // 赤帽運賃（時間制）

	// 比較・ランキング
	Rankings     []FareRanking // 金額順ランキング
	CheapestType string        // 最安運賃タイプ
	CheapestFare int           // 最安運賃額（円）
}

// CalculateAll 3運賃を一括計算する
func (s *FareCalculatorService) CalculateAll(req *FareCalculationRequest) (*FareComparisonResult, error) {
	result := &FareComparisonResult{}

	// 1. 距離制運賃を計算
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

	// 2. 時間制運賃を計算
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

	// 3. 赤帽運賃（距離制）を計算
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

	// 4. 赤帽運賃（時間制）を計算
	// 総作業時間 = 走行時間 + 荷役時間
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

	// 5. ランキングを生成
	result.Rankings = s.createRankings(result)

	// 6. 最安値を設定
	if len(result.Rankings) > 0 {
		result.CheapestType = result.Rankings[0].Type
		result.CheapestFare = result.Rankings[0].Fare
	}

	return result, nil
}

// createRankings ランキングを生成（金額昇順）
func (s *FareCalculatorService) createRankings(result *FareComparisonResult) []FareRanking {
	rankings := []FareRanking{
		{Type: "距離制", Fare: result.DistanceFareResult.TotalFare},
		{Type: "時間制", Fare: result.TimeFareResult.TotalFare},
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

	// 各運賃の詳細
	result += "----------------------------------------\n"
	result += r.DistanceFareResult.Breakdown()
	result += "\n----------------------------------------\n"
	result += r.TimeFareResult.Breakdown()
	result += "\n----------------------------------------\n"
	result += r.AkabouDistanceResult.Breakdown()
	result += "\n----------------------------------------\n"
	result += r.AkabouTimeResult.Breakdown()

	return result
}
