package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestIndexHandler_Index(t *testing.T) {
	e := echo.New()

	// テスト用テンプレートレンダラーを設定
	e.Renderer = &mockRenderer{}

	handler := NewIndexHandler()

	tests := []struct {
		name           string
		wantStatusCode int
		wantTemplate   string
	}{
		{
			name:           "メイン画面を表示できる",
			wantStatusCode: http.StatusOK,
			wantTemplate:   "index.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Index(c)
			if err != nil {
				t.Errorf("Index() error = %v", err)
				return
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Index() status = %v, want %v", rec.Code, tt.wantStatusCode)
			}
		})
	}
}
