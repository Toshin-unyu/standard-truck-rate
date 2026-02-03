package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

// CalculateHandler 運賃計算ハンドラ
type CalculateHandler struct {
	fareCalculator *service.FareCalculatorService
}

// NewCalculateHandler 新しいCalculateHandlerを作成
// fareCalculatorがnilの場合はモックを使用
func NewCalculateHandler(fareCalculator *service.FareCalculatorService) *CalculateHandler {
	if fareCalculator == nil {
		// テスト用: モックサービスを使用
		return &CalculateHandler{
			fareCalculator: createMockFareCalculator(),
		}
	}

	return &CalculateHandler{
		fareCalculator: fareCalculator,
	}
}

// createMockFareCalculator テスト用のモックFareCalculatorを作成
func createMockFareCalculator() *service.FareCalculatorService {
	// モックリポジトリを使用
	distanceFare := service.NewDistanceFareService(&mockFareGetter{})
	timeFare := service.NewTimeFareService(&mockTimeFareGetter{})
	akabouFare := service.NewAkabouFareService()
	return service.NewFareCalculatorService(distanceFare, timeFare, akabouFare)
}

// CalculateRequest 運賃計算リクエスト
type CalculateRequest struct {
	// 新UI: 出発地/目的地入力
	Origin string `form:"origin"` // 出発地（住所）
	Dest   string `form:"dest"`   // 目的地（住所）

	// 旧UI互換: 直接指定（origin/destが指定されていない場合に使用）
	RegionCode     int `form:"region_code"`
	DistanceKm     int `form:"distance_km"`
	DrivingMinutes int `form:"driving_minutes"`

	// 共通パラメータ
	VehicleCode     int    `form:"vehicle_code"`
	LoadingMinutes  int    `form:"loading_minutes"`
	IsNight         bool   `form:"is_night"`
	IsHoliday       bool   `form:"is_holiday"`
	UseSimpleBaseKm bool   `form:"use_simple_base_km"`
	Area            string `form:"area"`
}

// Calculate 運賃を計算してHTMLフラグメントを返す（HTMX用）
// POST /api/fare/calculate
func (h *CalculateHandler) Calculate(c echo.Context) error {
	req, err := h.parseRequest(c)
	if err != nil {
		return c.Render(http.StatusOK, "error", map[string]string{"Error": err.Error()})
	}

	// バリデーション
	if err := h.validateRequest(req); err != nil {
		return c.Render(http.StatusOK, "error", map[string]string{"Error": err.Error()})
	}

	// 運賃計算
	result, err := h.fareCalculator.CalculateAll(&service.FareCalculationRequest{
		RegionCode:      req.RegionCode,
		VehicleCode:     req.VehicleCode,
		DistanceKm:      req.DistanceKm,
		DrivingMinutes:  req.DrivingMinutes,
		LoadingMinutes:  req.LoadingMinutes,
		IsNight:         req.IsNight,
		IsHoliday:       req.IsHoliday,
		UseSimpleBaseKm: req.UseSimpleBaseKm,
		Area:            req.Area,
	})
	if err != nil {
		return c.Render(http.StatusOK, "error", map[string]string{"Error": "運賃計算エラー: " + err.Error()})
	}

	return c.Render(http.StatusOK, "result", result)
}

// CalculateJSON 運賃を計算してJSONを返す（API用）
// POST /api/fare/calculate/json
func (h *CalculateHandler) CalculateJSON(c echo.Context) error {
	req, err := h.parseRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// バリデーション
	if err := h.validateRequest(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// 運賃計算
	result, err := h.fareCalculator.CalculateAll(&service.FareCalculationRequest{
		RegionCode:      req.RegionCode,
		VehicleCode:     req.VehicleCode,
		DistanceKm:      req.DistanceKm,
		DrivingMinutes:  req.DrivingMinutes,
		LoadingMinutes:  req.LoadingMinutes,
		IsNight:         req.IsNight,
		IsHoliday:       req.IsHoliday,
		UseSimpleBaseKm: req.UseSimpleBaseKm,
		Area:            req.Area,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "運賃計算エラー: " + err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

// parseRequest フォームデータをパース
func (h *CalculateHandler) parseRequest(c echo.Context) (*CalculateRequest, error) {
	req := &CalculateRequest{}

	// 出発地/目的地（新UI）
	req.Origin = c.FormValue("origin")
	req.Dest = c.FormValue("dest")

	// 各フィールドを手動でパース（デフォルト値対応）
	if v := c.FormValue("region_code"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.RegionCode = n
		}
	} else {
		req.RegionCode = 3 // デフォルト: 関東
	}

	if v := c.FormValue("vehicle_code"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.VehicleCode = n
		}
	} else {
		req.VehicleCode = 3 // デフォルト: 大型車
	}

	if v := c.FormValue("distance_km"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.DistanceKm = n
		}
	}

	if v := c.FormValue("driving_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.DrivingMinutes = n
		}
	} else {
		req.DrivingMinutes = 60 // デフォルト: 60分
	}

	if v := c.FormValue("loading_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.LoadingMinutes = n
		}
	} else {
		req.LoadingMinutes = 60 // デフォルト: 60分
	}

	req.IsNight = c.FormValue("is_night") == "true"
	req.IsHoliday = c.FormValue("is_holiday") == "true"
	req.UseSimpleBaseKm = c.FormValue("use_simple_base_km") == "true"
	req.Area = c.FormValue("area")

	// origin/dest が指定されている場合、ルート情報から距離・時間・運輸局を取得
	if req.Origin != "" && req.Dest != "" {
		if err := h.resolveRouteInfo(req); err != nil {
			return nil, err
		}
	}

	return req, nil
}

