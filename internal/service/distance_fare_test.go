package service

import (
	"testing"
)

// モック用のSupabaseクライアント
type mockSupabaseClient struct {
	fareYen int
	err     error
}

func (m *mockSupabaseClient) GetDistanceFareYen(regionCode, vehicleCode, distanceKm int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return m.fareYen, nil
}

func TestDistanceFareService_Calculate(t *testing.T) {
	tests := []struct {
		name           string
		regionCode     int
		vehicleCode    int
		distanceKm     int
		isNight        bool
		isHoliday      bool
		baseFare       int // モックの基本運賃
		wantTotalFare  int
		wantNightRate  float64
		wantHolidayRate float64
	}{
		{
			name:           "基本運賃のみ（割増なし）",
			regionCode:     3, // 関東
			vehicleCode:    3, // 大型
			distanceKm:     100,
			isNight:        false,
			isHoliday:      false,
			baseFare:       50000,
			wantTotalFare:  50000,
			wantNightRate:  1.0,
			wantHolidayRate: 1.0,
		},
		{
			name:           "深夜割増（3割増）",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			isNight:        true,
			isHoliday:      false,
			baseFare:       50000,
			wantTotalFare:  65000, // 50000 * 1.3
			wantNightRate:  1.3,
			wantHolidayRate: 1.0,
		},
		{
			name:           "休日割増（2割増）",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			isNight:        false,
			isHoliday:      true,
			baseFare:       50000,
			wantTotalFare:  60000, // 50000 * 1.2
			wantNightRate:  1.0,
			wantHolidayRate: 1.2,
		},
		{
			name:           "深夜＋休日割増（1.56倍）",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			isNight:        true,
			isHoliday:      true,
			baseFare:       50000,
			wantTotalFare:  78000, // 50000 * 1.3 * 1.2 = 78000
			wantNightRate:  1.3,
			wantHolidayRate: 1.2,
		},
		{
			name:           "端数の丸め確認",
			regionCode:     1, // 北海道
			vehicleCode:    1, // 小型
			distanceKm:     55, // → 60kmに丸められる
			isNight:        false,
			isHoliday:      false,
			baseFare:       30000,
			wantTotalFare:  30000,
			wantNightRate:  1.0,
			wantHolidayRate: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モッククライアントを使用
			mock := &mockSupabaseClient{fareYen: tt.baseFare}
			service := NewDistanceFareService(mock)

			result, err := service.Calculate(
				tt.regionCode,
				tt.vehicleCode,
				tt.distanceKm,
				tt.isNight,
				tt.isHoliday,
			)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}

			// 合計運賃の検証
			if result.TotalFare != tt.wantTotalFare {
				t.Errorf("TotalFare = %d, want %d", result.TotalFare, tt.wantTotalFare)
			}

			// 基本運賃の検証
			if result.BaseFare != tt.baseFare {
				t.Errorf("BaseFare = %d, want %d", result.BaseFare, tt.baseFare)
			}

			// 割増率の検証
			if result.NightRate != tt.wantNightRate {
				t.Errorf("NightRate = %f, want %f", result.NightRate, tt.wantNightRate)
			}
			if result.HolidayRate != tt.wantHolidayRate {
				t.Errorf("HolidayRate = %f, want %f", result.HolidayRate, tt.wantHolidayRate)
			}
		})
	}
}

func TestDistanceFareService_Calculate_Error(t *testing.T) {
	tests := []struct {
		name        string
		regionCode  int
		vehicleCode int
		distanceKm  int
	}{
		{
			name:        "無効な運輸局コード",
			regionCode:  99,
			vehicleCode: 1,
			distanceKm:  10,
		},
		{
			name:        "無効な車格コード",
			regionCode:  1,
			vehicleCode: 99,
			distanceKm:  10,
		},
		{
			name:        "無効な距離（0以下）",
			regionCode:  1,
			vehicleCode: 1,
			distanceKm:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSupabaseClient{fareYen: 10000}
			service := NewDistanceFareService(mock)

			_, err := service.Calculate(
				tt.regionCode,
				tt.vehicleCode,
				tt.distanceKm,
				false,
				false,
			)

			if err == nil {
				t.Error("エラーが期待されたが、発生しなかった")
			}
		})
	}
}

func TestDistanceFareResult_Breakdown(t *testing.T) {
	// 計算根拠の出力テスト
	mock := &mockSupabaseClient{fareYen: 50000}
	service := NewDistanceFareService(mock)

	result, err := service.Calculate(3, 3, 105, true, true)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	// 丸め後の距離が正しいか
	if result.RoundedKm != 110 { // 105km → 110km
		t.Errorf("RoundedKm = %d, want 110", result.RoundedKm)
	}

	// 割増額が正しいか
	// 基本運賃: 50000
	// 深夜割増額: 50000 * 0.3 = 15000
	// 休日割増額: (50000 + 15000) * 0.2 = 13000
	// 合計: 50000 + 15000 + 13000 = 78000
	expectedNightSurcharge := 15000
	expectedHolidaySurcharge := 13000

	if result.NightSurcharge != expectedNightSurcharge {
		t.Errorf("NightSurcharge = %d, want %d", result.NightSurcharge, expectedNightSurcharge)
	}
	if result.HolidaySurcharge != expectedHolidaySurcharge {
		t.Errorf("HolidaySurcharge = %d, want %d", result.HolidaySurcharge, expectedHolidaySurcharge)
	}
}
