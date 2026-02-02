package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

// ApiUsageHandler API使用量ハンドラ
type ApiUsageHandler struct {
	usageService *service.ApiUsageService
}

// NewApiUsageHandler 新しいApiUsageHandlerを作成
func NewApiUsageHandler(usageService *service.ApiUsageService) *ApiUsageHandler {
	return &ApiUsageHandler{
		usageService: usageService,
	}
}

// GetUsage 現在の使用量統計を取得
// GET /api/usage
func (h *ApiUsageHandler) GetUsage(c echo.Context) error {
	stats, err := h.usageService.GetStats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "使用量統計の取得に失敗しました",
		})
	}

	return c.JSON(http.StatusOK, stats)
}
