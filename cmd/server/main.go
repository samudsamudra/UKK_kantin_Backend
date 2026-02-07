package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/samudsamudra/UKK_kantin/internal/app"
	"github.com/samudsamudra/UKK_kantin/internal/routes"
	"github.com/samudsamudra/UKK_kantin/internal/seed"
)

func main() {
	// =========================
	// Load ENV
	// =========================
	appEnv := os.Getenv("APP_ENV")

	switch appEnv {
	case "role-rework":
		log.Println("[ENV] loading .env.role-rework")
		_ = godotenv.Load(".env.role-rework")
	default:
		log.Println("[ENV] loading .env")
		_ = godotenv.Load(".env")
	}

	// =========================
	// Config
	// =========================
	port := os.Getenv("PORT")
	if port == "" {
		port = "6767"
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	// =========================
	// Disable Gin default route print
	// =========================
	gin.DebugPrintRouteFunc = func(string, string, string, int) {}

	// =========================
	// Router
	// =========================
	r := gin.New()
	r.Use(app.PrettyLogger()) // 🔥 LOGGER WARNA
	r.Use(gin.Recovery())

	// =========================
	// Database
	// =========================
	app.InitDB()
	app.RunMigrations()

	// =========================
	// Seed (REALISTIC)
	// =========================
	seed.SeedSuperAdmin()
	seed.SeedStans()
	seed.SeedMenus()
	seed.SeedSiswas()

	// =========================
	// Health Check
	// =========================
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"env":    appEnv,
		})
	})

	// =========================
	// Routes
	// =========================
	routes.Register(r)

	// =========================
	// API REPORT (STARTUP)
	// =========================
	app.PrintRoutesReport(r, appEnv, port)

	// =========================
	// Run Server
	// =========================
	addr := fmt.Sprintf(":%s", port)
	log.Printf("[START] listening on %s\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}