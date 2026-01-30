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

// HighwayHandler 高速道路関連のハンドラ
type HighwayHandler struct {
	mainDB     *sql.DB
	cacheDB    *sql.DB
	icRepo     *repository.HighwayICRepository
	tollRepo   *repository.HighwayTollRepository
	drivePlaza *service.DrivePlazaClient
}

// NewHighwayHandler ハンドラを作成する
func NewHighwayHandler(mainDB, cacheDB *sql.DB) *HighwayHandler {
	return &HighwayHandler{
		mainDB:     mainDB,
		cacheDB:    cacheDB,
		icRepo:     repository.NewHighwayICRepository(mainDB),
		tollRepo:   repository.NewHighwayTollRepository(cacheDB),
		drivePlaza: service.NewDrivePlazaClient(),
	}
}

// SearchICResponse IC検索レスポンス
type SearchICResponse struct {
	ICs []*ICItem `json:"ics"`
}

// ICItem IC情報
type ICItem struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	RoadName string `json:"road_name"`
	Display  string `json:"display"`
}

// SearchIC IC名で検索するAPI
// GET /api/highway/ic/search?q=東京
func (h *HighwayHandler) SearchIC(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "検索キーワードが必要です"})
	}

	// 名前で検索（部分一致）
	ics, err := h.icRepo.SearchByName(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "検索エラー"})
	}

	// 結果が少なければ読みでも検索
	if len(ics) < 10 {
		yomiICs, err := h.icRepo.SearchByYomi(query)
		if err == nil {
			existingCodes := make(map[string]bool)
			for _, ic := range ics {
				existingCodes[ic.Code] = true
			}
			for _, ic := range yomiICs {
				if !existingCodes[ic.Code] {
					ics = append(ics, ic)
				}
			}
		}
	}

	// 最大20件に制限
	if len(ics) > 20 {
		ics = ics[:20]
	}

	items := make([]*ICItem, len(ics))
	for i, ic := range ics {
		items[i] = &ICItem{
			Code:     ic.Code,
			Name:     ic.Name,
			RoadName: ic.RoadName,
			Display:  ic.Name + " " + ic.RoadName,
		}
	}

	return c.JSON(http.StatusOK, &SearchICResponse{ICs: items})
}

// TollResponse 料金取得レスポンス
type TollResponse struct {
	Success     bool    `json:"success"`
	Error       string  `json:"error,omitempty"`
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

// GetToll 高速料金を取得するAPI
// GET /api/highway/toll?origin=東京&dest=名古屋&car_type=3
func (h *HighwayHandler) GetToll(c echo.Context) error {
	originIC := c.QueryParam("origin")
	destIC := c.QueryParam("dest")
	carTypeStr := c.QueryParam("car_type")

	if originIC == "" || destIC == "" {
		return c.JSON(http.StatusOK, &TollResponse{
			Success: false,
			Error:   "出発IC・到着ICは必須です",
		})
	}

	carType := model.CarTypeLarge // デフォルト: 大型車
	if carTypeStr != "" {
		ct, err := strconv.Atoi(carTypeStr)
		if err != nil || ct < 0 || ct > 4 {
			return c.JSON(http.StatusOK, &TollResponse{
				Success: false,
				Error:   "車種区分が不正です（0-4）",
			})
		}
		carType = ct
	}

	// キャッシュを確認
	if h.tollRepo.Exists(originIC, destIC, carType) {
		toll, err := h.tollRepo.Get(originIC, destIC, carType)
		if err == nil {
			return c.JSON(http.StatusOK, buildTollResponse(toll, true))
		}
	}

	// ドラぷらから取得
	toll, err := h.drivePlaza.FetchToll(originIC, destIC, carType)
	if err != nil {
		return c.JSON(http.StatusOK, &TollResponse{
			Success: false,
			Error:   "料金取得エラー: " + err.Error(),
		})
	}

	// キャッシュに保存
	h.tollRepo.Upsert(toll)

	return c.JSON(http.StatusOK, buildTollResponse(toll, false))
}

func buildTollResponse(toll *model.HighwayToll, fromCache bool) *TollResponse {
	carTypeNames := map[int]string{
		0: "軽自動車等",
		1: "普通車",
		2: "中型車",
		3: "大型車",
		4: "特大車",
	}

	return &TollResponse{
		Success:     true,
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
