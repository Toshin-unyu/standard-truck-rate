package repository

import (
	"database/sql"
	"time"

	"github.com/y-suzuki/standard-truck-rate/internal/model"
)

// RouteCacheRepository ルートキャッシュのリポジトリ
type RouteCacheRepository struct {
	db *sql.DB
}

// NewRouteCacheRepository リポジトリを作成する
func NewRouteCacheRepository(db *sql.DB) *RouteCacheRepository {
	return &RouteCacheRepository{db: db}
}

// Create ルートキャッシュを作成する
func (r *RouteCacheRepository) Create(cache *model.RouteCache) error {
	_, err := r.db.Exec(`
		INSERT INTO route_cache (origin, dest, distance_km, duration_min, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, cache.Origin, cache.Dest, cache.DistanceKm, cache.DurationMin, time.Now())
	return err
}

// Get origin/destでルートキャッシュを取得する
func (r *RouteCacheRepository) Get(origin, dest string) (*model.RouteCache, error) {
	cache := &model.RouteCache{}
	err := r.db.QueryRow(`
		SELECT origin, dest, distance_km, duration_min, created_at
		FROM route_cache WHERE origin = ? AND dest = ?
	`, origin, dest).Scan(&cache.Origin, &cache.Dest, &cache.DistanceKm, &cache.DurationMin, &cache.CreatedAt)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

// GetAll 全ルートキャッシュを取得する
func (r *RouteCacheRepository) GetAll() ([]*model.RouteCache, error) {
	rows, err := r.db.Query(`
		SELECT origin, dest, distance_km, duration_min, created_at
		FROM route_cache ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var caches []*model.RouteCache
	for rows.Next() {
		c := &model.RouteCache{}
		if err := rows.Scan(&c.Origin, &c.Dest, &c.DistanceKm, &c.DurationMin, &c.CreatedAt); err != nil {
			return nil, err
		}
		caches = append(caches, c)
	}
	return caches, rows.Err()
}

// Upsert ルートキャッシュを作成または更新する
func (r *RouteCacheRepository) Upsert(cache *model.RouteCache) error {
	_, err := r.db.Exec(`
		INSERT INTO route_cache (origin, dest, distance_km, duration_min, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(origin, dest) DO UPDATE SET
			distance_km = excluded.distance_km,
			duration_min = excluded.duration_min,
			created_at = excluded.created_at
	`, cache.Origin, cache.Dest, cache.DistanceKm, cache.DurationMin, time.Now())
	return err
}

// Delete ルートキャッシュを削除する
func (r *RouteCacheRepository) Delete(origin, dest string) error {
	_, err := r.db.Exec(`DELETE FROM route_cache WHERE origin = ? AND dest = ?`, origin, dest)
	return err
}

// Exists ルートキャッシュが存在するか確認する
func (r *RouteCacheRepository) Exists(origin, dest string) bool {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM route_cache WHERE origin = ? AND dest = ?
	`, origin, dest).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}
