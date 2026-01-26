package routes

import (
	"io.lazydoge/aclove/handlers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, categoryHandler *handlers.CategoryHandler) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	{
		categories := api.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.GetTree)
			categories.GET("/:id", categoryHandler.Get)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
			categories.GET("/:id/children", categoryHandler.GetChildren)
		}
	}
}
