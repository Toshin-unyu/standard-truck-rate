package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRouteHandler_GetRoute(t *testing.T) {
	e := echo.New()

	handler := NewRouteHandler(nil, nil)

	tests := []struct {
		name           string
		origin         string
		dest           string
		wantStatusCode int
		wantSuccess    bool
		wantError      string
	}{
		{
			name:           "正常なルート取得",
			origin:         "東京都千代田区",
			dest:           "大阪府大阪市",
			wantStatusCode: http.StatusOK,
			wantSuccess:    true,
		},
		{
			name:           "出発地が空の場合エラー",
			origin:         "",
			dest:           "大阪府大阪市",
			wantStatusCode: http.StatusOK,
			wantSuccess:    false,
			wantError:      "出発地が指定されていません",
		},
		{
			name:           "目的地が空の場合エラー",
			origin:         "東京都千代田区",
			dest:           "",
			wantStatusCode: http.StatusOK,
			wantSuccess:    false,
			wantError:      "目的地が指定されていません",
		},
		{
			name:           "出発地と目的地が同じ場合エラー",
			origin:         "東京都千代田区",
			dest:           "東京都千代田区",
			wantStatusCode: http.StatusOK,
			wantSuccess:    false,
			wantError:      "出発地と目的地が同じです",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/route?origin="+tt.origin+"&dest="+tt.dest, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.GetRoute(c)
			if err != nil {
				t.Errorf("GetRoute() error = %v", err)
				return
			}

			if rec.Code != tt.wantStatusCode {
				t.Errorf("GetRoute() status = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			var resp RouteResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Errorf("GetRoute() JSON parse error = %v", err)
				return
			}

			if resp.Success != tt.wantSuccess {
				t.Errorf("GetRoute() success = %v, want %v", resp.Success, tt.wantSuccess)
			}

			if tt.wantError != "" && resp.Error != tt.wantError {
				t.Errorf("GetRoute() error = %v, want %v", resp.Error, tt.wantError)
			}
		})
	}
}
