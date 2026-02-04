package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

// CalculateHandler 運賃計算ハンドラ
type CalculateHandler struct {
	fareCalculator     *service.FareCalculatorService
	cachedRouteService *service.CachedRouteService
	apiUsageService    *service.ApiUsageService
	geocodingClient    service.GeocodingClient
	// 高速料金関連
	icRepo     *repository.HighwayICRepository
	tollRepo   *repository.HighwayTollRepository
	drivePlaza *service.DrivePlazaClient
}

// NewCalculateHandler 新しいCalculateHandlerを作成
func NewCalculateHandler(fareCalculator *service.FareCalculatorService, cachedRouteService *service.CachedRouteService, apiUsageService *service.ApiUsageService, geocodingClient service.GeocodingClient, mainDB, cacheDB *sql.DB) *CalculateHandler {
	// デフォルト値の設定
	if fareCalculator == nil {
		fareCalculator = createMockFareCalculator()
	}
	if geocodingClient == nil {
		geocodingClient = service.NewMockGeocodingClient()
	}

	h := &CalculateHandler{
		fareCalculator:     fareCalculator,
		cachedRouteService: cachedRouteService,
		apiUsageService:    apiUsageService,
		geocodingClient:    geocodingClient,
	}

	// 高速料金関連（DBが渡された場合のみ初期化）
	if mainDB != nil && cacheDB != nil {
		h.icRepo = repository.NewHighwayICRepository(mainDB)
		h.tollRepo = repository.NewHighwayTollRepository(cacheDB)
		h.drivePlaza = service.NewDrivePlazaClient()
	}

	return h
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

	// 高速道路パラメータ
	UseHighway bool   `form:"use_highway"` // 高速道路使用
	OriginIC   string `form:"origin_ic"`   // 乗IC
	DestIC     string `form:"dest_ic"`     // 降IC
}

// CalculateResultWithHighway 運賃計算結果＋高速料金
type CalculateResultWithHighway struct {
	*service.FareComparisonResult
	// 高速料金
	UseHighway   bool              `json:"use_highway"`
	HighwayToll  *HighwayTollInfo  `json:"highway_toll,omitempty"`
	HighwayError string            `json:"highway_error,omitempty"`
	// 合計金額
	TotalWithHighway *TotalWithHighway `json:"total_with_highway,omitempty"`
}

// HighwayTollInfo 高速料金情報
type HighwayTollInfo struct {
	OriginIC    string  `json:"origin_ic"`
	DestIC      string  `json:"dest_ic"`
	CarType     int     `json:"car_type"`
	CarTypeName string  `json:"car_type_name"`
	NormalToll  int     `json:"normal_toll"`
	EtcToll     int     `json:"etc_toll"`
	Etc2Toll    int     `json:"etc2_toll"`
	DistanceKm  float64 `json:"distance_km"`
	DurationMin int     `json:"duration_min"`
	FromCache   bool    `json:"from_cache"`
}

