package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

// RouteHandler ルート情報ハンドラ
type RouteHandler struct {
	routeService *service.CachedRouteService
}

// NewRouteHandler 新しいRouteHandlerを作成
func NewRouteHandler(cacheDB *sql.DB, routeClient service.RouteClient) *RouteHandler {
	if cacheDB == nil || routeClient == nil {
		// テスト用: モッククライアントを使用
		return &RouteHandler{
			routeService: service.NewCachedRouteService(
				service.NewMockRoutesClient(),
				&mockCacheStore{},
				24*time.Hour,
			),
		}
	}

	cacheStore := repository.NewRouteCacheRepository(cacheDB)
	return &RouteHandler{
		routeService: service.NewCachedRouteService(
			routeClient,
			cacheStore,
			24*time.Hour,
		),
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
	route, err := h.routeService.GetRoute(origin, dest)
	if err != nil {
		return c.JSON(http.StatusOK, &RouteResponse{
			Success: false,
			Error:   "ルート取得エラー: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, &RouteResponse{
		Success:     true,
		Origin:      route.Origin,
		Dest:        route.Dest,
		DistanceKm:  route.DistanceKm,
		DurationMin: route.DurationMin,
		FromCache:   false,
	})
}

// mockCacheStore テスト用のモックキャッシュストア
type mockCacheStore struct{}

func (s *mockCacheStore) Get(origin, dest string) (*model.RouteCache, error) {
	return nil, nil
}

func (s *mockCacheStore) Upsert(cache *model.RouteCache) error {
	return nil
}
