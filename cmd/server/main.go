package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/samudsamudra/UKK_kantin/internal/app"
	"github.com/samudsamudra/UKK_kantin/internal/routes"
)

func main() {
	// =========================
	// Load ENV (safe switch)
	// =========================
	appEnv := os.Getenv("APP_ENV")

	switch appEnv {
	case "role-rework":
		log.Println("[ENV] loading .env.role-rework")
		if err := godotenv.Load(".env.role-rework"); err != nil {
			log.Println("[WARN] failed to load .env.role-rework")
		}
	default:
		log.Println("[ENV] loading .env")
		if err := godotenv.Load(".env"); err != nil {
			log.Println("[WARN] failed to load .env")
		}
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
	// Router
	// =========================
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// =========================
	// Database
	// =========================
	app.InitDB()
	app.RunMigrations()

	// =========================
	// Health check
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

	addr := fmt.Sprintf(":%s", port)
	log.Printf(
		"starting server on %s | mode=%s | env=%s",
		addr,
		mode,
		appEnv,
	)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
