package service

import (
	"errors"
	"strings"
)

// RegionInfo 運輸局情報
type RegionInfo struct {
	Code int
	Name string
}

// 都道府県 → 運輸局マッピング
var prefectureToRegion = map[string]RegionInfo{
	// 1: 北海道
	"北海道": {1, "北海道"},

	// 2: 東北
	"青森県": {2, "東北"},
	"岩手県": {2, "東北"},
	"宮城県": {2, "東北"},
	"秋田県": {2, "東北"},
	"山形県": {2, "東北"},
	"福島県": {2, "東北"},

	// 3: 関東
	"茨城県":  {3, "関東"},
	"栃木県":  {3, "関東"},
	"群馬県":  {3, "関東"},
	"埼玉県":  {3, "関東"},
	"千葉県":  {3, "関東"},
	"東京都":  {3, "関東"},
	"神奈川県": {3, "関東"},
	"山梨県":  {3, "関東"},

	// 4: 北陸信越
	"新潟県": {4, "北陸信越"},
	"富山県": {4, "北陸信越"},
	"石川県": {4, "北陸信越"},
	"長野県": {4, "北陸信越"},

	// 5: 中部
	"福井県": {5, "中部"},
	"岐阜県": {5, "中部"},
	"静岡県": {5, "中部"},
	"愛知県": {5, "中部"},
	"三重県": {5, "中部"},

	// 6: 近畿
	"滋賀県":  {6, "近畿"},
	"京都府":  {6, "近畿"},
	"大阪府":  {6, "近畿"},
	"兵庫県":  {6, "近畿"},
	"奈良県":  {6, "近畿"},
	"和歌山県": {6, "近畿"},

	// 7: 中国
	"鳥取県": {7, "中国"},
	"島根県": {7, "中国"},
	"岡山県": {7, "中国"},
	"広島県": {7, "中国"},
	"山口県": {7, "中国"},

	// 8: 四国
	"徳島県": {8, "四国"},
	"香川県": {8, "四国"},
	"愛媛県": {8, "四国"},
	"高知県": {8, "四国"},

	// 9: 九州
	"福岡県":  {9, "九州"},
	"佐賀県":  {9, "九州"},
	"長崎県":  {9, "九州"},
	"熊本県":  {9, "九州"},
	"大分県":  {9, "九州"},
	"宮崎県":  {9, "九州"},
	"鹿児島県": {9, "九州"},

	// 10: 沖縄
	"沖縄県": {10, "沖縄"},
}

// 東京23区の一覧
var tokyo23Wards = []string{
	"千代田区", "中央区", "港区", "新宿区", "文京区",
	"台東区", "墨田区", "江東区", "品川区", "目黒区",
	"大田区", "世田谷区", "渋谷区", "中野区", "杉並区",
	"豊島区", "北区", "荒川区", "板橋区", "練馬区",
	"足立区", "葛飾区", "江戸川区",
}

// ResolveRegionCode 都道府県名から運輸局コードを取得
func ResolveRegionCode(prefecture string) (int, error) {
	if prefecture == "" {
		return 0, errors.New("都道府県が指定されていません")
	}

	info, ok := prefectureToRegion[prefecture]
	if !ok {
		return 0, errors.New("不明な都道府県: " + prefecture)
	}

	return info.Code, nil
}

// ResolveRegionName 都道府県名から運輸局名を取得
func ResolveRegionName(prefecture string) (string, error) {
	if prefecture == "" {
		return "", errors.New("都道府県が指定されていません")
	}

	info, ok := prefectureToRegion[prefecture]
	if !ok {
		return "", errors.New("不明な都道府県: " + prefecture)
	}

	return info.Name, nil
}

// ResolveAkabouArea 住所から赤帽地区を判定
// 東京23区または大阪市内の場合、該当する地区名を返す
// それ以外の場合は空文字列を返す
func ResolveAkabouArea(address string) string {
	// 東京23区の判定
	if strings.Contains(address, "東京都") {
		for _, ward := range tokyo23Wards {
			if strings.Contains(address, ward) {
				return "東京23区"
			}
		}
	}

	// 大阪市内の判定
	if strings.Contains(address, "大阪府") && strings.Contains(address, "大阪市") {
		return "大阪市内"
	}

	return ""
}
