// 距離制運賃計算の動作確認用ツール
// 使用方法: go run cmd/tools/check_distance_fare/main.go
package main

import (
	"fmt"
	"os"

	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

func main() {
	// Supabaseクライアント作成
	url := os.Getenv("SUPABASE_URL")
	if url == "" {
		url = "https://pwnkpkeelpsxlvyxsaml.supabase.co"
	}
	key := os.Getenv("SUPABASE_ANON_KEY")
	if key == "" {
		key = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InB3bmtwa2VlbHBzeGx2eXhzYW1sIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NDczNzc2ODAsImV4cCI6MjA2Mjk1MzY4MH0.wExTYvz9s-HBKKJr12miIyeuAgKKtrH7W8yo8RRhUbI"
	}

	supabaseClient := service.NewJtaSupabaseClient(url, key)
	adapter := service.NewJtaSupabaseClientAdapter(supabaseClient)
	fareService := service.NewDistanceFareService(adapter)

	fmt.Println("========================================")
	fmt.Println("距離制運賃計算 動作確認")
	fmt.Println("========================================")

	// テストケース
	testCases := []struct {
		name        string
		regionCode  int
		vehicleCode int
		distanceKm  int
		isNight     bool
		isHoliday   bool
	}{
		{"関東・大型・100km（割増なし）", 3, 3, 100, false, false},
		{"関東・大型・100km（深夜）", 3, 3, 100, true, false},
		{"関東・大型・100km（休日）", 3, 3, 100, false, true},
		{"関東・大型・100km（深夜＋休日）", 3, 3, 100, true, true},
		{"北海道・小型・50km", 1, 1, 50, false, false},
		{"沖縄・中型・30km", 10, 2, 30, false, false},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.name)

		result, err := fareService.Calculate(
			tc.regionCode,
			tc.vehicleCode,
			tc.distanceKm,
			tc.isNight,
			tc.isHoliday,
		)

		if err != nil {
			fmt.Printf("エラー: %v\n", err)
			continue
		}

		fmt.Print(result.Breakdown())
	}

	fmt.Println("\n========================================")
	fmt.Println("確認完了")
	fmt.Println("========================================")
}
