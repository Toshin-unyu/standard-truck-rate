package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitMainDB メインDB（str.db）を初期化し、全テーブルを作成する
func InitMainDB(dbPath string) (*sql.DB, error) {
	if err := ensureDir(dbPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := createMainTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// InitCacheDB キャッシュDB（cache.db）を初期化する
func InitCacheDB(dbPath string) (*sql.DB, error) {
	if err := ensureDir(dbPath); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := createCacheTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return os.MkdirAll(dir, 0755)
}

func createMainTables(db *sql.DB) error {
	schemas := []string{
		// トラ協時間制・基礎額
		`CREATE TABLE IF NOT EXISTS jta_time_base_fares (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			region_code INTEGER NOT NULL,
			vehicle_code INTEGER NOT NULL,
			hours INTEGER NOT NULL,
			base_km INTEGER NOT NULL,
			fare_yen INTEGER NOT NULL,
			UNIQUE(region_code, vehicle_code, hours)
		)`,

		// トラ協時間制・加算額
		`CREATE TABLE IF NOT EXISTS jta_time_surcharges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			region_code INTEGER NOT NULL,
			vehicle_code INTEGER NOT NULL,
			surcharge_type TEXT NOT NULL,
			fare_yen INTEGER NOT NULL,
			UNIQUE(region_code, vehicle_code, surcharge_type)
		)`,

		// 赤帽距離制運賃
		`CREATE TABLE IF NOT EXISTS akabou_distance_fares (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			min_km INTEGER NOT NULL,
			max_km INTEGER,
			base_fare INTEGER,
			per_km_rate INTEGER
		)`,

		// 赤帽時間制運賃
		`CREATE TABLE IF NOT EXISTS akabou_time_fares (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			base_hours INTEGER NOT NULL,
			base_km INTEGER NOT NULL,
			base_fare INTEGER NOT NULL,
			overtime_rate INTEGER NOT NULL
		)`,

		// 赤帽割増料金
		`CREATE TABLE IF NOT EXISTS akabou_surcharges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			surcharge_type TEXT NOT NULL UNIQUE,
			rate_percent INTEGER NOT NULL,
			description TEXT
		)`,

		// 赤帽地区割増
		`CREATE TABLE IF NOT EXISTS akabou_area_surcharges (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			area_name TEXT NOT NULL UNIQUE,
			surcharge_amount INTEGER NOT NULL
		)`,

		// 赤帽付帯料金
		`CREATE TABLE IF NOT EXISTS akabou_additional_fees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			fee_type TEXT NOT NULL UNIQUE,
			free_minutes INTEGER NOT NULL,
			unit_minutes INTEGER NOT NULL,
			fee_amount INTEGER NOT NULL
		)`,

		// API使用量
		`CREATE TABLE IF NOT EXISTS api_usage (
			year_month TEXT PRIMARY KEY,
			request_count INTEGER NOT NULL DEFAULT 0,
			limit_count INTEGER NOT NULL DEFAULT 9000,
			last_updated DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return err
		}
	}

	return nil
}

func createCacheTables(db *sql.DB) error {
	schema := `CREATE TABLE IF NOT EXISTS route_cache (
		origin TEXT NOT NULL,
		dest TEXT NOT NULL,
		distance_km REAL NOT NULL,
		duration_min INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (origin, dest)
	)`

	_, err := db.Exec(schema)
	return err
}
