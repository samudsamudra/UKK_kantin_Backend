package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/samudsamudra/UKK_kantin/internal/routes"
	"github.com/samudsamudra/UKK_kantin/internal/app"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6767"
	}

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// DB
	app.InitDB()
	app.RunMigrations()

	// health
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Register all routes
	routes.Register(r)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("starting server on %s (mode=%s)", addr, mode)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
