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

	handler := NewRouteHandler(nil, nil, nil)

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
			origin:         "東京都千代田区丸の内",
			dest:           "大阪府大阪市中央区",
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

func TestRouteHandler_GetRoute_WithRegionInfo(t *testing.T) {
	e := echo.New()

	handler := NewRouteHandler(nil, nil, nil)

	tests := []struct {
		name           string
		origin         string
		dest           string
		wantPrefecture string
		wantRegionCode int
		wantRegionName string
		wantAkabouArea string
	}{
		{
			name:           "東京23区からの出発",
			origin:         "東京都千代田区丸の内1-1-1",
			dest:           "大阪府大阪市中央区",
			wantPrefecture: "東京都",
			wantRegionCode: 3,
			wantRegionName: "関東",
			wantAkabouArea: "東京23区",
		},
		{
			name:           "大阪市内からの出発",
			origin:         "大阪府大阪市北区梅田",
			dest:           "東京都新宿区",
			wantPrefecture: "大阪府",
			wantRegionCode: 6,
			wantRegionName: "近畿",
			wantAkabouArea: "大阪市内",
		},
		{
			name:           "北海道からの出発",
			origin:         "北海道札幌市中央区",
			dest:           "東京都渋谷区",
			wantPrefecture: "北海道",
			wantRegionCode: 1,
			wantRegionName: "北海道",
			wantAkabouArea: "",
		},
		{
			name:           "沖縄からの出発",
			origin:         "沖縄県那覇市",
			dest:           "東京都港区",
			wantPrefecture: "沖縄県",
			wantRegionCode: 10,
			wantRegionName: "沖縄",
			wantAkabouArea: "",
		},
		{
			name:           "横浜からの出発（東京23区外）",
			origin:         "神奈川県横浜市中区",
			dest:           "東京都千代田区",
			wantPrefecture: "神奈川県",
			wantRegionCode: 3,
			wantRegionName: "関東",
			wantAkabouArea: "",
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

			var resp RouteResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Errorf("GetRoute() JSON parse error = %v", err)
				return
			}

			if !resp.Success {
				t.Errorf("GetRoute() success = false, error = %v", resp.Error)
				return
			}

			if resp.Prefecture != tt.wantPrefecture {
				t.Errorf("GetRoute() Prefecture = %v, want %v", resp.Prefecture, tt.wantPrefecture)
			}

			if resp.RegionCode != tt.wantRegionCode {
				t.Errorf("GetRoute() RegionCode = %v, want %v", resp.RegionCode, tt.wantRegionCode)
			}

			if resp.RegionName != tt.wantRegionName {
				t.Errorf("GetRoute() RegionName = %v, want %v", resp.RegionName, tt.wantRegionName)
			}

			if resp.AkabouArea != tt.wantAkabouArea {
				t.Errorf("GetRoute() AkabouArea = %v, want %v", resp.AkabouArea, tt.wantAkabouArea)
			}
		})
	}
}
