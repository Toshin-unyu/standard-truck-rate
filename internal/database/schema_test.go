package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// TestInitMainDB メインDB（str.db）のスキーマ初期化テスト
func TestInitMainDB(t *testing.T) {
	// テスト用一時ディレクトリ
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	// スキーマ初期化
	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	// 期待するテーブル一覧
	expectedTables := []string{
		"jta_time_base_fares",
		"jta_time_surcharges",
		"akabou_distance_fares",
		"akabou_time_fares",
		"akabou_surcharges",
		"akabou_area_surcharges",
		"akabou_additional_fees",
		"api_usage",
		"highway_ic_master",
	}

	// 各テーブルの存在確認
	for _, table := range expectedTables {
		if !tableExists(t, db, table) {
			t.Errorf("テーブル %s が存在しない", table)
		}
	}
}

// TestInitCacheDB キャッシュDB（cache.db）のスキーマ初期化テスト
func TestInitCacheDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cache.db")

	db, err := InitCacheDB(dbPath)
	if err != nil {
		t.Fatalf("InitCacheDB failed: %v", err)
	}
	defer db.Close()

	expectedTables := []string{
		"route_cache",
		"highway_toll_cache",
	}

	for _, table := range expectedTables {
		if !tableExists(t, db, table) {
			t.Errorf("テーブル %s が存在しない", table)
		}
	}
}

// TestJtaTimeBaseFaresSchema jta_time_base_faresテーブルのカラム確認
func TestJtaTimeBaseFaresSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":           "INTEGER",
		"region_code":  "INTEGER",
		"vehicle_code": "INTEGER",
		"hours":        "INTEGER",
		"base_km":      "INTEGER",
		"fare_yen":     "INTEGER",
	}

	checkTableColumns(t, db, "jta_time_base_fares", expectedColumns)
}

// TestJtaTimeSurchargesSchema jta_time_surchargesテーブルのカラム確認
func TestJtaTimeSurchargesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":             "INTEGER",
		"region_code":    "INTEGER",
		"vehicle_code":   "INTEGER",
		"surcharge_type": "TEXT",
		"fare_yen":       "INTEGER",
	}

	checkTableColumns(t, db, "jta_time_surcharges", expectedColumns)
}

// TestAkabouDistanceFaresSchema akabou_distance_faresテーブルのカラム確認
func TestAkabouDistanceFaresSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":          "INTEGER",
		"min_km":      "INTEGER",
		"max_km":      "INTEGER",
		"base_fare":   "INTEGER",
		"per_km_rate": "INTEGER",
	}

	checkTableColumns(t, db, "akabou_distance_fares", expectedColumns)
}

// TestAkabouTimeFaresSchema akabou_time_faresテーブルのカラム確認
func TestAkabouTimeFaresSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":            "INTEGER",
		"base_hours":    "INTEGER",
		"base_km":       "INTEGER",
		"base_fare":     "INTEGER",
		"overtime_rate": "INTEGER",
	}

	checkTableColumns(t, db, "akabou_time_fares", expectedColumns)
}

// TestAkabouSurchargesSchema akabou_surchargesテーブルのカラム確認
func TestAkabouSurchargesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":             "INTEGER",
		"surcharge_type": "TEXT",
		"rate_percent":   "INTEGER",
		"description":    "TEXT",
	}

	checkTableColumns(t, db, "akabou_surcharges", expectedColumns)
}

// TestAkabouAreaSurchargesSchema akabou_area_surchargesテーブルのカラム確認
func TestAkabouAreaSurchargesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":               "INTEGER",
		"area_name":        "TEXT",
		"surcharge_amount": "INTEGER",
	}

	checkTableColumns(t, db, "akabou_area_surcharges", expectedColumns)
}

