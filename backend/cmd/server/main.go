package main

import (
    "log"
    "net/http"
    "github.com/jj/novelist/internal/config"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()

    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    log.Printf("Server starting on port %s", cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
