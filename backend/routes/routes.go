package routes

import (
	"io.lazydoge/aclove/handlers"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, messageHandler *handlers.MessageHandler) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/api/hello", messageHandler.Hello)

	api := router.Group("/api")
	{
		api.POST("/messages", messageHandler.Create)
		api.GET("/messages", messageHandler.List)
	}
}
