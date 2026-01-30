package main

import (
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

	// テンプレート設定
	t := &Template{
		templates: template.Must(template.ParseGlob("web/templates/*.html")),
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
		return c.String(200, "Standard Truck Rate - トラック運賃簡易予測システム")
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
