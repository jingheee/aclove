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
	postHandler *handlers.PostHandler,
	uploadHandler *handlers.UploadHandler,
) {
	router.GET("/health", func(c *gin.Context) {
		models.JSONSuccess(c, gin.H{"status": "ok"})
	})

	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)
	requireActiveUser := middleware.RequireActiveUser()

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

		posts := api.Group("/posts")
		{
			posts.GET("", postHandler.List)
			posts.GET("/:id", postHandler.Get)
			posts.POST("", sessionMiddleware.Handler(), requireActiveUser, postHandler.Create)
			posts.PUT("/:id", sessionMiddleware.Handler(), requireActiveUser, postHandler.Update)
			posts.DELETE("/:id", sessionMiddleware.Handler(), requireActiveUser, postHandler.Delete)
		}

		upload := api.Group("/upload")
		{
			upload.POST("/image", sessionMiddleware.Handler(), requireActiveUser, uploadHandler.UploadImage)
			upload.POST("/file", sessionMiddleware.Handler(), requireActiveUser, uploadHandler.UploadFile)
			upload.DELETE("/file", sessionMiddleware.Handler(), uploadHandler.DeleteFile)
			upload.GET("/stats", uploadHandler.GetStats)
			upload.GET("/health", uploadHandler.HealthCheck)
		}
	}
}

func wrapCategoryGet(h *handlers.CategoryHandler, c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := parseCategoryID(ctx)
		if err != nil {
			models.JSONBadRequest(ctx, "无效的分类ID")
			return
		}

		var cachedResp models.CategoryResponse
		if err := c.Get(ctx.Request.Context(), id, &cachedResp); err == nil && cachedResp.ID != 0 {
			models.JSONSuccess(ctx, cachedResp)
			return
		}

		category, err := h.Get(ctx)
		if err != nil {
			models.JSONNotFound(ctx, "分类不存在")
			return
		}

		if err := c.Set(ctx.Request.Context(), id, category, 0); err != nil {
			models.JSONInternalError(ctx, "设置缓存失败")
			return
		}

		models.JSONSuccess(ctx, category)
	}
}

func wrapCategoryUpdate(h *handlers.CategoryHandler, c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := parseCategoryID(ctx)
		if err != nil {
			models.JSONBadRequest(ctx, "无效的分类ID")
			return
		}

		if err := c.Delete(ctx.Request.Context(), id); err != nil {
			models.JSONInternalError(ctx, "删除缓存失败")
			return
		}

		category, err := h.Update(ctx)
		if err != nil {
			models.JSONInternalError(ctx, "更新分类失败")
			return
		}

		models.JSONSuccess(ctx, category)
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
