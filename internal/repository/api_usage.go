package repository

import (
	"database/sql"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// ApiUsageRepository API使用量のリポジトリ
type ApiUsageRepository struct {
	db *sql.DB
}

// NewApiUsageRepository リポジトリを作成する
func NewApiUsageRepository(db *sql.DB) *ApiUsageRepository {
	return &ApiUsageRepository{db: db}
}

// Create API使用量を作成する
func (r *ApiUsageRepository) Create(usage *model.ApiUsage) error {
	_, err := r.db.Exec(`
		INSERT INTO api_usage (year_month, request_count, limit_count, last_updated)
		VALUES (?, ?, ?, ?)
	`, usage.YearMonth, usage.RequestCount, usage.LimitCount, time.Now())
	return err
}

// GetByYearMonth 年月でAPI使用量を取得する
func (r *ApiUsageRepository) GetByYearMonth(yearMonth string) (*model.ApiUsage, error) {
	usage := &model.ApiUsage{}
	err := r.db.QueryRow(`
		SELECT year_month, request_count, limit_count, last_updated
		FROM api_usage WHERE year_month = ?
	`, yearMonth).Scan(&usage.YearMonth, &usage.RequestCount, &usage.LimitCount, &usage.LastUpdated)
	if err != nil {
		return nil, err
	}
	return usage, nil
}

// GetCurrent 現在月のAPI使用量を取得する
func (r *ApiUsageRepository) GetCurrent() (*model.ApiUsage, error) {
	currentYearMonth := time.Now().Format("2006-01")
	return r.GetByYearMonth(currentYearMonth)
}

// GetOrCreateCurrent 現在月のAPI使用量を取得、なければ作成する
func (r *ApiUsageRepository) GetOrCreateCurrent() (*model.ApiUsage, error) {
	currentYearMonth := time.Now().Format("2006-01")

	usage, err := r.GetByYearMonth(currentYearMonth)
	if err == nil {
		return usage, nil
	}

	// 存在しない場合は作成
	if err == sql.ErrNoRows {
		newUsage := &model.ApiUsage{
			YearMonth:    currentYearMonth,
			RequestCount: 0,
			LimitCount:   9000, // デフォルト上限
		}
		if err := r.Create(newUsage); err != nil {
			return nil, err
		}
		return r.GetByYearMonth(currentYearMonth)
	}

	return nil, err
}

// IncrementCount 指定年月のリクエスト数を1増やす
func (r *ApiUsageRepository) IncrementCount(yearMonth string) error {
	_, err := r.db.Exec(`
		UPDATE api_usage
		SET request_count = request_count + 1, last_updated = ?
		WHERE year_month = ?
	`, time.Now(), yearMonth)
	return err
}

// Update API使用量を更新する
func (r *ApiUsageRepository) Update(usage *model.ApiUsage) error {
	_, err := r.db.Exec(`
		UPDATE api_usage
		SET request_count = ?, limit_count = ?, last_updated = ?
		WHERE year_month = ?
	`, usage.RequestCount, usage.LimitCount, time.Now(), usage.YearMonth)
	return err
}

// GetAll 全API使用量を取得する
func (r *ApiUsageRepository) GetAll() ([]*model.ApiUsage, error) {
	rows, err := r.db.Query(`
		SELECT year_month, request_count, limit_count, last_updated
		FROM api_usage ORDER BY year_month DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usages []*model.ApiUsage
	for rows.Next() {
		u := &model.ApiUsage{}
		if err := rows.Scan(&u.YearMonth, &u.RequestCount, &u.LimitCount, &u.LastUpdated); err != nil {
			return nil, err
		}
		usages = append(usages, u)
	}
	return usages, rows.Err()
}
