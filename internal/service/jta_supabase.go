package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// JtaSupabaseClient トラ協Supabase APIクライアント
type JtaSupabaseClient struct {
	baseURL    string
	anonKey    string
	httpClient *http.Client
}

// NewJtaSupabaseClient 新しいSupabaseクライアントを作成
func NewJtaSupabaseClient(baseURL, anonKey string) *JtaSupabaseClient {
	return &JtaSupabaseClient{
		baseURL: baseURL,
		anonKey: anonKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDistanceFare 距離制運賃を取得
// regionCode: 運輸局コード (1-10)
// vehicleCode: 車格コード (1-4)
// distanceKm: 距離 (km)
func (c *JtaSupabaseClient) GetDistanceFare(regionCode, vehicleCode, distanceKm int) (*model.JtaDistanceFare, error) {
	// 距離を丸める
	roundedKm := RoundDistance(distanceKm, regionCode)

	// クエリパラメータを構築
	endpoint := fmt.Sprintf("%s/rest/v1/fare_rates", c.baseURL)
	params := url.Values{}
	params.Set("region_code", fmt.Sprintf("eq.%d", regionCode))
	params.Set("vehicle_code", fmt.Sprintf("eq.%d", vehicleCode))
	params.Set("upto_km", fmt.Sprintf("eq.%d", roundedKm))

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API呼び出しエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APIエラー: ステータスコード %d", resp.StatusCode)
	}

	var fares []model.JtaDistanceFare
	if err := json.NewDecoder(resp.Body).Decode(&fares); err != nil {
		return nil, fmt.Errorf("JSONデコードエラー: %w", err)
	}

	if len(fares) == 0 {
		return nil, fmt.Errorf("運賃データが見つかりません: region=%d, vehicle=%d, distance=%d",
			regionCode, vehicleCode, roundedKm)
	}

	return &fares[0], nil
}

// GetChargeData 付帯料金データを全件取得
func (c *JtaSupabaseClient) GetChargeData() ([]model.JtaChargeData, error) {
	endpoint := fmt.Sprintf("%s/rest/v1/charge_data", c.baseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API呼び出しエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("APIエラー: ステータスコード %d", resp.StatusCode)
	}

	var charges []model.JtaChargeData
	if err := json.NewDecoder(resp.Body).Decode(&charges); err != nil {
		return nil, fmt.Errorf("JSONデコードエラー: %w", err)
	}

	return charges, nil
}

// setHeaders 共通ヘッダーを設定
func (c *JtaSupabaseClient) setHeaders(req *http.Request) {
	req.Header.Set("apikey", c.anonKey)
	req.Header.Set("Authorization", "Bearer "+c.anonKey)
	req.Header.Set("Content-Type", "application/json")
}

// RoundDistance 距離を運賃計算用に丸める
// 距離 ≤ 200km  → 10km単位で切り上げ
// 距離 ≤ 500km  → 20km単位で切り上げ
// 距離 > 500km  → 50km単位で切り上げ
// ※沖縄(regionCode=10)は特別ルールあり（未実装）
func RoundDistance(distanceKm, regionCode int) int {
	// TODO: 沖縄の特別ルール対応
	if regionCode == 10 {
		// 沖縄は5km, 10km区分（詳細要確認）
		// 暫定的に10km単位で切り上げ
		return roundUp(distanceKm, 10)
	}

	switch {
	case distanceKm <= 200:
		return roundUp(distanceKm, 10)
	case distanceKm <= 500:
		return roundUp(distanceKm, 20)
	default:
		return roundUp(distanceKm, 50)
	}
}

// roundUp 指定単位で切り上げ
func roundUp(value, unit int) int {
	if value%unit == 0 {
		return value
	}
	return ((value / unit) + 1) * unit
}
