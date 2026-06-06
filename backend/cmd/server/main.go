package main

import (
    "log"
    "github.com/jj/novelist/internal/config"
    "github.com/gin-gonic/gin"
)

func main() {
    cfg := config.Load()
    _ = cfg

    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    log.Printf("Server starting on port %s", cfg.ServerPort)
    if err := r.Run(":" + cfg.ServerPort); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}
