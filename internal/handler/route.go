package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

// RouteHandler ルート情報ハンドラ
type RouteHandler struct {
	routeService    *service.CachedRouteService
	apiUsageService *service.ApiUsageService
}

// NewRouteHandler 新しいRouteHandlerを作成
func NewRouteHandler(cacheDB *sql.DB, routeClient service.RouteClient, apiUsageService *service.ApiUsageService) *RouteHandler {
	if cacheDB == nil || routeClient == nil {
		// テスト用: モッククライアントを使用
		return &RouteHandler{
			routeService: service.NewCachedRouteService(
				service.NewMockRoutesClient(),
				&mockCacheStore{},
				0, // 無期限
			),
			apiUsageService: apiUsageService,
		}
	}

	cacheStore := repository.NewRouteCacheRepository(cacheDB)
	return &RouteHandler{
		routeService: service.NewCachedRouteService(
			routeClient,
			cacheStore,
			0, // 無期限
		),
		apiUsageService: apiUsageService,
	}
}

// RouteResponse ルート取得レスポンス
type RouteResponse struct {
	Success     bool    `json:"success"`
	Error       string  `json:"error,omitempty"`
	Origin      string  `json:"origin"`
	Dest        string  `json:"dest"`
	DistanceKm  float64 `json:"distance_km"`
	DurationMin int     `json:"duration_min"`
	FromCache   bool    `json:"from_cache"`
	// 運輸局・地区判定情報（出発地ベース）
	Prefecture string `json:"prefecture"`   // 都道府県
	RegionCode int    `json:"region_code"`  // 運輸局コード（1-10）
	RegionName string `json:"region_name"`  // 運輸局名
	AkabouArea string `json:"akabou_area"`  // 赤帽地区（東京23区/大阪市内/空文字）
}

// GetRoute ルート情報を取得するAPI
// GET /api/route?origin=東京&dest=大阪
func (h *RouteHandler) GetRoute(c echo.Context) error {
	origin := c.QueryParam("origin")
	dest := c.QueryParam("dest")

	// バリデーション
	if origin == "" {
		return c.JSON(http.StatusOK, &RouteResponse{
			Success: false,
			Error:   "出発地が指定されていません",
		})
	}
	if dest == "" {
		return c.JSON(http.StatusOK, &RouteResponse{
			Success: false,
			Error:   "目的地が指定されていません",
		})
	}
	if origin == dest {
		return c.JSON(http.StatusOK, &RouteResponse{
			Success: false,
			Error:   "出発地と目的地が同じです",
		})
	}

	// ルート情報を取得
	result, err := h.routeService.GetRoute(origin, dest)
	if err != nil {
		return c.JSON(http.StatusOK, &RouteResponse{
			Success: false,
			Error:   "ルート取得エラー: " + err.Error(),
		})
	}

	// キャッシュミス時（API呼び出し時）はAPI使用量をカウントアップ
	if !result.FromCache && h.apiUsageService != nil {
		_ = h.apiUsageService.IncrementAndCheck()
	}

	// 出発地から都道府県・運輸局情報を取得
	prefecture, regionCode, regionName, akabouArea := resolveRegionInfo(origin)

	return c.JSON(http.StatusOK, &RouteResponse{
		Success:     true,
		Origin:      result.Route.Origin,
		Dest:        result.Route.Dest,
		DistanceKm:  result.Route.DistanceKm,
		DurationMin: result.Route.DurationMin,
		FromCache:   result.FromCache,
		Prefecture:  prefecture,
		RegionCode:  regionCode,
		RegionName:  regionName,
		AkabouArea:  akabouArea,
	})
}

// resolveRegionInfo 住所から運輸局情報を取得
func resolveRegionInfo(address string) (prefecture string, regionCode int, regionName string, akabouArea string) {
	// 住所から都道府県を抽出
	pref, ok := service.ExtractPrefectureFromAddress(address)
	if !ok {
		return "", 0, "", ""
	}
	prefecture = pref

	// 都道府県から運輸局コードを取得
	code, err := service.ResolveRegionCode(prefecture)
	if err != nil {
		return prefecture, 0, "", ""
	}
	regionCode = code

	// 都道府県から運輸局名を取得
	name, err := service.ResolveRegionName(prefecture)
	if err != nil {
		return prefecture, regionCode, "", ""
	}
	regionName = name

	// 赤帽地区を判定
	akabouArea = service.ResolveAkabouArea(address)

	return prefecture, regionCode, regionName, akabouArea
}

// mockCacheStore テスト用のモックキャッシュストア
type mockCacheStore struct{}

func (s *mockCacheStore) Get(origin, dest string) (*model.RouteCache, error) {
	return nil, nil
}

func (s *mockCacheStore) Upsert(cache *model.RouteCache) error {
	return nil
}
