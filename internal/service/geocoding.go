package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// AddressComponents 住所の構成要素
type AddressComponents struct {
	Prefecture string // 都道府県
	City       string // 市区町村
	Address    string // 詳細住所
}

// GeocodingClient Geocoding APIクライアントインターフェース
type GeocodingClient interface {
	GetPrefecture(address string) (string, error)
	GetAddressComponents(address string) (*AddressComponents, error)
}

// GoogleGeocodingClient Google Maps Geocoding APIクライアント
type GoogleGeocodingClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewGoogleGeocodingClient 新しいGoogleGeocodingClientを作成
func NewGoogleGeocodingClient(apiKey string) *GoogleGeocodingClient {
	return &GoogleGeocodingClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://maps.googleapis.com/maps/api/geocode/json",
	}
}

// geocodingAPIResponse Geocoding API レスポンス構造体
type geocodingAPIResponse struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
	} `json:"results"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetPrefecture 住所から都道府県を取得
func (c *GoogleGeocodingClient) GetPrefecture(address string) (string, error) {
	components, err := c.GetAddressComponents(address)
	if err != nil {
		return "", err
	}
	return components.Prefecture, nil
}

// GetAddressComponents 住所から構成要素を取得
func (c *GoogleGeocodingClient) GetAddressComponents(address string) (*AddressComponents, error) {
	if address == "" {
		return nil, errors.New("住所が指定されていません")
	}

	if c.apiKey == "" {
		return nil, errors.New("Google Maps APIキーが設定されていません")
	}

	// URLを構築
	reqURL := fmt.Sprintf("%s?address=%s&key=%s&language=ja",
		c.baseURL,
		url.QueryEscape(address),
		c.apiKey,
	)

	// HTTPリクエスト送信
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("API呼び出しエラー: %w", err)
	}
	defer resp.Body.Close()

	// レスポンス読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	// レスポンスパース
	var apiResp geocodingAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("レスポンスJSONパースエラー: %w", err)
	}

	// ステータスチェック
	if apiResp.Status != "OK" {
		if apiResp.ErrorMessage != "" {
			return nil, fmt.Errorf("API エラー [%s]: %s", apiResp.Status, apiResp.ErrorMessage)
		}
		return nil, fmt.Errorf("API エラー: %s", apiResp.Status)
	}

	if len(apiResp.Results) == 0 {
		return nil, errors.New("住所が見つかりません")
	}

	// 住所構成要素を抽出
	components := &AddressComponents{}
	for _, comp := range apiResp.Results[0].AddressComponents {
		for _, t := range comp.Types {
			switch t {
			case "administrative_area_level_1":
				components.Prefecture = comp.LongName
			case "locality":
				components.City = comp.LongName
			case "sublocality_level_1", "ward":
				// 東京23区の場合、localityがないのでwardを使用
				if components.City == "" {
					components.City = comp.LongName
				}
			}
		}
	}

	components.Address = apiResp.Results[0].FormattedAddress

	return components, nil
}

// MockGeocodingClient モック用Geocodingクライアント
type MockGeocodingClient struct {
	mockData map[string]*AddressComponents
}

// NewMockGeocodingClient 新しいモッククライアントを作成
func NewMockGeocodingClient() *MockGeocodingClient {
	return &MockGeocodingClient{
		mockData: make(map[string]*AddressComponents),
	}
}

// GetPrefecture モック都道府県を返す
func (c *MockGeocodingClient) GetPrefecture(address string) (string, error) {
	if address == "" {
		return "", errors.New("住所が指定されていません")
	}

	// カスタムモックデータがあればそれを返す
	if data, ok := c.mockData[address]; ok {
		return data.Prefecture, nil
	}

	// 住所から都道府県を抽出（フォールバック）
	prefecture, ok := ExtractPrefectureFromAddress(address)
	if !ok {
		return "", errors.New("都道府県を特定できません: " + address)
	}

	return prefecture, nil
}

// GetAddressComponents モック住所構成要素を返す
func (c *MockGeocodingClient) GetAddressComponents(address string) (*AddressComponents, error) {
	if address == "" {
		return nil, errors.New("住所が指定されていません")
	}

	// カスタムモックデータがあればそれを返す
	if data, ok := c.mockData[address]; ok {
		return data, nil
	}

	// 住所から構成要素を抽出
	components := extractAddressComponents(address)
	if components.Prefecture == "" {
		return nil, errors.New("都道府県を特定できません: " + address)
	}

	return components, nil
}

// SetMockData モックデータを設定
func (c *MockGeocodingClient) SetMockData(address string, prefecture, city string) {
	c.mockData[address] = &AddressComponents{
		Prefecture: prefecture,
		City:       city,
		Address:    address,
	}
}

// ExtractPrefectureFromAddress 住所文字列から都道府県を抽出
func ExtractPrefectureFromAddress(address string) (string, bool) {
	if address == "" {
		return "", false
	}

	// 省略形から正式名称へのマッピング（先にチェック）
	shortNames := map[string]string{
		"東京":  "東京都",
		"大阪":  "大阪府",
		"京都":  "京都府",
		"北海道": "北海道",
	}

	// 省略形をチェック（住所に含まれていれば）
	for short, full := range shortNames {
		if strings.Contains(address, short) {
			return full, true
		}
	}

	// 都道府県のパターン（正式名称）
	prefecturePatterns := []string{
		"北海道",
		"東京都",
		"大阪府",
		"京都府",
	}

	// 都道府県（〜県）のパターン
	prefectures := []string{
		"青森県", "岩手県", "宮城県", "秋田県", "山形県", "福島県",
		"茨城県", "栃木県", "群馬県", "埼玉県", "千葉県", "神奈川県",
		"新潟県", "富山県", "石川県", "福井県", "山梨県", "長野県",
		"岐阜県", "静岡県", "愛知県", "三重県",
		"滋賀県", "兵庫県", "奈良県", "和歌山県",
		"鳥取県", "島根県", "岡山県", "広島県", "山口県",
		"徳島県", "香川県", "愛媛県", "高知県",
		"福岡県", "佐賀県", "長崎県", "熊本県", "大分県", "宮崎県", "鹿児島県",
		"沖縄県",
	}

	// 特殊な都道府県を先にチェック（正式名称）
	for _, pref := range prefecturePatterns {
		if strings.Contains(address, pref) {
			return pref, true
		}
	}

	// 一般的な県をチェック（正式名称）
	for _, pref := range prefectures {
		if strings.Contains(address, pref) {
			return pref, true
		}
	}

	// 県名の省略形もチェック（「県」なしでも検索）
	prefectureShortNames := map[string]string{
		"青森": "青森県", "岩手": "岩手県", "宮城": "宮城県", "秋田": "秋田県", "山形": "山形県", "福島": "福島県",
		"茨城": "茨城県", "栃木": "栃木県", "群馬": "群馬県", "埼玉": "埼玉県", "千葉": "千葉県", "神奈川": "神奈川県",
		"新潟": "新潟県", "富山": "富山県", "石川": "石川県", "福井": "福井県", "山梨": "山梨県", "長野": "長野県",
		"岐阜": "岐阜県", "静岡": "静岡県", "愛知": "愛知県", "三重": "三重県",
		"滋賀": "滋賀県", "兵庫": "兵庫県", "奈良": "奈良県", "和歌山": "和歌山県",
		"鳥取": "鳥取県", "島根": "島根県", "岡山": "岡山県", "広島": "広島県", "山口": "山口県",
		"徳島": "徳島県", "香川": "香川県", "愛媛": "愛媛県", "高知": "高知県",
		"福岡": "福岡県", "佐賀": "佐賀県", "長崎": "長崎県", "熊本": "熊本県", "大分": "大分県", "宮崎": "宮崎県", "鹿児島": "鹿児島県",
		"沖縄": "沖縄県",
	}

	for short, full := range prefectureShortNames {
		if strings.Contains(address, short) {
			return full, true
		}
	}

	return "", false
}

// extractAddressComponents 住所文字列から構成要素を抽出
func extractAddressComponents(address string) *AddressComponents {
	components := &AddressComponents{
		Address: address,
	}

	// 都道府県を抽出
	prefecture, ok := ExtractPrefectureFromAddress(address)
	if !ok {
		return components
	}
	components.Prefecture = prefecture

	// 都道府県以降の部分を取得
	remaining := strings.TrimPrefix(address, prefecture)

	// 市区町村を抽出
	// パターン: 〜市、〜区、〜町、〜村、〜郡
	cityPatterns := []struct {
		suffix string
		regex  *regexp.Regexp
	}{
		{"市", regexp.MustCompile(`^([^市]+市)`)},
		{"区", regexp.MustCompile(`^([^区]+区)`)},
		{"町", regexp.MustCompile(`^([^町]+町)`)},
		{"村", regexp.MustCompile(`^([^村]+村)`)},
	}

	for _, p := range cityPatterns {
		if matches := p.regex.FindStringSubmatch(remaining); len(matches) > 1 {
			components.City = matches[1]
			break
		}
	}

	return components
}
