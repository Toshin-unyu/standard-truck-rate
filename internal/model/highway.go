package model

import "time"

// HighwayIC 高速道路ICマスタ
type HighwayIC struct {
	Code      string    `json:"code"`       // IC識別コード（7桁）
	Name      string    `json:"name"`       // IC名称
	Yomi      string    `json:"yomi"`       // 読み仮名（ひらがな）
	Type      int       `json:"type"`       // 種別（1=IC, 2=SA/PA）
	RoadNo    string    `json:"road_no"`    // 路線番号
	RoadName  string    `json:"road_name"`  // 路線名（ナンバリング付き）
	UpdatedAt time.Time `json:"updated_at"` // 最終更新日時
}

// HighwayToll 高速料金キャッシュ
type HighwayToll struct {
	OriginIC    string    `json:"origin_ic"`    // 出発IC名
	DestIC      string    `json:"dest_ic"`      // 到着IC名
	CarType     int       `json:"car_type"`     // 車種区分（0-4）
	NormalToll  int       `json:"normal_toll"`  // 通常料金（円）
	EtcToll     int       `json:"etc_toll"`     // ETC料金（円）
	Etc2Toll    int       `json:"etc2_toll"`    // ETC2.0料金（円）
	DistanceKm  float64   `json:"distance_km"`  // 高速道路距離（km）
	DurationMin int       `json:"duration_min"` // 高速道路所要時間（分）
	CreatedAt   time.Time `json:"created_at"`   // 作成日時
}

// CarType 車種区分
const (
	CarTypeLight   = 0 // 軽自動車等（軽貨物/赤帽）
	CarTypeNormal  = 1 // 普通車
	CarTypeMedium  = 2 // 中型車（4t）
	CarTypeLarge   = 3 // 大型車（10t）
	CarTypeSpecial = 4 // 特大車（トレーラー）
)

// ICType 種別
const (
	ICTypeIC   = 1 // IC
	ICTypeSAPA = 2 // SA/PA
)
