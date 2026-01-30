package service

import (
	"testing"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// モック用の時間制運賃取得インターフェース
type mockTimeFareGetter struct {
	baseFare       *model.JtaTimeBaseFare
	distSurcharge  *model.JtaTimeSurcharge
	timeSurcharge  *model.JtaTimeSurcharge
	err            error
}

func (m *mockTimeFareGetter) GetBaseFare(regionCode, vehicleCode, hours int) (*model.JtaTimeBaseFare, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.baseFare, nil
}

func (m *mockTimeFareGetter) GetSurcharge(regionCode, vehicleCode int, surchargeType string) (*model.JtaTimeSurcharge, error) {
	if m.err != nil {
		return nil, m.err
	}
	if surchargeType == "distance" {
		return m.distSurcharge, nil
	}
	return m.timeSurcharge, nil
}

func TestTimeFareService_Calculate(t *testing.T) {
	// 関東・大型車の基礎額と加算額（要件定義書より）
	// 8時間制: 基礎額60,090円、距離加算630円/10km、時間加算4,180円/時間
	// 4時間制: 基礎額36,050円

	tests := []struct {
		name           string
		regionCode     int
		vehicleCode    int
		distanceKm     int
		drivingMinutes int     // 走行時間（分）
		loadingMinutes int     // 荷役時間（分）
		isNight        bool
		isHoliday      bool
		baseFare       int     // 基礎額
		baseKm         int     // 基礎走行キロ
		baseHours      int     // 基礎時間
		distSurcharge  int     // 距離超過加算（10kmあたり）
		timeSurcharge  int     // 時間超過加算（1時間あたり）
		wantHours      int     // 適用される時間制（4 or 8）
		wantTotalFare  int
	}{
		{
			name:           "8時間制・超過なし（大型車）",
			regionCode:     3, // 関東
			vehicleCode:    3, // 大型
			distanceKm:     100,
			drivingMinutes: 240, // 4時間
			loadingMinutes: 60,  // 1時間 → 合計5時間 → 8時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       60090,
			baseKm:         130, // 大型車は130km
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  60090, // 超過なし
		},
		{
			name:           "4時間制・超過なし（大型車）",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     50,
			drivingMinutes: 120, // 2時間
			loadingMinutes: 60,  // 1時間 → 合計3時間 → 4時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       36050,
			baseKm:         60, // 大型車は60km
			baseHours:      4,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      4,
			wantTotalFare:  36050,
		},
		{
			name:           "8時間制・距離超過あり（100km超過）",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     230, // 130km + 100km超過
			drivingMinutes: 300, // 5時間
			loadingMinutes: 60,  // 1時間 → 合計6時間 → 8時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       60090,
			baseKm:         130,
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  66390, // 60090 + (100/10)*630 = 60090 + 6300
		},
		{
			name:           "8時間制・時間超過あり（2時間超過）",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			drivingMinutes: 480, // 8時間
			loadingMinutes: 120, // 2時間 → 合計10時間 → 8時間制、2時間超過
			isNight:        false,
			isHoliday:      false,
			baseFare:       60090,
			baseKm:         130,
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  68450, // 60090 + 2*4180 = 60090 + 8360
		},
		{
			name:           "8時間制・距離超過＋時間超過",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     230, // 100km超過
			drivingMinutes: 540, // 9時間
			loadingMinutes: 60,  // 1時間 → 合計10時間、2時間超過
			isNight:        false,
			isHoliday:      false,
			baseFare:       60090,
			baseKm:         130,
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  74750, // 60090 + 6300 + 8360
		},
		{
			name:           "8時間制・深夜割増",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			drivingMinutes: 300,
			loadingMinutes: 60,
			isNight:        true,
			isHoliday:      false,
			baseFare:       60090,
			baseKm:         130,
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  78117, // 60090 * 1.3 = 78117
		},
		{
			name:           "8時間制・休日割増",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			drivingMinutes: 300,
			loadingMinutes: 60,
			isNight:        false,
			isHoliday:      true,
			baseFare:       60090,
			baseKm:         130,
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  72108, // 60090 * 1.2 = 72108
		},
		{
			name:           "8時間制・深夜＋休日割増",
			regionCode:     3,
			vehicleCode:    3,
			distanceKm:     100,
			drivingMinutes: 300,
			loadingMinutes: 60,
			isNight:        true,
			isHoliday:      true,
			baseFare:       60090,
			baseKm:         130,
			baseHours:      8,
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantTotalFare:  93740, // 60090 * 1.3 * 1.2 = 93740.4 → 93740
		},
		{
			name:           "小型車・8時間制（基礎走行キロ100km）",
			regionCode:     3,
			vehicleCode:    1, // 小型車
			distanceKm:     150, // 50km超過
			drivingMinutes: 300,
			loadingMinutes: 60,
			isNight:        false,
			isHoliday:      false,
			baseFare:       39380,
			baseKm:         100, // 小型車は100km
			baseHours:      8,
			distSurcharge:  350,
			timeSurcharge:  3710,
			wantHours:      8,
			wantTotalFare:  41130, // 39380 + (50/10)*350 = 39380 + 1750
		},
		{
			name:           "小型車・4時間制（基礎走行キロ50km）",
			regionCode:     3,
			vehicleCode:    1, // 小型車
			distanceKm:     70, // 20km超過
			drivingMinutes: 120,
			loadingMinutes: 60, // 合計3時間 → 4時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       23630,
			baseKm:         50, // 小型車は50km
			baseHours:      4,
			distSurcharge:  350,
			timeSurcharge:  3710,
			wantHours:      4,
			wantTotalFare:  24330, // 23630 + (20/10)*350 = 23630 + 700
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTimeFareGetter{
				baseFare: &model.JtaTimeBaseFare{
					RegionCode:  tt.regionCode,
					VehicleCode: tt.vehicleCode,
					Hours:       tt.wantHours,
					BaseKm:      tt.baseKm,
					FareYen:     tt.baseFare,
				},
				distSurcharge: &model.JtaTimeSurcharge{
					RegionCode:    tt.regionCode,
					VehicleCode:   tt.vehicleCode,
					SurchargeType: "distance",
					FareYen:       tt.distSurcharge,
				},
				timeSurcharge: &model.JtaTimeSurcharge{
					RegionCode:    tt.regionCode,
					VehicleCode:   tt.vehicleCode,
					SurchargeType: "time",
					FareYen:       tt.timeSurcharge,
				},
			}

			service := NewTimeFareService(mock)

			result, err := service.Calculate(
				tt.regionCode,
				tt.vehicleCode,
				tt.distanceKm,
				tt.drivingMinutes,
				tt.loadingMinutes,
				tt.isNight,
				tt.isHoliday,
				false, // useSimpleBaseKm = false（トラ協PDF版）
			)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}

			// 適用時間制の検証
			if result.AppliedHours != tt.wantHours {
				t.Errorf("AppliedHours = %d, want %d", result.AppliedHours, tt.wantHours)
			}

			// 合計運賃の検証
			if result.TotalFare != tt.wantTotalFare {
				t.Errorf("TotalFare = %d, want %d", result.TotalFare, tt.wantTotalFare)
			}
		})
	}
}