// resolveRouteInfo 出発地/目的地からルート情報を取得してリクエストに設定
func (h *CalculateHandler) resolveRouteInfo(req *CalculateRequest) error {
	// 出発地から都道府県を抽出
	prefecture, ok := service.ExtractPrefectureFromAddress(req.Origin)
	if !ok {
		return &ValidationError{Message: "出発地の都道府県を特定できません: " + req.Origin}
	}

	// 都道府県から運輸局コードを取得
	regionCode, err := service.ResolveRegionCode(prefecture)
	if err != nil {
		return &ValidationError{Message: "運輸局を特定できません: " + prefecture}
	}
	req.RegionCode = regionCode

	// 赤帽地区を判定
	if req.Area == "" {
		req.Area = service.ResolveAkabouArea(req.Origin)
	}

	// モックルートサービスから距離・時間を取得
	mockClient := service.NewMockRoutesClient()
	route, err := mockClient.GetRoute(req.Origin, req.Dest)
	if err != nil {
		return &ValidationError{Message: "ルート取得エラー: " + err.Error()}
	}

	req.DistanceKm = int(route.DistanceKm)
	req.DrivingMinutes = route.DurationMin

	return nil
}

// validateRequest リクエストをバリデーション
func (h *CalculateHandler) validateRequest(req *CalculateRequest) error {
	if req.DistanceKm <= 0 {
		return &ValidationError{Message: "距離を1km以上で入力してください"}
	}
	if req.RegionCode < 1 || req.RegionCode > 10 {
		return &ValidationError{Message: "運輸局コードが不正です（1-10）"}
	}
	if req.VehicleCode < 1 || req.VehicleCode > 4 {
		return &ValidationError{Message: "車格コードが不正です（1-4）"}
	}
	if req.DrivingMinutes <= 0 {
		return &ValidationError{Message: "走行時間を1分以上で入力してください"}
	}
	return nil
}

// ValidationError バリデーションエラー
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// mockFareGetter テスト用の距離制運賃取得モック
type mockFareGetter struct{}

func (m *mockFareGetter) GetDistanceFareYen(regionCode, vehicleCode, distanceKm int) (int, error) {
	// モック: 距離 * 100円
	return distanceKm * 100, nil
}

// mockTimeFareGetter テスト用の時間制運賃取得モック
type mockTimeFareGetter struct{}

func (m *mockTimeFareGetter) GetBaseFare(regionCode, vehicleCode, hours int) (*model.JtaTimeBaseFare, error) {
	// モック: 基礎運賃
	return &model.JtaTimeBaseFare{
		RegionCode:  regionCode,
		VehicleCode: vehicleCode,
		Hours:       hours,
		FareYen:     10000,
		BaseKm:      30,
	}, nil
}

func (m *mockTimeFareGetter) GetSurcharge(regionCode, vehicleCode int, surchargeType string) (*model.JtaTimeSurcharge, error) {
	// モック: 加算額
	fareYen := 0
	switch surchargeType {
	case "distance":
		fareYen = 50
	case "time":
		fareYen = 500
	}
	return &model.JtaTimeSurcharge{
		RegionCode:    regionCode,
		VehicleCode:   vehicleCode,
		SurchargeType: surchargeType,
		FareYen:       fareYen,
	}, nil
}
