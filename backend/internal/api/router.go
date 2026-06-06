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
		projects := protected.Group("/projects")
		{
			projects.GET("", GetProjects)
			projects.POST("", CreateProject)
			projects.GET("/:id", GetProject)
			projects.PUT("/:id", UpdateProject)
			projects.DELETE("/:id", DeleteProject)
			projects.GET("/:id/chapters", GetChapters)
			projects.POST("/:id/chapters", CreateChapter)
		}
		protected.GET("/chapters/:id", GetChapter)
		protected.PUT("/chapters/:id", UpdateChapter)
	}
}
