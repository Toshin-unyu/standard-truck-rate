package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// RouteClient ルート情報を取得するクライアントインターフェース
type RouteClient interface {
	GetRoute(origin, dest string) (*model.RouteCache, error)
}

// RouteCacheStore キャッシュストアインターフェース
type RouteCacheStore interface {
	Get(origin, dest string) (*model.RouteCache, error)
	Upsert(cache *model.RouteCache) error
}

// GoogleRoutesClient Google Maps Routes APIクライアント
type GoogleRoutesClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewGoogleRoutesClient 新しいGoogleRoutesClientを作成
func NewGoogleRoutesClient(apiKey string) *GoogleRoutesClient {
	return &GoogleRoutesClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://routes.googleapis.com/directions/v2:computeRoutes",
	}
}

// routesAPIRequest Routes API リクエスト構造体
type routesAPIRequest struct {
	Origin                   routesWaypoint `json:"origin"`
	Destination              routesWaypoint `json:"destination"`
	TravelMode               string         `json:"travelMode"`
	RoutingPreference        string         `json:"routingPreference"`
	ComputeAlternativeRoutes bool           `json:"computeAlternativeRoutes"`
	LanguageCode             string         `json:"languageCode"`
	Units                    string         `json:"units"`
}

type routesWaypoint struct {
	Address string `json:"address"`
}

