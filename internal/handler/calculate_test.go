package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// mockRenderer テスト用のモックレンダラー
type mockRenderer struct {
	lastTemplate string
	lastData     interface{}
}

func (r *mockRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	r.lastTemplate = name
	r.lastData = data
	w.Write([]byte("rendered: " + name))
	return nil
}

func TestCalculateHandler_Calculate(t *testing.T) {
	e := echo.New()
	renderer := &mockRenderer{}
	e.Renderer = renderer

	handler := NewCalculateHandler(nil)

	tests := []struct {
		name           string
		formData       url.Values
		wantStatusCode int
		wantTemplate   string
	}{
		{
			name: "正常な運賃計算リクエスト",
			formData: url.Values{
				"region_code":     {"3"},
				"vehicle_code":    {"3"},
				"distance_km":     {"100"},
				"driving_minutes": {"120"},
				"loading_minutes": {"60"},
			},
			wantStatusCode: http.StatusOK,
			wantTemplate:   "result",
		},
		{
			name: "深夜・休日割増あり",
			formData: url.Values{
				"region_code":     {"3"},
				"vehicle_code":    {"3"},
				"distance_km":     {"100"},
				"driving_minutes": {"120"},
				"loading_minutes": {"60"},
				"is_night":        {"true"},
				"is_holiday":      {"true"},
			},
			wantStatusCode: http.StatusOK,
			wantTemplate:   "result",
		},
		{
			name: "シンプル版基礎キロ使用",
			formData: url.Values{
				"region_code":        {"3"},
				"vehicle_code":       {"3"},
				"distance_km":        {"100"},
				"driving_minutes":    {"120"},
				"loading_minutes":    {"60"},
				"use_simple_base_km": {"true"},
			},
			wantStatusCode: http.StatusOK,
			wantTemplate:   "result",
		},
		{
			name: "赤帽地区指定あり",
			formData: url.Values{
				"region_code":     {"3"},
				"vehicle_code":    {"1"},
				"distance_km":     {"50"},
				"driving_minutes": {"60"},
				"loading_minutes": {"30"},
				"area":            {"東京23区"},
			},
			wantStatusCode: http.StatusOK,
			wantTemplate:   "result",
		},
		{
			name: "距離が未入力の場合エラー",
			formData: url.Values{
				"region_code":     {"3"},
				"vehicle_code":    {"3"},
				"driving_minutes": {"120"},
				"loading_minutes": {"60"},
			},
			wantStatusCode: http.StatusOK,
			wantTemplate:   "error",
		},
		{
			name: "距離が0以下の場合エラー",
			formData: url.Values{
				"region_code":     {"3"},
				"vehicle_code":    {"3"},
				"distance_km":     {"0"},
				"driving_minutes": {"120"},
				"loading_minutes": {"60"},
			},
			wantStatusCode: http.StatusOK,
			wantTemplate:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/fare/calculate",
				strings.NewReader(tt.formData.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Calculate(c)
			if err != nil {
				t.Errorf("Calculate() error = %v", err)
				return
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Calculate() status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			if renderer.lastTemplate != tt.wantTemplate {
				t.Errorf("Calculate() template = %v, want %v", renderer.lastTemplate, tt.wantTemplate)
			}
		})
	}
}

func TestCalculateHandler_CalculateJSON(t *testing.T) {
	e := echo.New()

	handler := NewCalculateHandler(nil)

	tests := []struct {
		name           string
		formData       url.Values
		wantStatusCode int
	}{
		{
			name: "JSON形式で運賃計算結果を取得",
			formData: url.Values{
				"region_code":     {"3"},
				"vehicle_code":    {"3"},
				"distance_km":     {"100"},
				"driving_minutes": {"120"},
				"loading_minutes": {"60"},
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/fare/calculate/json",
				strings.NewReader(tt.formData.Encode()))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.CalculateJSON(c)
			if err != nil {
				t.Errorf("CalculateJSON() error = %v", err)
				return
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("CalculateJSON() status = %v, want %v", rec.Code, tt.wantStatusCode)
			}
		})
	}
}
