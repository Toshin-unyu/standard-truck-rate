package main

import (
	"log"
	"path/filepath"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
)

func main() {
	dataDir := "data"

	// メインDB初期化
	mainDBPath := filepath.Join(dataDir, "str.db")
	mainDB, err := database.InitMainDB(mainDBPath)
	if err != nil {
		log.Fatalf("メインDB初期化エラー: %v", err)
	}
	mainDB.Close()
	log.Printf("メインDB作成完了: %s", mainDBPath)

	// キャッシュDB初期化
	cacheDBPath := filepath.Join(dataDir, "cache.db")
	cacheDB, err := database.InitCacheDB(cacheDBPath)
	if err != nil {
		log.Fatalf("キャッシュDB初期化エラー: %v", err)
	}
	cacheDB.Close()
	log.Printf("キャッシュDB作成完了: %s", cacheDBPath)

	log.Println("全データベース初期化完了")
}
