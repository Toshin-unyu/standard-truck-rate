package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
)

func main() {
	// コマンドライン引数
	dbPath := flag.String("db", "data/str.db", "メインDBのパス")
	dryRun := flag.Bool("dry-run", false, "実際にDBに書き込まない（確認用）")
	flag.Parse()

	log.Println("=== ICマスタ取得ツール ===")

	// 1. ドラぷらAPIからICリストを取得
	log.Println("ドラぷらAPIからICリストを取得中...")
	client := service.NewDrivePlazaClient()
	ics, err := client.FetchICList()
	if err != nil {
		log.Fatalf("ICリスト取得エラー: %v", err)
	}
	log.Printf("取得件数: %d件", len(ics))

	if *dryRun {
		log.Println("--- dry-runモード：DBへの書き込みをスキップ ---")
		// 最初の10件を表示
		log.Println("取得データ（先頭10件）:")
		for i, ic := range ics {
			if i >= 10 {
				break
			}
			fmt.Printf("  %s: %s (%s) - %s\n", ic.Code, ic.Name, ic.Yomi, ic.RoadName)
		}
		os.Exit(0)
	}

	// 2. DBに保存
	absPath, err := filepath.Abs(*dbPath)
	if err != nil {
		log.Fatalf("パス解決エラー: %v", err)
	}

	log.Printf("DB: %s", absPath)
	db, err := database.InitMainDB(absPath)
	if err != nil {
		log.Fatalf("DB初期化エラー: %v", err)
	}
	defer db.Close()

	repo := repository.NewHighwayICRepository(db)

	// 既存データを削除（全件置換）
	log.Println("既存データを削除中...")
	if err := repo.DeleteAll(); err != nil {
		log.Fatalf("削除エラー: %v", err)
	}

	// 一括登録
	log.Println("ICマスタを登録中...")
	if err := repo.BulkCreate(ics); err != nil {
		log.Fatalf("登録エラー: %v", err)
	}

	// 件数確認
	count, err := repo.Count()
	if err != nil {
		log.Fatalf("件数取得エラー: %v", err)
	}

	log.Printf("登録完了: %d件", count)
	log.Println("=== 完了 ===")
}
