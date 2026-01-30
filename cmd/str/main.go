package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/handler"
)

// Template テンプレートレンダラー
type Template struct {
	templates *template.Template
}

// Render テンプレートをレンダリング
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// テンプレート関数
var templateFuncs = template.FuncMap{
	// 数値を3桁区切りでフォーマット
	"formatNumber": func(n int) string {
		// 簡易的な3桁区切り実装
		str := fmt.Sprintf("%d", n)
		result := ""
		for i, c := range str {
			if i > 0 && (len(str)-i)%3 == 0 {
				result += ","
			}
			result += string(c)
		}
		return result
	},
	// 運輸局コードを名称に変換
	"regionName": func(code int) string {
		names := map[int]string{
			1: "北海道", 2: "東北", 3: "関東", 4: "北陸信越", 5: "中部",
			6: "近畿", 7: "中国", 8: "四国", 9: "九州", 10: "沖縄",
		}
		if name, ok := names[code]; ok {
			return name
		}
		return fmt.Sprintf("不明(%d)", code)
	},
	// 車格コードを名称に変換
	"vehicleName": func(code int) string {
		names := map[int]string{
			1: "小型車（2t）", 2: "中型車（4t）", 3: "大型車（10t）", 4: "トレーラー（20t）",
		}
		if name, ok := names[code]; ok {
			return name
		}
		return fmt.Sprintf("不明(%d)", code)
	},
	// 分を時間:分形式に変換
	"formatDuration": func(minutes int) string {
		hours := minutes / 60
		mins := minutes % 60
		if hours > 0 {
			return fmt.Sprintf("%d時間%d分", hours, mins)
		}
		return fmt.Sprintf("%d分", mins)
	},
}

func main() {
	// DB初期化
	mainDBPath := os.Getenv("MAIN_DB_PATH")
	if mainDBPath == "" {
		mainDBPath = "data/str.db"
	}
	cacheDBPath := os.Getenv("CACHE_DB_PATH")
	if cacheDBPath == "" {
		cacheDBPath = "data/cache.db"
	}

	mainDB, err := database.InitMainDB(mainDBPath)
	if err != nil {
		log.Fatalf("メインDB初期化エラー: %v", err)
	}
	defer mainDB.Close()

	cacheDB, err := database.InitCacheDB(cacheDBPath)
	if err != nil {
		log.Fatalf("キャッシュDB初期化エラー: %v", err)
	}
	defer cacheDB.Close()

	e := echo.New()

	// テンプレート設定（サブディレクトリも含めて読み込み）
	tmpl := template.New("").Funcs(templateFuncs)
	tmpl = template.Must(tmpl.ParseGlob("web/templates/*.html"))
	tmpl = template.Must(tmpl.ParseGlob("web/templates/partials/*.html"))
	t := &Template{
		templates: tmpl,
	}
	e.Renderer = t

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Static files
	e.Static("/static", "web/static")

	// ハンドラ
	highwayHandler := handler.NewHighwayHandler(mainDB, cacheDB)

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index.html", nil)
	})

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 高速道路料金ページ
	e.GET("/highway", func(c echo.Context) error {
		return c.Render(200, "highway.html", nil)
	})

	// 高速道路料金API
	e.GET("/api/highway/ic/search", highwayHandler.SearchIC)
	e.GET("/api/highway/toll", highwayHandler.GetToll)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
