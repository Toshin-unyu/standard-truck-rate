package repository

import (
	"database/sql"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// HighwayTollRepository 高速料金キャッシュのリポジトリ
type HighwayTollRepository struct {
	db *sql.DB
}

// NewHighwayTollRepository リポジトリを作成する
func NewHighwayTollRepository(db *sql.DB) *HighwayTollRepository {
	return &HighwayTollRepository{db: db}
}

// Create 高速料金キャッシュを作成する
func (r *HighwayTollRepository) Create(toll *model.HighwayToll) error {
	_, err := r.db.Exec(`
		INSERT INTO highway_toll_cache (origin_ic, dest_ic, car_type, normal_toll, etc_toll, etc2_toll, distance_km, duration_min, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, toll.OriginIC, toll.DestIC, toll.CarType, toll.NormalToll, toll.EtcToll, toll.Etc2Toll, toll.DistanceKm, toll.DurationMin, time.Now())
	return err
}

// Get 高速料金キャッシュを取得する
func (r *HighwayTollRepository) Get(originIC, destIC string, carType int) (*model.HighwayToll, error) {
	toll := &model.HighwayToll{}
	err := r.db.QueryRow(`
		SELECT origin_ic, dest_ic, car_type, normal_toll, etc_toll, etc2_toll, distance_km, duration_min, created_at
		FROM highway_toll_cache WHERE origin_ic = ? AND dest_ic = ? AND car_type = ?
	`, originIC, destIC, carType).Scan(
		&toll.OriginIC, &toll.DestIC, &toll.CarType,
		&toll.NormalToll, &toll.EtcToll, &toll.Etc2Toll,
		&toll.DistanceKm, &toll.DurationMin, &toll.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return toll, nil
}

// Upsert 高速料金キャッシュを作成または更新する
func (r *HighwayTollRepository) Upsert(toll *model.HighwayToll) error {
	_, err := r.db.Exec(`
		INSERT INTO highway_toll_cache (origin_ic, dest_ic, car_type, normal_toll, etc_toll, etc2_toll, distance_km, duration_min, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(origin_ic, dest_ic, car_type) DO UPDATE SET
			normal_toll = excluded.normal_toll,
			etc_toll = excluded.etc_toll,
			etc2_toll = excluded.etc2_toll,
			distance_km = excluded.distance_km,
			duration_min = excluded.duration_min,
			created_at = excluded.created_at
	`, toll.OriginIC, toll.DestIC, toll.CarType, toll.NormalToll, toll.EtcToll, toll.Etc2Toll, toll.DistanceKm, toll.DurationMin, time.Now())
	return err
}

// Delete 高速料金キャッシュを削除する
func (r *HighwayTollRepository) Delete(originIC, destIC string, carType int) error {
	_, err := r.db.Exec(`
		DELETE FROM highway_toll_cache WHERE origin_ic = ? AND dest_ic = ? AND car_type = ?
	`, originIC, destIC, carType)
	return err
}

// Exists 高速料金キャッシュが存在するか確認する
func (r *HighwayTollRepository) Exists(originIC, destIC string, carType int) bool {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM highway_toll_cache WHERE origin_ic = ? AND dest_ic = ? AND car_type = ?
	`, originIC, destIC, carType).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
