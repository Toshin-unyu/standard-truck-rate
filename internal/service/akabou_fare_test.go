package service

import (
	"testing"
)

// =============================================================================
// 距離制運賃テスト
// =============================================================================

func TestAkabouDistanceFare_Basic(t *testing.T) {
	service := NewAkabouFareService()

	tests := []struct {
		name       string
		distanceKm int
		wantFare   int
	}{
		// 基本料金（20km以内）
		{"1km", 1, 5500},
		{"10km", 10, 5500},
		{"20km", 20, 5500},

		// 21-50km区間（+242円/km）
		{"21km", 21, 5500 + 242*1},       // 5,742
		{"30km", 30, 5500 + 242*10},      // 7,920
		{"50km", 50, 5500 + 242*30},      // 12,760

		// 51-100km区間（+187円/km）
		{"51km", 51, 5500 + 242*30 + 187*1},   // 12,947
		{"75km", 75, 5500 + 242*30 + 187*25},  // 17,435
		{"100km", 100, 5500 + 242*30 + 187*50}, // 22,110

		// 101-150km区間（+154円/km）
		{"101km", 101, 5500 + 242*30 + 187*50 + 154*1},   // 22,264
		{"125km", 125, 5500 + 242*30 + 187*50 + 154*25},  // 25,960
		{"150km", 150, 5500 + 242*30 + 187*50 + 154*50},  // 29,810

		// 151km以上区間（+132円/km）
		{"151km", 151, 5500 + 242*30 + 187*50 + 154*50 + 132*1},   // 29,942
		{"200km", 200, 5500 + 242*30 + 187*50 + 154*50 + 132*50},  // 36,410
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateDistanceFare(tt.distanceKm, false, false, "")
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if result.TotalFare != tt.wantFare {
				t.Errorf("距離 %dkm: 期待値 %d円, 実際 %d円", tt.distanceKm, tt.wantFare, result.TotalFare)
			}
		})
	}
}

func TestAkabouDistanceFare_WithSurcharge(t *testing.T) {
	service := NewAkabouFareService()

	// 基本料金 5,500円 で検証
	baseFare := 5500

	tests := []struct {
		name      string
		isNight   bool
		isHoliday bool
		wantFare  int
	}{
		// 割増なし
		{"割増なし", false, false, baseFare},

		// 深夜割増（+30%）
		{"深夜のみ", true, false, int(float64(baseFare) * 1.3)}, // 7,150

		// 休日割増（+20%）
		{"休日のみ", false, true, int(float64(baseFare) * 1.2)}, // 6,600

		// 深夜+休日（深夜適用後に休日適用）
		{"深夜+休日", true, true, int(float64(int(float64(baseFare)*1.3)) * 1.2)}, // 8,580
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateDistanceFare(10, tt.isNight, tt.isHoliday, "")
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if result.TotalFare != tt.wantFare {
				t.Errorf("%s: 期待値 %d円, 実際 %d円", tt.name, tt.wantFare, result.TotalFare)
			}
		})
	}
}

func TestAkabouDistanceFare_WithAreaSurcharge(t *testing.T) {
	service := NewAkabouFareService()

	baseFare := 5500
	areaSurcharge := 440

	tests := []struct {
		name     string
		area     string
		wantFare int
	}{
		{"地区割増なし", "", baseFare},
		{"東京23区", "東京23区", baseFare + areaSurcharge},
		{"大阪市内", "大阪市内", baseFare + areaSurcharge},
		{"その他地域", "横浜市", baseFare}, // 対象外
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateDistanceFare(10, false, false, tt.area)
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if result.TotalFare != tt.wantFare {
				t.Errorf("%s: 期待値 %d円, 実際 %d円", tt.name, tt.wantFare, result.TotalFare)
			}
		})
	}
}

func TestAkabouDistanceFare_Invalid(t *testing.T) {
	service := NewAkabouFareService()

	tests := []struct {
		name       string
		distanceKm int
	}{
		{"0km", 0},
		{"負の距離", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CalculateDistanceFare(tt.distanceKm, false, false, "")
			if err == nil {
				t.Errorf("%s: エラーが期待されたが、nilが返された", tt.name)
			}
		})
	}
}

// =============================================================================
// 時間制運賃テスト
// =============================================================================