// TestAkabouAdditionalFeesSchema akabou_additional_feesテーブルのカラム確認
func TestAkabouAdditionalFeesSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"id":           "INTEGER",
		"fee_type":     "TEXT",
		"free_minutes": "INTEGER",
		"unit_minutes": "INTEGER",
		"fee_amount":   "INTEGER",
	}

	checkTableColumns(t, db, "akabou_additional_fees", expectedColumns)
}

// TestApiUsageSchema api_usageテーブルのカラム確認
func TestApiUsageSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"year_month":    "TEXT",
		"request_count": "INTEGER",
		"limit_count":   "INTEGER",
		"last_updated":  "DATETIME",
	}

	checkTableColumns(t, db, "api_usage", expectedColumns)
}

// TestRouteCacheSchema route_cacheテーブルのカラム確認
func TestRouteCacheSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cache.db")

	db, err := InitCacheDB(dbPath)
	if err != nil {
		t.Fatalf("InitCacheDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"origin":       "TEXT",
		"dest":         "TEXT",
		"distance_km":  "REAL",
		"duration_min": "INTEGER",
		"created_at":   "DATETIME",
	}

	checkTableColumns(t, db, "route_cache", expectedColumns)
}

// TestHighwayICMasterSchema highway_ic_masterテーブルのカラム確認
func TestHighwayICMasterSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"code":       "TEXT",
		"name":       "TEXT",
		"yomi":       "TEXT",
		"type":       "INTEGER",
		"road_no":    "TEXT",
		"road_name":  "TEXT",
		"updated_at": "DATETIME",
	}

	checkTableColumns(t, db, "highway_ic_master", expectedColumns)
}

// TestHighwayTollCacheSchema highway_toll_cacheテーブルのカラム確認
func TestHighwayTollCacheSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "cache.db")

	db, err := InitCacheDB(dbPath)
	if err != nil {
		t.Fatalf("InitCacheDB failed: %v", err)
	}
	defer db.Close()

	expectedColumns := map[string]string{
		"origin_ic":    "TEXT",
		"dest_ic":      "TEXT",
		"car_type":     "INTEGER",
		"normal_toll":  "INTEGER",
		"etc_toll":     "INTEGER",
		"etc2_toll":    "INTEGER",
		"distance_km":  "REAL",
		"duration_min": "INTEGER",
		"created_at":   "DATETIME",
	}

	checkTableColumns(t, db, "highway_toll_cache", expectedColumns)
}

// TestInitMainDBIdempotent 複数回初期化しても問題ないことを確認
func TestInitMainDBIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	// 1回目
	db1, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("1回目のInitMainDB failed: %v", err)
	}
	db1.Close()

	// 2回目（エラーにならないこと）
	db2, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("2回目のInitMainDB failed: %v", err)
	}
	db2.Close()
}

// TestDBFileCreated DBファイルが作成されることを確認
func TestDBFileCreated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "str.db")

	db, err := InitMainDB(dbPath)
	if err != nil {
		t.Fatalf("InitMainDB failed: %v", err)
	}
	db.Close()

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("DBファイルが作成されていない")
	}
}

// ヘルパー関数

func tableExists(t *testing.T, db *sql.DB, tableName string) bool {
	t.Helper()
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name=?`
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	return err == nil
}

func checkTableColumns(t *testing.T, db *sql.DB, tableName string, expectedColumns map[string]string) {
	t.Helper()

	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		t.Fatalf("PRAGMA table_info failed: %v", err)
	}
	defer rows.Close()

	actualColumns := make(map[string]string)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("rows.Scan failed: %v", err)
		}
		actualColumns[name] = colType
	}

	for colName, expectedType := range expectedColumns {
		actualType, exists := actualColumns[colName]
		if !exists {
			t.Errorf("テーブル %s にカラム %s が存在しない", tableName, colName)
			continue
		}
		if actualType != expectedType {
			t.Errorf("テーブル %s のカラム %s: 期待=%s, 実際=%s", tableName, colName, expectedType, actualType)
		}
	}
}
