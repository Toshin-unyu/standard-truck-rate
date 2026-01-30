package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// IndexHandler メイン画面ハンドラ
type IndexHandler struct{}

// NewIndexHandler 新しいIndexHandlerを作成
func NewIndexHandler() *IndexHandler {
	return &IndexHandler{}
}

// Index メイン画面を表示
// GET /
func (h *IndexHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}
