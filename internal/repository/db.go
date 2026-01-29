package repository

import (
	"database/sql"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
)

// DB データベース接続を管理する構造体
type DB struct {
	mainDB  *sql.DB
	cacheDB *sql.DB
}

// NewDB メインDBとキャッシュDBを初期化して接続する
func NewDB(mainDBPath, cacheDBPath string) (*DB, error) {
	mainDB, err := database.InitMainDB(mainDBPath)
	if err != nil {
		return nil, err
	}

	cacheDB, err := database.InitCacheDB(cacheDBPath)
	if err != nil {
		mainDB.Close()
		return nil, err
	}

	return &DB{
		mainDB:  mainDB,
		cacheDB: cacheDB,
	}, nil
}

// MainDB メインDBの接続を返す
func (d *DB) MainDB() *sql.DB {
	return d.mainDB
}

// CacheDB キャッシュDBの接続を返す
func (d *DB) CacheDB() *sql.DB {
	return d.cacheDB
}

// Close 両方のDB接続を閉じる
func (d *DB) Close() error {
	var firstErr error

	if err := d.mainDB.Close(); err != nil {
		firstErr = err
	}

	if err := d.cacheDB.Close(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}
