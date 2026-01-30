package service

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// DrivePlazaClient ドラぷらAPIクライアント
type DrivePlazaClient struct {
	httpClient    *http.Client
	icSearchURL   string
	tollSearchURL string
}

// NewDrivePlazaClient 新しいドラぷらクライアントを作成
func NewDrivePlazaClient() *DrivePlazaClient {
	return &DrivePlazaClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		icSearchURL:   "https://www.driveplaza.com/community/icsearch_api.php",
		tollSearchURL: "https://www.driveplaza.com/dp/SearchQuick",
	}
}

// NexcoIC XMLパース用構造体
type NexcoIC struct {
	XMLName xml.Name `xml:"NexcoIC"`
	Items   []ICItem `xml:"IcItem"`
}

// ICItem XMLパース用構造体
type ICItem struct {
	Code     string `xml:"Code"`
	Name     string `xml:"Name"`
	Yomi     string `xml:"Yomi"`
	Type     int    `xml:"Type"`
	RoadNo   string `xml:"RoadNo"`
	RoadName string `xml:"RoadName"`
}

// FetchICList ドラぷらからICリストを取得する
func (c *DrivePlazaClient) FetchICList() ([]*model.HighwayIC, error) {
	// 全件取得（val_word=空文字）
	reqURL := fmt.Sprintf("%s?val_word=", c.icSearchURL)

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("IC検索APIエラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IC検索API HTTPエラー: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	return ParseICListXML(body)
}

// ParseICListXML XMLをパースしてICリストを返す
func ParseICListXML(data []byte) ([]*model.HighwayIC, error) {
	var nexcoIC NexcoIC
	if err := xml.Unmarshal(data, &nexcoIC); err != nil {
		return nil, fmt.Errorf("XMLパースエラー: %w", err)
	}

	ics := make([]*model.HighwayIC, len(nexcoIC.Items))
	for i, item := range nexcoIC.Items {
		ics[i] = &model.HighwayIC{
			Code:      item.Code,
			Name:      item.Name,
			Yomi:      item.Yomi,
			Type:      item.Type,
			RoadNo:    item.RoadNo,
			RoadName:  item.RoadName,
			UpdatedAt: time.Now(),
		}
	}

	return ics, nil
}

// FetchToll 高速料金を取得する
func (c *DrivePlazaClient) FetchToll(originIC, destIC string, carType int) (*model.HighwayToll, error) {
	// URLパラメータを構築
	params := url.Values{}
	params.Set("startPlaceKana", originIC)
	params.Set("arrivePlaceKana", destIC)
	params.Set("carType", strconv.Itoa(carType))

	reqURL := fmt.Sprintf("%s?%s", c.tollSearchURL, params.Encode())

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("料金検索エラー: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("料金検索HTTPエラー: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
	}

	toll, err := ParseTollHTML(body)
	if err != nil {
		return nil, err
	}

	toll.OriginIC = originIC
	toll.DestIC = destIC
	toll.CarType = carType

	return toll, nil
}

// ParseTollHTML HTMLをパースして料金情報を抽出する
func ParseTollHTML(data []byte) (*model.HighwayToll, error) {
	html := string(data)

	toll := &model.HighwayToll{
		CreatedAt: time.Now(),
	}

	// 通常料金: <dl class="li-price"><dt>通常料金</dt><dd><em>11,900</em>円</dd></dl>
	normalPattern := regexp.MustCompile(`<dt>通常料金</dt>\s*<dd><em>([0-9,]+)</em>円</dd>`)
	if matches := normalPattern.FindStringSubmatch(html); len(matches) > 1 {
		toll.NormalToll = parseTollAmount(matches[1])
	}

	// ETC料金: <span id="fee_etc1">11,900</span>
	etcPattern := regexp.MustCompile(`<span id="fee_etc1">([0-9,]+)</span>`)
	if matches := etcPattern.FindStringSubmatch(html); len(matches) > 1 {
		toll.EtcToll = parseTollAmount(matches[1])
	}

	// ETC2.0料金: <span id="fee_etc21">11,900</span>
	etc2Pattern := regexp.MustCompile(`<span id="fee_etc21">([0-9,]+)</span>`)
	if matches := etc2Pattern.FindStringSubmatch(html); len(matches) > 1 {
		toll.Etc2Toll = parseTollAmount(matches[1])
	}

	// 距離: <dl class="li-distance">...<dd>314.6km</dd></dl>
	distancePattern := regexp.MustCompile(`(?s)class="li-distance"[^>]*>.*?<dd>([0-9.]+)km</dd>`)
	if matches := distancePattern.FindStringSubmatch(html); len(matches) > 1 {
		toll.DistanceKm = parseDistance(matches[1] + "km")
	}

	// 所要時間: <dl class="li-normal">...<dd>3時間6分</dd></dl>
	durationPattern := regexp.MustCompile(`(?s)class="li-normal"[^>]*>.*?<dd>([^<]+)</dd>`)
	if matches := durationPattern.FindStringSubmatch(html); len(matches) > 1 {
		toll.DurationMin = parseDuration(matches[1])
	}

	// 料金が取得できなかった場合はエラー
	if toll.NormalToll == 0 && toll.EtcToll == 0 {
		return nil, fmt.Errorf("料金情報が見つかりません")
	}

	return toll, nil
}

// parseTollAmount "8,350円" -> 8350
func parseTollAmount(s string) int {
	// カンマと「円」を除去
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "円", "")
	s = strings.TrimSpace(s)

	amount, _ := strconv.Atoi(s)
	return amount
}

// parseDistance "325.5km" -> 325.5
func parseDistance(s string) float64 {
	s = strings.ReplaceAll(s, "km", "")
	s = strings.TrimSpace(s)

	distance, _ := strconv.ParseFloat(s, 64)
	return distance
}

// parseDuration "3時間30分" -> 210 (分)
func parseDuration(s string) int {
	s = strings.TrimSpace(s)

	hours := 0
	minutes := 0

	// 時間を抽出
	hourPattern := regexp.MustCompile(`(\d+)時間`)
	if matches := hourPattern.FindStringSubmatch(s); len(matches) > 1 {
		hours, _ = strconv.Atoi(matches[1])
	}

	// 分を抽出
	minPattern := regexp.MustCompile(`(\d+)分`)
	if matches := minPattern.FindStringSubmatch(s); len(matches) > 1 {
		minutes, _ = strconv.Atoi(matches[1])
	}

	return hours*60 + minutes
}