func TestTimeFareService_Calculate_Error(t *testing.T) {
	tests := []struct {
		name           string
		regionCode     int
		vehicleCode    int
		distanceKm     int
		drivingMinutes int
		loadingMinutes int
	}{
		{
			name:           "無効な運輸局コード",
			regionCode:     99,
			vehicleCode:    1,
			distanceKm:     100,
			drivingMinutes: 180,
			loadingMinutes: 60,
		},
		{
			name:           "無効な車格コード",
			regionCode:     1,
			vehicleCode:    99,
			distanceKm:     100,
			drivingMinutes: 180,
			loadingMinutes: 60,
		},
		{
			name:           "無効な距離（0以下）",
			regionCode:     1,
			vehicleCode:    1,
			distanceKm:     0,
			drivingMinutes: 180,
			loadingMinutes: 60,
		},
		{
			name:           "無効な走行時間（0以下）",
			regionCode:     1,
			vehicleCode:    1,
			distanceKm:     100,
			drivingMinutes: 0,
			loadingMinutes: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTimeFareGetter{
				baseFare: &model.JtaTimeBaseFare{
					FareYen: 10000,
					BaseKm:  100,
				},
				distSurcharge: &model.JtaTimeSurcharge{FareYen: 100},
				timeSurcharge: &model.JtaTimeSurcharge{FareYen: 1000},
			}
			service := NewTimeFareService(mock)

			_, err := service.Calculate(
				tt.regionCode,
				tt.vehicleCode,
				tt.distanceKm,
				tt.drivingMinutes,
				tt.loadingMinutes,
				false,
				false,
				false, // useSimpleBaseKm
			)

			if err == nil {
				t.Error("エラーが期待されたが、発生しなかった")
			}
		})
	}
}

