package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/y-suzuki/standard-truck-rate/internal/database"
	"github.com/y-suzuki/standard-truck-rate/internal/handler"
	"github.com/y-suzuki/standard-truck-rate/internal/model"
	"github.com/y-suzuki/standard-truck-rate/internal/repository"
	"github.com/y-suzuki/standard-truck-rate/internal/service"
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
	// 減算
	"sub": func(a, b int) int {
		return a - b
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

	// サービス作成
	fareCalculator := createFareCalculatorService(mainDB)

	// ルートクライアント・Geocodingクライアント作成（モック or Google API）
	var routeClient service.RouteClient
	var geocodingClient service.GeocodingClient
	googleAPIKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if googleAPIKey != "" {
		routeClient = service.NewGoogleRoutesClient(googleAPIKey)
		geocodingClient = service.NewGoogleGeocodingClient(googleAPIKey)
		log.Println("Google Maps APIを使用します")
	} else {
		log.Println("GOOGLE_MAPS_API_KEYが未設定のため、モッククライアントを使用します")
		routeClient = service.NewMockRoutesClient()
		geocodingClient = service.NewMockGeocodingClient()
	}

	// API使用量サービス（ルートハンドラで使用するため先に作成）
	apiUsageRepo := repository.NewApiUsageRepository(mainDB)
	apiUsageService := service.NewApiUsageService(apiUsageRepo)

	// ハンドラ
	highwayHandler := handler.NewHighwayHandler(mainDB, cacheDB)
	indexHandler := handler.NewIndexHandler()
	calculateHandler := handler.NewCalculateHandler(fareCalculator, routeClient, geocodingClient, mainDB, cacheDB)
	routeHandler := handler.NewRouteHandler(cacheDB, routeClient, apiUsageService)
	apiUsageHandler := handler.NewApiUsageHandler(apiUsageService)

	// Routes
	e.GET("/", indexHandler.Index)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// 運賃計算API
	e.POST("/api/fare/calculate", calculateHandler.Calculate)
	e.POST("/api/fare/calculate/json", calculateHandler.CalculateJSON)

	// ルート情報API
	e.GET("/api/route", routeHandler.GetRoute)

	// 高速道路料金API
	e.GET("/api/highway/ic/search", highwayHandler.SearchIC)
	e.GET("/api/highway/toll", highwayHandler.GetToll)

	// API使用量
	e.GET("/api/usage", apiUsageHandler.GetUsage)

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

// createFareCalculatorService 運賃計算サービスを作成
func createFareCalculatorService(mainDB *sql.DB) *service.FareCalculatorService {
	// Supabase設定
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	var distanceFareService *service.DistanceFareService

	if supabaseURL != "" && supabaseKey != "" {
		// Supabaseクライアントを使用
		supabaseClient := service.NewJtaSupabaseClient(supabaseURL, supabaseKey)
		adapter := service.NewJtaSupabaseClientAdapter(supabaseClient)
		distanceFareService = service.NewDistanceFareService(adapter)
	} else {
		log.Println("SUPABASE_URL/SUPABASE_ANON_KEYが未設定のため、距離制運賃はモックを使用します")
		distanceFareService = service.NewDistanceFareService(&mockFareGetter{})
	}

	// 時間制運賃（DBから取得）
	timeFareRepo := repository.NewJtaTimeFareRepository(mainDB)
	timeFareService := service.NewTimeFareService(timeFareRepo)

	// 赤帽運賃
	akabouFareService := service.NewAkabouFareService()

	return service.NewFareCalculatorService(distanceFareService, timeFareService, akabouFareService)
}

// mockFareGetter 距離制運賃のモック
type mockFareGetter struct{}

func (m *mockFareGetter) GetDistanceFareYen(regionCode, vehicleCode, distanceKm int) (int, error) {
	// モック: 基本的な運賃計算
	baseFare := 10000 + distanceKm*100
	return baseFare, nil
}

// mockTimeFareGetter 時間制運賃のモック
type mockTimeFareGetter struct{}

func (m *mockTimeFareGetter) GetBaseFare(regionCode, vehicleCode, hours int) (*model.JtaTimeBaseFare, error) {
	return &model.JtaTimeBaseFare{
		RegionCode:  regionCode,
		VehicleCode: vehicleCode,
		Hours:       hours,
		FareYen:     15000,
		BaseKm:      30,
	}, nil
}

func (m *mockTimeFareGetter) GetSurcharge(regionCode, vehicleCode int, surchargeType string) (*model.JtaTimeSurcharge, error) {
	fareYen := 0
	switch surchargeType {
	case "distance":
		fareYen = 50
	case "time":
		fareYen = 500
	}
	return &model.JtaTimeSurcharge{
		RegionCode:    regionCode,
		VehicleCode:   vehicleCode,
		SurchargeType: surchargeType,
		FareYen:       fareYen,
	}, nil
}
