package service

import (
	"os"
	"testing"
)

func getTestClient(t *testing.T) *JtaSupabaseClient {
	// 環境変数またはハードコードされた値を使用
	url := os.Getenv("SUPABASE_URL")
	if url == "" {
		url = "https://pwnkpkeelpsxlvyxsaml.supabase.co"
	}
	key := os.Getenv("SUPABASE_ANON_KEY")
	if key == "" {
		key = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InB3bmtwa2VlbHBzeGx2eXhzYW1sIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDczNzc2ODAsImV4cCI6MjA2Mjk1MzY4MH0.wExTYvz9s-HBKKJr12miIyeuAgKKtrH7W8yo8RRhUbI"
	}
	return NewJtaSupabaseClient(url, key)
}

func TestNewJtaSupabaseClient(t *testing.T) {
	client := getTestClient(t)
	if client == nil {
		t.Fatal("クライアントの作成に失敗")
	}
}

func TestGetDistanceFare(t *testing.T) {
	client := getTestClient(t)

	tests := []struct {
		name        string
		regionCode  int
		vehicleCode int
		distanceKm  int
		wantErr     bool
	}{
		{
			name:        "北海道・小型車・10km",
			regionCode:  1,
			vehicleCode: 1,
			distanceKm:  10,
			wantErr:     false,
		},
		{
			name:        "関東・大型車・100km",
			regionCode:  3,
			vehicleCode: 3,
			distanceKm:  100,
			wantErr:     false,
		},
		{
			name:        "無効な運輸局コード",
			regionCode:  99,
			vehicleCode: 1,
			distanceKm:  10,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fare, err := client.GetDistanceFare(tt.regionCode, tt.vehicleCode, tt.distanceKm)
			if tt.wantErr {
				if err == nil {
					t.Error("エラーが期待されたが、発生しなかった")
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if fare == nil {
				t.Fatal("運賃データがnil")
			}
			if fare.FareYen <= 0 {
				t.Errorf("運賃が不正: %d", fare.FareYen)
			}
			t.Logf("取得成功: region=%d, vehicle=%d, distance=%d -> fare=%d円",
				tt.regionCode, tt.vehicleCode, tt.distanceKm, fare.FareYen)
		})
	}
}

func TestGetChargeData(t *testing.T) {
	client := getTestClient(t)

	charges, err := client.GetChargeData()
	if err != nil {
		t.Fatalf("付帯料金取得エラー: %v", err)
	}
	if len(charges) == 0 {
		t.Fatal("付帯料金データが0件")
	}
	t.Logf("付帯料金データ: %d件取得", len(charges))
}

func TestRoundDistance(t *testing.T) {
	tests := []struct {
		distance int
		region   int
		expected int
	}{
		// 通常地域（200km以下は10km単位）
		{5, 1, 10},
		{10, 1, 10},
		{15, 1, 20},
		{100, 1, 100},
		{105, 1, 110},
		// 通常地域（200km超500km以下は20km単位）
		{210, 1, 220},
		{300, 1, 300},
		{310, 1, 320},
		// 通常地域（500km超は50km単位）
		{510, 1, 550},
		{600, 1, 600},
		{620, 1, 650},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := RoundDistance(tt.distance, tt.region)
			if result != tt.expected {
				t.Errorf("RoundDistance(%d, %d) = %d, want %d",
					tt.distance, tt.region, result, tt.expected)
			}
		})
	}
}
