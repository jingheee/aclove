package routes

import (
	"io.lazydoge/aclove/handlers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, messageHandler *handlers.MessageHandler, categoryHandler *handlers.CategoryHandler) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/api/hello", messageHandler.Hello)

	api := router.Group("/api")
	{
		api.POST("/messages", messageHandler.Create)
		api.GET("/messages", messageHandler.List)

		categories := api.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.List)
			categories.GET("/:id", categoryHandler.Get)
			categories.PUT("/:id", categoryHandler.Update)
			categories.DELETE("/:id", categoryHandler.Delete)
			categories.GET("/:id/children", categoryHandler.GetChildren)
		}
	}
}