// routesAPIResponse Routes API レスポンス構造体
type routesAPIResponse struct {
	Routes []struct {
		DistanceMeters int    `json:"distanceMeters"`
		Duration       string `json:"duration"` // "3600s" 形式
	} `json:"routes"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// GetRoute Google Maps Routes APIを使用してルート情報を取得
func (c *GoogleRoutesClient) GetRoute(origin, dest string) (*model.RouteCache, error) {
	// バリデーション
	if err := validateRouteInput(origin, dest); err != nil {
		return nil, err
	}

	// APIキーチェック
	if c.apiKey == "" {
		return nil, errors.New("Google Maps APIキーが設定されていません")
	}

	// リクエスト構築
	reqBody := routesAPIRequest{
		Origin:                   routesWaypoint{Address: origin},
		Destination:              routesWaypoint{Address: dest},
		TravelMode:               "DRIVE",
		RoutingPreference:        "TRAFFIC_AWARE",
		ComputeAlternativeRoutes: false,
		LanguageCode:             "ja",
		Units:                    "METRIC",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("リクエストJSON作成エラー: %w", err)
	}

	// HTTPリクエスト作成
	req, err := http.NewRequest("POST", c.baseURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエスト作成エラー: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", "routes.duration,routes.distanceMeters")

	// リクエスト送信
	resp, err := c.httpClient.Do(req)
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
	var apiResp routesAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("レスポンスJSONパースエラー: %w", err)
	}

	// エラーチェック
	if apiResp.Error != nil {
		return nil, fmt.Errorf("API エラー [%s]: %s", apiResp.Error.Status, apiResp.Error.Message)
	}

	// ルートが見つからない場合
	if len(apiResp.Routes) == 0 {
		return nil, errors.New("ルートが見つかりません")
	}

	route := apiResp.Routes[0]

	// 距離をkmに変換
	distanceKm := float64(route.DistanceMeters) / 1000.0

	// 所要時間を分に変換（"3600s" -> 60分）
	durationMin := parseDurationSeconds(route.Duration)

	return &model.RouteCache{
		Origin:      origin,
		Dest:        dest,
		DistanceKm:  distanceKm,
		DurationMin: durationMin,
		CreatedAt:   time.Now(),
	}, nil
}

// parseDurationSeconds "3600s" 形式の文字列を分に変換
func parseDurationSeconds(duration string) int {
	duration = strings.TrimSuffix(duration, "s")
	var seconds int
	fmt.Sscanf(duration, "%d", &seconds)
	return seconds / 60
}

// MockRoutesClient モック用ルートクライアント
type MockRoutesClient struct {
	// モックデータ（固定値または動的に設定可能）
	mockData map[string]*model.RouteCache
}

// NewMockRoutesClient 新しいモッククライアントを作成
func NewMockRoutesClient() *MockRoutesClient {
	return &MockRoutesClient{
		mockData: make(map[string]*model.RouteCache),
	}
}

// GetRoute モックルート情報を返す
func (c *MockRoutesClient) GetRoute(origin, dest string) (*model.RouteCache, error) {
	// バリデーション
	if err := validateRouteInput(origin, dest); err != nil {
		return nil, err
	}

	// カスタムモックデータがあればそれを返す
	key := origin + "|" + dest
	if data, ok := c.mockData[key]; ok {
		return data, nil
	}

	// デフォルトのモックデータを生成
	// 住所から簡易的な距離を計算（実際のAPIの代わり）
	distanceKm := estimateDistance(origin, dest)
	durationMin := estimateDuration(distanceKm)

	return &model.RouteCache{
		Origin:      origin,
		Dest:        dest,
		DistanceKm:  distanceKm,
		DurationMin: durationMin,
		CreatedAt:   time.Now(),
	}, nil
}

// SetMockRoute モックデータを設定
func (c *MockRoutesClient) SetMockRoute(origin, dest string, distanceKm float64, durationMin int) {
	key := origin + "|" + dest
	c.mockData[key] = &model.RouteCache{
		Origin:      origin,
		Dest:        dest,
		DistanceKm:  distanceKm,
		DurationMin: durationMin,
		CreatedAt:   time.Now(),
	}
}

// estimateDistance 簡易的な距離推定（モック用）
func estimateDistance(origin, dest string) float64 {
	// 主要都市間の概算距離
	distances := map[string]float64{
		"東京|大阪":  500.0,
		"東京|名古屋": 350.0,
		"東京|福岡":  1000.0,
		"大阪|福岡":  600.0,
		"名古屋|大阪": 180.0,
	}

	// 都市名を抽出して検索
	originCity := extractCity(origin)
	destCity := extractCity(dest)

	key := originCity + "|" + destCity
	if d, ok := distances[key]; ok {
		return d
	}

	// 逆方向もチェック
	reverseKey := destCity + "|" + originCity
	if d, ok := distances[reverseKey]; ok {
		return d
	}

	// デフォルト値
	return 100.0
}

// extractCity 住所から都市名を抽出
func extractCity(address string) string {
	cities := []string{"東京", "大阪", "名古屋", "福岡", "札幌", "仙台", "横浜", "神戸", "京都", "広島"}
	for _, city := range cities {
		if strings.Contains(address, city) {
			return city
		}
	}
	return ""
}

// estimateDuration 距離から所要時間を推定（平均時速60km想定）
func estimateDuration(distanceKm float64) int {
	// 平均時速60kmで計算
	hours := distanceKm / 60.0
	return int(hours * 60) // 分に変換
}

// validateRouteInput 入力バリデーション
func validateRouteInput(origin, dest string) error {
	if origin == "" {
		return errors.New("出発地が指定されていません")
	}
	if dest == "" {
		return errors.New("目的地が指定されていません")
	}
	if origin == dest {
		return errors.New("出発地と目的地が同じです")
	}
	return nil
}

// CachedRouteService キャッシュ付きルートサービス
type CachedRouteService struct {
	client   RouteClient
	store    RouteCacheStore
	cacheTTL time.Duration
}

// NewCachedRouteService 新しいキャッシュ付きルートサービスを作成
func NewCachedRouteService(client RouteClient, store RouteCacheStore, cacheTTL time.Duration) *CachedRouteService {
	return &CachedRouteService{
		client:   client,
		store:    store,
		cacheTTL: cacheTTL,
	}
}

// GetRoute キャッシュを確認し、なければAPIから取得
func (s *CachedRouteService) GetRoute(origin, dest string) (*model.RouteCache, error) {
	// バリデーション
	if err := validateRouteInput(origin, dest); err != nil {
		return nil, err
	}

	// キャッシュを確認
	cached, err := s.store.Get(origin, dest)
	if err == nil && cached != nil {
		// キャッシュの有効期限をチェック
		if time.Since(cached.CreatedAt) < s.cacheTTL {
			return cached, nil
		}
	}

	// APIから取得
	route, err := s.client.GetRoute(origin, dest)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	if err := s.store.Upsert(route); err != nil {
		// キャッシュ保存エラーは無視してルート情報を返す
		// ログに記録するのが望ましいが、ここでは省略
	}

	return route, nil
}
