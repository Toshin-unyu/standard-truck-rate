package repository

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDB(t *testing.T) {
	// テスト用一時ディレクトリ
	tmpDir := t.TempDir()
	mainDBPath := filepath.Join(tmpDir, "test_main.db")
	cacheDBPath := filepath.Join(tmpDir, "test_cache.db")

	db, err := NewDB(mainDBPath, cacheDBPath)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	// メインDBが作成されていること
	if _, err := os.Stat(mainDBPath); os.IsNotExist(err) {
		t.Error("メインDBファイルが作成されていない")
	}

	// キャッシュDBが作成されていること
	if _, err := os.Stat(cacheDBPath); os.IsNotExist(err) {
		t.Error("キャッシュDBファイルが作成されていない")
	}
}

func TestDB_MainDB(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDB(
		filepath.Join(tmpDir, "main.db"),
		filepath.Join(tmpDir, "cache.db"),
	)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	mainDB := db.MainDB()
	if mainDB == nil {
		t.Error("MainDB() returned nil")
	}

	// pingで接続確認
	if err := mainDB.Ping(); err != nil {
		t.Errorf("MainDB Ping() error = %v", err)
	}
}

func TestDB_CacheDB(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDB(
		filepath.Join(tmpDir, "main.db"),
		filepath.Join(tmpDir, "cache.db"),
	)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	cacheDB := db.CacheDB()
	if cacheDB == nil {
		t.Error("CacheDB() returned nil")
	}

	// pingで接続確認
	if err := cacheDB.Ping(); err != nil {
		t.Errorf("CacheDB Ping() error = %v", err)
	}
}

func TestDB_Close(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDB(
		filepath.Join(tmpDir, "main.db"),
		filepath.Join(tmpDir, "cache.db"),
	)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}

	// Close後はPingがエラーになること
	if err := db.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if err := db.MainDB().Ping(); err == nil {
		t.Error("Close後もMainDBがPingできてしまう")
	}
}

func TestDB_TablesExist(t *testing.T) {
	tmpDir := t.TempDir()
	db, err := NewDB(
		filepath.Join(tmpDir, "main.db"),
		filepath.Join(tmpDir, "cache.db"),
	)
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	// メインDBのテーブル確認
	mainTables := []string{
		"jta_time_base_fares",
		"jta_time_surcharges",
		"akabou_distance_fares",
		"akabou_time_fares",
		"akabou_surcharges",
		"akabou_area_surcharges",
		"akabou_additional_fees",
		"api_usage",
	}

	for _, table := range mainTables {
		var name string
		err := db.MainDB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		if err != nil {
			t.Errorf("メインDBにテーブル %s が存在しない: %v", table, err)
		}
	}

	// キャッシュDBのテーブル確認
	var name string
	err = db.CacheDB().QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='route_cache'",
	).Scan(&name)
	if err != nil {
		t.Errorf("キャッシュDBにテーブル route_cache が存在しない: %v", err)
	}
}
