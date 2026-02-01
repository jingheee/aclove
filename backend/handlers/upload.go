package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/middleware"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/models/query"
	"io.lazydoge/aclove/storage/minio"
)

// UploadHandler 文件上传处理器
type UploadHandler struct {
	storageService *minio.Service
	attachmentRepo *query.AttachmentRepo
}

// NewUploadHandler 创建上传处理器
func NewUploadHandler(storageService *minio.Service, attachmentRepo *query.AttachmentRepo) *UploadHandler {
	return &UploadHandler{
		storageService: storageService,
		attachmentRepo: attachmentRepo,
	}
}

// UploadImage 上传图片
func (h *UploadHandler) UploadImage(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		models.JSONUnauthorized(c, "请先获取会话")
		return
	}

	// 限制请求大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50*1024*1024) // 50MB

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			models.JSONBadRequest(c, "文件过大，最大支持 50MB")
			return
		}
		logger.Warn("获取上传文件失败", "error", err, "user_id", user.ID)
		models.JSONBadRequest(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 验证文件
	if err := h.validateImageFile(header); err != nil {
		logger.Warn("文件验证失败", "error", err, "user_id", user.ID, "filename", header.Filename)
		models.JSONBadRequest(c, err.Error())
		return
	}

	// 上传文件
	result, err := h.storageService.UploadImage(c.Request.Context(), header.Filename, file, header.Size)
	if err != nil {
		switch {
		case err == minio.ErrFileTooLarge:
			models.JSONBadRequest(c, "文件过大")
		case err == minio.ErrInvalidFileType:
			models.JSONBadRequest(c, "不支持的文件类型")
		case err == minio.ErrServiceUnavailable:
			logger.Error("MinIO 服务不可用", "error", err)
			models.JSONServiceUnavailable(c, "文件存储服务暂时不可用，请稍后重试")
		default:
			logger.Error("上传图片失败", "error", err, "user_id", user.ID, "filename", header.Filename)
			models.JSONInternalError(c, "上传失败，请稍后重试")
		}
		return
	}

	// 记录附件信息到数据库
	attachment := &query.AttachmentDO{
		ID:          models.GenerateSnowflakeID(),
		UserID:      int64(user.ID),
		Filename:    result.Filename,
		Size:        int64(result.Size),
		ContentType: result.ContentType,
		MinioKey:    result.Key,
		MinioURL:    result.URL,
		FileType:    "image",
	}

	if err := h.attachmentRepo.Create(c.Request.Context(), attachment); err != nil {
		logger.Error("记录附件信息失败", "error", err, "user_id", int64(user.ID), "filename", header.Filename)
		// 记录失败不影响上传结果，继续返回成功
	}

	logger.Info("图片上传成功",
		"user_id", user.ID,
		"filename", header.Filename,
		"size", result.Size,
		"url", result.URL,
		"attachment_id", attachment.ID,
	)

	models.JSONSuccess(c, gin.H{
		"url":          result.URL,
		"key":          result.Key,
		"size":         result.Size,
		"content_type": result.ContentType,
		"filename":     result.Filename,
		"attachment_id": attachment.ID,
		"download_url": "/api/download/" + fmt.Sprintf("%d", attachment.ID),
	})
}

// UploadFile 上传通用文件
func (h *UploadHandler) UploadFile(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		models.JSONUnauthorized(c, "请先获取会话")
		return
	}

	// 限制请求大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 100*1024*1024) // 100MB

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			models.JSONBadRequest(c, "文件过大，最大支持 100MB")
			return
		}
		logger.Warn("获取上传文件失败", "error", err, "user_id", user.ID)
		models.JSONBadRequest(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 上传文件
	result, err := h.storageService.UploadFile(c.Request.Context(), header.Filename, file, header.Size)
	if err != nil {
		switch {
		case err == minio.ErrFileTooLarge:
			models.JSONBadRequest(c, "文件过大")
		case err == minio.ErrInvalidFileType:
			models.JSONBadRequest(c, "不支持的文件类型")
		case err == minio.ErrServiceUnavailable:
			logger.Error("MinIO 服务不可用", "error", err)
			models.JSONServiceUnavailable(c, "文件存储服务暂时不可用，请稍后重试")
		default:
			logger.Error("上传文件失败", "error", err, "user_id", user.ID, "filename", header.Filename)
			models.JSONInternalError(c, "上传失败，请稍后重试")
		}
		return
	}

	// 记录附件信息到数据库
	attachment := &query.AttachmentDO{
		ID:          models.GenerateSnowflakeID(),
		UserID:      int64(user.ID),
		Filename:    result.Filename,
		Size:        int64(result.Size),
		ContentType: result.ContentType,
		MinioKey:    result.Key,
		MinioURL:    result.URL,
		FileType:    "file",
	}

	if err := h.attachmentRepo.Create(c.Request.Context(), attachment); err != nil {
		logger.Error("记录附件信息失败", "error", err, "user_id", int64(user.ID), "filename", header.Filename)
		// 记录失败不影响上传结果，继续返回成功
	}

	logger.Info("文件上传成功",
		"user_id", user.ID,
		"filename", header.Filename,
		"size", result.Size,
		"url", result.URL,
		"attachment_id", attachment.ID,
	)

	models.JSONSuccess(c, gin.H{
		"url":           result.URL,
		"key":           result.Key,
		"size":          result.Size,
		"content_type":  result.ContentType,
		"filename":      result.Filename,
		"attachment_id": attachment.ID,
		"download_url":  "/api/download/" + fmt.Sprintf("%d", attachment.ID),
	})
}

// DeleteFile 删除文件
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	user, exists := middleware.GetAnonymousUser(c)
	if !exists {
		models.JSONUnauthorized(c, "请先获取会话")
		return
	}

	var req struct {
		URL string `json:"url" binding:"required,url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		models.JSONBadRequest(c, "无效的请求数据")
		return
	}

	if err := h.storageService.DeleteFile(c.Request.Context(), req.URL); err != nil {
		logger.Error("删除文件失败", "error", err, "user_id", user.ID, "url", req.URL)
		models.JSONInternalError(c, "删除失败")
		return
	}

	logger.Info("文件删除成功", "user_id", user.ID, "url", req.URL)
	models.JSONSuccess(c, gin.H{"message": "删除成功"})
}

// GetStats 获取存储统计
func (h *UploadHandler) GetStats(c *gin.Context) {
	stats := h.storageService.GetStats()
	models.JSONSuccess(c, stats)
}

// HealthCheck 健康检查
func (h *UploadHandler) HealthCheck(c *gin.Context) {
	isHealthy := h.storageService.IsHealthy(c.Request.Context())
	if !isHealthy {
		models.JSONServiceUnavailable(c, "存储服务不可用")
		return
	}
	models.JSONSuccess(c, gin.H{"status": "healthy"})
}

func (h *UploadHandler) validateImageFile(header *multipart.FileHeader) error {
	// 验证文件大小（最大 10MB）
	const maxImageSize = 10 * 1024 * 1024
	if header.Size > maxImageSize {
		return fmt.Errorf("图片大小不能超过 10MB")
	}

	// 验证文件类型
	contentType := header.Header.Get("Content-Type")
	validImageTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	isValid := false
	for _, t := range validImageTypes {
		if strings.HasPrefix(contentType, t) {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("只支持 JPEG、PNG、GIF、WebP、SVG 格式的图片")
	}

	return nil
}
