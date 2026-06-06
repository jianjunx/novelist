package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/config"
	"github.com/jj/novelist/internal/store"
)

func main() {
	cfg := config.Load()
	store.InitDB(cfg.DatabaseURL)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
