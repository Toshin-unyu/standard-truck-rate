package model

import "time"

// ApiUsage API使用量
type ApiUsage struct {
	YearMonth    string    `json:"year_month"`    // 年月 (YYYY-MM形式)
	RequestCount int       `json:"request_count"` // リクエスト数
	LimitCount   int       `json:"limit_count"`   // 上限数（デフォルト9000）
	LastUpdated  time.Time `json:"last_updated"`  // 最終更新日時
}

// UsagePercent 使用率を計算（%）
func (a *ApiUsage) UsagePercent() float64 {
	if a.LimitCount == 0 {
		return 0
	}
	return float64(a.RequestCount) / float64(a.LimitCount) * 100
}

// IsWarning 警告レベルか（80-94%）
func (a *ApiUsage) IsWarning() bool {
	p := a.UsagePercent()
	return p >= 80 && p < 95
}

// IsCritical 危険レベルか（95%以上）
func (a *ApiUsage) IsCritical() bool {
	return a.UsagePercent() >= 95
}
