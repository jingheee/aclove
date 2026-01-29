package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/cache"
	"io.lazydoge/aclove/handlers"
	"io.lazydoge/aclove/middleware"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/service"
	"io.lazydoge/aclove/session"
)

func Setup(
	router *gin.Engine,
	categoryHandler *handlers.CategoryHandler,
	categoryCache *cache.Cache,
	anonymousUserSvc *service.AnonymousUserService,
	anonymousUserHandler *handlers.AnonymousUserHandler,
	sessionManager *session.Manager,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)

	api := router.Group("/api")
	{
		users := api.Group("/users")
		{
			users.GET("/me", sessionMiddleware.Handler(), anonymousUserHandler.GetCurrentUser)
		}

		admin := api.Group("/admin")
		{
			admin.POST("/users/ban", anonymousUserHandler.BanUser)
			admin.POST("/users/unban", anonymousUserHandler.UnbanUser)
			admin.POST("/users/cooldown", anonymousUserHandler.SetCooldown)
		}

		categories := api.Group("/categories")
		{
			categories.POST("", categoryHandler.Create)
			categories.GET("", categoryHandler.GetTree)
			categories.GET("/:id", wrapCategoryGet(categoryHandler, categoryCache))
			categories.PUT("/:id", wrapCategoryUpdate(categoryHandler, categoryCache))
			categories.DELETE("/:id", categoryHandler.Delete)
			categories.GET("/:id/children", categoryHandler.GetChildren)
		}
	}
}

func wrapCategoryGet(h *handlers.CategoryHandler, c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := parseCategoryID(ctx)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "无效的分类ID"})
			return
		}

		var cachedResp models.CategoryResponse
		if err := c.Get(ctx.Request.Context(), id, &cachedResp); err == nil && cachedResp.ID != 0 {
			ctx.JSON(200, cachedResp)
			return
		}

		category, err := h.Get(ctx)
		if err != nil {
			ctx.JSON(404, gin.H{"error": "分类不存在"})
			return
		}

		if err := c.Set(ctx.Request.Context(), id, category, 0); err != nil {
			ctx.JSON(500, gin.H{"error": "设置缓存失败"})
			return
		}

		ctx.JSON(200, category)
	}
}

func wrapCategoryUpdate(h *handlers.CategoryHandler, c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := parseCategoryID(ctx)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "无效的分类ID"})
			return
		}

		if err := c.Delete(ctx.Request.Context(), id); err != nil {
			ctx.JSON(500, gin.H{"error": "删除缓存失败"})
			return
		}

		category, err := h.Update(ctx)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "更新分类失败"})
			return
		}

		ctx.JSON(200, category)
	}
}

func parseCategoryID(ctx *gin.Context) (int64, error) {
	idStr := ctx.Param("id")
	if idStr == "" {
		return 0, nil
	}
	var id int64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	return id, err
}
