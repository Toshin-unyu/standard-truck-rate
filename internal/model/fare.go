package model

// JtaTimeBaseFare トラ協時間制・基礎額
type JtaTimeBaseFare struct {
	ID          int64 `json:"id"`
	RegionCode  int   `json:"region_code"`  // 運輸局コード (1-10)
	VehicleCode int   `json:"vehicle_code"` // 車格コード (1-4)
	Hours       int   `json:"hours"`        // 時間制区分 (4 or 8)
	BaseKm      int   `json:"base_km"`      // 基礎走行キロ
	FareYen     int   `json:"fare_yen"`     // 運賃（円）
}

// JtaTimeSurcharge トラ協時間制・加算額
type JtaTimeSurcharge struct {
	ID            int64  `json:"id"`
	RegionCode    int    `json:"region_code"`    // 運輸局コード (1-10)
	VehicleCode   int    `json:"vehicle_code"`   // 車格コード (1-4)
	SurchargeType string `json:"surcharge_type"` // 加算種別 ("distance" or "time")
	FareYen       int    `json:"fare_yen"`       // 加算額（円）
}

// AkabouDistanceFare 赤帽距離制運賃
type AkabouDistanceFare struct {
	ID        int64  `json:"id"`
	MinKm     int    `json:"min_km"`      // 最小距離（km）
	MaxKm     *int   `json:"max_km"`      // 最大距離（km）、NULLの場合は上限なし
	BaseFare  *int   `json:"base_fare"`   // 基本運賃（円）
	PerKmRate *int   `json:"per_km_rate"` // 1kmあたり運賃（円）
}

// AkabouTimeFare 赤帽時間制運賃
type AkabouTimeFare struct {
	ID           int64 `json:"id"`
	BaseHours    int   `json:"base_hours"`    // 基本時間
	BaseKm       int   `json:"base_km"`       // 基本走行キロ
	BaseFare     int   `json:"base_fare"`     // 基本運賃（円）
	OvertimeRate int   `json:"overtime_rate"` // 超過料金（円/時間）
}

// AkabouSurcharge 赤帽割増料金
type AkabouSurcharge struct {
	ID            int64   `json:"id"`
	SurchargeType string  `json:"surcharge_type"` // 割増種別 ("holiday" or "night")
	RatePercent   int     `json:"rate_percent"`   // 割増率（%）
	Description   *string `json:"description"`    // 説明
}

// AkabouAreaSurcharge 赤帽地区割増
type AkabouAreaSurcharge struct {
	ID              int64  `json:"id"`
	AreaName        string `json:"area_name"`        // 地区名
	SurchargeAmount int    `json:"surcharge_amount"` // 割増額（円）
}

// AkabouAdditionalFee 赤帽付帯料金
type AkabouAdditionalFee struct {
	ID          int64  `json:"id"`
	FeeType     string `json:"fee_type"`     // 料金種別 ("work" or "waiting")
	FreeMinutes int    `json:"free_minutes"` // 無料時間（分）
	UnitMinutes int    `json:"unit_minutes"` // 単位時間（分）
	FeeAmount   int    `json:"fee_amount"`   // 料金（円）
}
