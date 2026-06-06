package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jj/novelist/internal/auth"
)

func SetupRouter(r *gin.Engine) {
	api := r.Group("/api")

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", Register)
		authGroup.POST("/login", Login)
		authGroup.GET("/me", auth.AuthMiddleware(), GetMe)
	}

	protected := api.Group("")
	protected.Use(auth.AuthMiddleware())
	{
		// Routes will be added in subsequent tasks
		_ = protected
	}
}
