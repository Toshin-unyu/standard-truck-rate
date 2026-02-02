package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

// mockApiUsageStore テスト用のモックストア
type mockApiUsageStore struct {
	usage *model.ApiUsage
	err   error
}

func (m *mockApiUsageStore) GetOrCreateCurrent() (*model.ApiUsage, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.usage, nil
}

func (m *mockApiUsageStore) IncrementCount(yearMonth string) error {
	return m.err
}

func TestApiUsageHandler_GetUsage(t *testing.T) {
	tests := []struct {
		name       string
		store      *mockApiUsageStore
		wantStatus int
		wantLevel  string
	}{
		{
			name: "正常に使用量を取得",
			store: &mockApiUsageStore{
				usage: &model.ApiUsage{
					YearMonth:    "2026-02",
					RequestCount: 100,
					LimitCount:   9000,
				},
			},
			wantStatus: http.StatusOK,
			wantLevel:  "ok",
		},
		{
			name: "警告レベル（80%以上）",
			store: &mockApiUsageStore{
				usage: &model.ApiUsage{
					YearMonth:    "2026-02",
					RequestCount: 7500,
					LimitCount:   9000,
				},
			},
			wantStatus: http.StatusOK,
			wantLevel:  "warning",
		},
		{
			name: "危険レベル（95%以上）",
			store: &mockApiUsageStore{
				usage: &model.ApiUsage{
					YearMonth:    "2026-02",
					RequestCount: 8600,
					LimitCount:   9000,
				},
			},
			wantStatus: http.StatusOK,
			wantLevel:  "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/usage", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			usageService := service.NewApiUsageService(tt.store)
			h := NewApiUsageHandler(usageService)

			if err := h.GetUsage(c); err != nil {
				t.Fatalf("GetUsage() error = %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
