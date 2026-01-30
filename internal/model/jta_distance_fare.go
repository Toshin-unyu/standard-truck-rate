package model

// JtaDistanceFare トラ協距離制運賃（Supabase fare_rates テーブル）
type JtaDistanceFare struct {
	RegionCode  int `json:"region_code"`  // 運輸局コード (1-10)
	VehicleCode int `json:"vehicle_code"` // 車格コード (1-4)
	UptoKm      int `json:"upto_km"`      // 距離上限 (km)
	FareYen     int `json:"fare_yen"`     // 運賃 (円)
}

// JtaChargeData トラ協付帯料金（Supabase charge_data テーブル）
type JtaChargeData struct {
	IDCode      int  `json:"id_code"`      // ID
	VehicleCode int  `json:"vehicle_code"` // 車格コード (1-4)
	TimeCode    int  `json:"time_code"`    // 時間コード
	ChargeYen   int  `json:"charge_yen"`   // 料金 (円)
	Per1MinYen  *int `json:"1m_yen"`       // 1分あたり料金 (円)、nullの場合あり
}