func TestAkabouTimeFare_Basic(t *testing.T) {
	service := NewAkabouFareService()

	// 基本料金: 6,050円（2時間・20km迄）
	// 超過30分ごと: +1,375円

	tests := []struct {
		name        string
		durationMin int // 作業時間（分）
		wantFare    int
	}{
		// 基本料金内（2時間=120分以内）
		{"30分", 30, 6050},
		{"60分", 60, 6050},
		{"90分", 90, 6050},
		{"120分", 120, 6050},

		// 超過（30分単位で切り上げ）
		{"121分（+30分超過）", 121, 6050 + 1375},       // 7,425
		{"150分（+30分超過）", 150, 6050 + 1375},       // 7,425
		{"151分（+60分超過）", 151, 6050 + 1375*2},     // 8,800
		{"180分（+60分超過）", 180, 6050 + 1375*2},     // 8,800
		{"240分（+120分超過）", 240, 6050 + 1375*4},    // 11,550
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateTimeFare(tt.durationMin, false, false, "")
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if result.TotalFare != tt.wantFare {
				t.Errorf("%s: 期待値 %d円, 実際 %d円", tt.name, tt.wantFare, result.TotalFare)
			}
		})
	}
}

func TestAkabouTimeFare_WithSurcharge(t *testing.T) {
	service := NewAkabouFareService()

	// 基本料金 6,050円 で検証
	baseFare := 6050

	tests := []struct {
		name      string
		isNight   bool
		isHoliday bool
		wantFare  int
	}{
		// 割増なし
		{"割増なし", false, false, baseFare},

		// 深夜割増（+30%）
		{"深夜のみ", true, false, int(float64(baseFare) * 1.3)}, // 7,865

		// 休日割増（+20%）
		{"休日のみ", false, true, int(float64(baseFare) * 1.2)}, // 7,260

		// 深夜+休日
		{"深夜+休日", true, true, int(float64(int(float64(baseFare)*1.3)) * 1.2)}, // 9,438
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateTimeFare(60, tt.isNight, tt.isHoliday, "")
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if result.TotalFare != tt.wantFare {
				t.Errorf("%s: 期待値 %d円, 実際 %d円", tt.name, tt.wantFare, result.TotalFare)
			}
		})
	}
}

func TestAkabouTimeFare_WithAreaSurcharge(t *testing.T) {
	service := NewAkabouFareService()

	baseFare := 6050
	areaSurcharge := 440

	tests := []struct {
		name     string
		area     string
		wantFare int
	}{
		{"地区割増なし", "", baseFare},
		{"東京23区", "東京23区", baseFare + areaSurcharge},
		{"大阪市内", "大阪市内", baseFare + areaSurcharge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateTimeFare(60, false, false, tt.area)
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if result.TotalFare != tt.wantFare {
				t.Errorf("%s: 期待値 %d円, 実際 %d円", tt.name, tt.wantFare, result.TotalFare)
			}
		})
	}
}

func TestAkabouTimeFare_Invalid(t *testing.T) {
	service := NewAkabouFareService()

	tests := []struct {
		name        string
		durationMin int
	}{
		{"0分", 0},
		{"負の時間", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CalculateTimeFare(tt.durationMin, false, false, "")
			if err == nil {
				t.Errorf("%s: エラーが期待されたが、nilが返された", tt.name)
			}
		})
	}
}

// =============================================================================
// 計算根拠（Breakdown）テスト
// =============================================================================

func TestAkabouDistanceFareResult_Breakdown(t *testing.T) {
	service := NewAkabouFareService()

	result, err := service.CalculateDistanceFare(100, true, false, "東京23区")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	breakdown := result.Breakdown()
	if breakdown == "" {
		t.Error("Breakdownが空文字列")
	}

	// 必要な情報が含まれているか確認
	if !contains(breakdown, "100km") {
		t.Error("距離情報が含まれていない")
	}
	if !contains(breakdown, "深夜") {
		t.Error("深夜割増情報が含まれていない")
	}
	if !contains(breakdown, "東京23区") {
		t.Error("地区割増情報が含まれていない")
	}
}

func TestAkabouTimeFareResult_Breakdown(t *testing.T) {
	service := NewAkabouFareService()

	result, err := service.CalculateTimeFare(180, false, true, "大阪市内")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	breakdown := result.Breakdown()
	if breakdown == "" {
		t.Error("Breakdownが空文字列")
	}

	// 必要な情報が含まれているか確認
	if !contains(breakdown, "180分") || !contains(breakdown, "3時間") {
		t.Error("時間情報が含まれていない")
	}
	if !contains(breakdown, "休日") {
		t.Error("休日割増情報が含まれていない")
	}
}

// ヘルパー関数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