func TestTimeFareService_Calculate_SimpleBaseKm(t *testing.T) {
	// シンプル版: 4時間制=30km、8時間制=50km（車格に関係なく固定）
	tests := []struct {
		name           string
		regionCode     int
		vehicleCode    int
		distanceKm     int
		drivingMinutes int
		loadingMinutes int
		isNight        bool
		isHoliday      bool
		baseFare       int
		dbBaseKm       int // DBから返される値（トラ協PDF版）
		distSurcharge  int
		timeSurcharge  int
		wantHours      int
		wantBaseKm     int // シンプル版で使用される基礎走行キロ
		wantTotalFare  int
	}{
		{
			name:           "シンプル版・4時間制（基礎走行キロ30km固定）",
			regionCode:     3,
			vehicleCode:    3, // 大型車（トラ協PDF版なら60km）
			distanceKm:     50, // 30km + 20km超過
			drivingMinutes: 120,
			loadingMinutes: 60, // 合計3時間 → 4時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       36050,
			dbBaseKm:       60,  // DBの値（使用されない）
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      4,
			wantBaseKm:     30, // シンプル版
			wantTotalFare:  37310, // 36050 + (20/10)*630 = 36050 + 1260
		},
		{
			name:           "シンプル版・8時間制（基礎走行キロ50km固定）",
			regionCode:     3,
			vehicleCode:    3, // 大型車（トラ協PDF版なら130km）
			distanceKm:     100, // 50km + 50km超過
			drivingMinutes: 300,
			loadingMinutes: 60, // 合計6時間 → 8時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       60090,
			dbBaseKm:       130, // DBの値（使用されない）
			distSurcharge:  630,
			timeSurcharge:  4180,
			wantHours:      8,
			wantBaseKm:     50, // シンプル版
			wantTotalFare:  63240, // 60090 + (50/10)*630 = 60090 + 3150
		},
		{
			name:           "シンプル版・小型車も同じ基礎走行キロ",
			regionCode:     3,
			vehicleCode:    1, // 小型車（トラ協PDF版なら4時間制50km）
			distanceKm:     40, // 30km + 10km超過
			drivingMinutes: 120,
			loadingMinutes: 60, // 合計3時間 → 4時間制
			isNight:        false,
			isHoliday:      false,
			baseFare:       23630,
			dbBaseKm:       50,  // DBの値（使用されない）
			distSurcharge:  350,
			timeSurcharge:  3710,
			wantHours:      4,
			wantBaseKm:     30, // シンプル版（車格に関係なく固定）
			wantTotalFare:  23980, // 23630 + (10/10)*350 = 23630 + 350
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTimeFareGetter{
				baseFare: &model.JtaTimeBaseFare{
					RegionCode:  tt.regionCode,
					VehicleCode: tt.vehicleCode,
					Hours:       tt.wantHours,
					BaseKm:      tt.dbBaseKm, // DBから返される値
					FareYen:     tt.baseFare,
				},
				distSurcharge: &model.JtaTimeSurcharge{
					RegionCode:    tt.regionCode,
					VehicleCode:   tt.vehicleCode,
					SurchargeType: "distance",
					FareYen:       tt.distSurcharge,
				},
				timeSurcharge: &model.JtaTimeSurcharge{
					RegionCode:    tt.regionCode,
					VehicleCode:   tt.vehicleCode,
					SurchargeType: "time",
					FareYen:       tt.timeSurcharge,
				},
			}

			service := NewTimeFareService(mock)

			result, err := service.Calculate(
				tt.regionCode,
				tt.vehicleCode,
				tt.distanceKm,
				tt.drivingMinutes,
				tt.loadingMinutes,
				tt.isNight,
				tt.isHoliday,
				true, // useSimpleBaseKm = true
			)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}

			// 基礎走行キロの検証（シンプル版の値が使用されていること）
			if result.BaseKm != tt.wantBaseKm {
				t.Errorf("BaseKm = %d, want %d", result.BaseKm, tt.wantBaseKm)
			}

			// 合計運賃の検証
			if result.TotalFare != tt.wantTotalFare {
				t.Errorf("TotalFare = %d, want %d", result.TotalFare, tt.wantTotalFare)
			}
		})
	}
}

func TestDetermineHoursSystem(t *testing.T) {
	tests := []struct {
		name         string
		totalMinutes int
		wantHours    int
	}{
		{"3時間 → 4時間制", 180, 4},
		{"4時間ちょうど → 4時間制", 240, 4},
		{"4時間1分 → 8時間制", 241, 8},
		{"6時間 → 8時間制", 360, 8},
		{"8時間 → 8時間制", 480, 8},
		{"10時間 → 8時間制", 600, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineHoursSystem(tt.totalMinutes)
			if result != tt.wantHours {
				t.Errorf("DetermineHoursSystem(%d) = %d, want %d", tt.totalMinutes, result, tt.wantHours)
			}
		})
	}
}