// TotalWithHighway 運賃＋高速代の合計
type TotalWithHighway struct {
	MinFare      int `json:"min_fare"`       // 最安運賃
	MaxFare      int `json:"max_fare"`       // 最高運賃
	HighwayToll  int `json:"highway_toll"`   // 高速代（ETC料金）
	MinTotal     int `json:"min_total"`      // 最安合計
	MaxTotal     int `json:"max_total"`      // 最高合計
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
	fareResult, err := h.fareCalculator.CalculateAll(&service.FareCalculationRequest{
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

	// 結果を構築
	result := &CalculateResultWithHighway{
		FareComparisonResult: fareResult,
		UseHighway:           req.UseHighway,
	}

	// 高速料金を取得（高速道路使用時）
	if req.UseHighway && req.OriginIC != "" && req.DestIC != "" {
		// 車格から高速料金車種を自動マッピング
		highwayCarType := vehicleCodeToHighwayCarType(req.VehicleCode)
		tollInfo, tollErr := h.fetchHighwayToll(req.OriginIC, req.DestIC, highwayCarType)
		if tollErr != nil {
			result.HighwayError = tollErr.Error()
		} else {
			result.HighwayToll = tollInfo
			// 合計金額を計算
			result.TotalWithHighway = &TotalWithHighway{
				MinFare:     fareResult.CheapestFare,
				MaxFare:     fareResult.Rankings[len(fareResult.Rankings)-1].Fare,
				HighwayToll: tollInfo.EtcToll, // ETC料金を使用
				MinTotal:    fareResult.CheapestFare + tollInfo.EtcToll,
				MaxTotal:    fareResult.Rankings[len(fareResult.Rankings)-1].Fare + tollInfo.EtcToll,
			}
		}
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
	fareResult, err := h.fareCalculator.CalculateAll(&service.FareCalculationRequest{
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

	// 結果を構築
	result := &CalculateResultWithHighway{
		FareComparisonResult: fareResult,
		UseHighway:           req.UseHighway,
	}

	// 高速料金を取得（高速道路使用時）
	if req.UseHighway && req.OriginIC != "" && req.DestIC != "" {
		// 車格から高速料金車種を自動マッピング
		highwayCarType := vehicleCodeToHighwayCarType(req.VehicleCode)
		tollInfo, tollErr := h.fetchHighwayToll(req.OriginIC, req.DestIC, highwayCarType)
		if tollErr != nil {
			result.HighwayError = tollErr.Error()
		} else {
			result.HighwayToll = tollInfo
			// 合計金額を計算
			result.TotalWithHighway = &TotalWithHighway{
				MinFare:     fareResult.CheapestFare,
				MaxFare:     fareResult.Rankings[len(fareResult.Rankings)-1].Fare,
				HighwayToll: tollInfo.EtcToll,
				MinTotal:    fareResult.CheapestFare + tollInfo.EtcToll,
				MaxTotal:    fareResult.Rankings[len(fareResult.Rankings)-1].Fare + tollInfo.EtcToll,
			}
		}
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

	// 高速道路パラメータ
	req.UseHighway = c.FormValue("use_highway") == "true"
	req.OriginIC = c.FormValue("origin_ic")
	req.DestIC = c.FormValue("dest_ic")

	// origin/dest が指定されている場合、ルート情報から距離・時間・運輸局を取得
	// ただし、距離と走行時間が手入力されている場合はスキップ（API上限到達時の手入力モード対応）
	if req.Origin != "" && req.Dest != "" {
		// 手入力値があるかチェック
		hasManualDistance := c.FormValue("distance_km") != ""
		hasManualDriving := c.FormValue("driving_minutes") != ""

		if hasManualDistance && hasManualDriving {
			// 手入力モード：API呼び出しをスキップ
			// 赤帽地区の判定のみ行う
			if req.Area == "" {
				req.Area = service.ResolveAkabouArea(req.Origin)
			}
		} else {
			// 自動取得モード
			if err := h.resolveRouteInfo(req); err != nil {
				return nil, err
			}
		}
	}

	return req, nil
}

// resolveRouteInfo 出発地/目的地からルート情報を取得してリクエストに設定
func (h *CalculateHandler) resolveRouteInfo(req *CalculateRequest) error {
	// Geocoding APIで出発地から都道府県を取得
	prefecture, err := h.geocodingClient.GetPrefecture(req.Origin)
	if err != nil {
		return &ValidationError{Message: "出発地の都道府県を特定できません: " + req.Origin + " (" + err.Error() + ")"}
	}

	// 都道府県から運輸局コードを取得
	regionCode, err := service.ResolveRegionCode(prefecture)
	if err != nil {
		return &ValidationError{Message: "運輸局を特定できません: " + prefecture}
	}
	req.RegionCode = regionCode

	// 赤帽地区を判定（Geocodingで取得した住所情報を使用）
	if req.Area == "" {
		// Geocodingで詳細住所を取得して判定
		components, err := h.geocodingClient.GetAddressComponents(req.Origin)
		if err == nil && components != nil {
			req.Area = service.ResolveAkabouArea(components.Address)
		} else {
			req.Area = service.ResolveAkabouArea(req.Origin)
		}
	}

	// Routes APIで距離・時間を取得（キャッシュ付き）
	if h.cachedRouteService == nil {
		return &ValidationError{Message: "ルートサービスが初期化されていません"}
	}

	result, err := h.cachedRouteService.GetRoute(req.Origin, req.Dest)
	if err != nil {
		return &ValidationError{Message: "ルート取得エラー: " + err.Error()}
	}

	// キャッシュミス時（API呼び出し時）はAPI使用量をカウントアップ
	if !result.FromCache && h.apiUsageService != nil {
		_ = h.apiUsageService.IncrementAndCheck()
	}

	req.DistanceKm = int(result.Route.DistanceKm)
	req.DrivingMinutes = result.Route.DurationMin

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
	if req.VehicleCode < 0 || req.VehicleCode > 4 {
		return &ValidationError{Message: "車格コードが不正です（0-4）"}
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

// vehicleCodeToHighwayCarType 車格コードから高速料金車種を自動マッピング
func vehicleCodeToHighwayCarType(vehicleCode int) int {
	switch vehicleCode {
	case 0: // 軽貨物/赤帽
		return model.CarTypeLight // 軽自動車等
	case 1: // 小型車（2t）
		return model.CarTypeNormal // 普通車
	case 2: // 中型車（4t）
		return model.CarTypeMedium // 中型車
	case 3: // 大型車（10t）
		return model.CarTypeLarge // 大型車
	case 4: // トレーラー（20t）
		return model.CarTypeSpecial // 特大車
	default:
		return model.CarTypeLarge // デフォルト: 大型車
	}
}

// fetchHighwayToll 高速料金を取得
func (h *CalculateHandler) fetchHighwayToll(originIC, destIC string, carType int) (*HighwayTollInfo, error) {
	if h.tollRepo == nil || h.drivePlaza == nil {
		return nil, &ValidationError{Message: "高速料金取得機能が初期化されていません"}
	}

	// キャッシュを確認
	if h.tollRepo.Exists(originIC, destIC, carType) {
		toll, err := h.tollRepo.Get(originIC, destIC, carType)
		if err == nil {
			return buildHighwayTollInfo(toll, true), nil
		}
	}

	// ドラぷらから取得
	toll, err := h.drivePlaza.FetchToll(originIC, destIC, carType)
	if err != nil {
		return nil, &ValidationError{Message: "高速料金取得エラー: " + err.Error()}
	}

	// キャッシュに保存
	h.tollRepo.Upsert(toll)

	return buildHighwayTollInfo(toll, false), nil
}

// buildHighwayTollInfo HighwayTollからHighwayTollInfoを作成
func buildHighwayTollInfo(toll *model.HighwayToll, fromCache bool) *HighwayTollInfo {
	// 高速道路の車種区分名
	carTypeNames := map[int]string{
		0: "軽自動車等",
		1: "普通車",
		2: "中型車",
		3: "大型車",
		4: "特大車",
	}

	return &HighwayTollInfo{
		OriginIC:    toll.OriginIC,
		DestIC:      toll.DestIC,
		CarType:     toll.CarType,
		CarTypeName: carTypeNames[toll.CarType],
		NormalToll:  toll.NormalToll,
		EtcToll:     toll.EtcToll,
		Etc2Toll:    toll.Etc2Toll,
		DistanceKm:  toll.DistanceKm,
		DurationMin: toll.DurationMin,
		FromCache:   fromCache,
	}
}
