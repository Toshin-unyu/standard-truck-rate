package model

import "time"

// RouteCache ルートキャッシュ（距離・時間）
type RouteCache struct {
	Origin      string    `json:"origin"`       // 出発地
	Dest        string    `json:"dest"`         // 目的地
	DistanceKm  float64   `json:"distance_km"`  // 距離（km）
	DurationMin int       `json:"duration_min"` // 所要時間（分）
	CreatedAt   time.Time `json:"created_at"`   // 作成日時
}
