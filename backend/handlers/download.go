package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"io.lazydoge/aclove/logger"
	"io.lazydoge/aclove/models"
	"io.lazydoge/aclove/models/query"
	"io.lazydoge/aclove/storage/minio"
)

// DownloadHandler 文件下载处理器
type DownloadHandler struct {
	attachmentRepo *query.AttachmentRepo
	storageService *minio.Service
}

// NewDownloadHandler 创建下载处理器
func NewDownloadHandler(attachmentRepo *query.AttachmentRepo, storageService *minio.Service) *DownloadHandler {
	return &DownloadHandler{
		attachmentRepo: attachmentRepo,
		storageService: storageService,
	}
}

// Download 通用文件下载接口
// 根据附件ID（雪花ID）下载文件
func (h *DownloadHandler) Download(c *gin.Context) {
	attachmentIDStr := c.Param("id")
	if attachmentIDStr == "" {
		models.JSONBadRequest(c, "缺少附件ID")
		return
	}

	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		models.JSONBadRequest(c, "无效的附件ID")
		return
	}

	// 查询附件信息
	attachment, err := h.attachmentRepo.GetByID(c.Request.Context(), attachmentID)
	if err != nil {
		if err == query.ErrAttachmentNotFound {
			models.JSONNotFound(c, "附件不存在")
			return
		}
		logger.Error("查询附件信息失败", "error", err, "attachment_id", attachmentID)
		models.JSONInternalError(c, "获取附件信息失败")
		return
	}

	// 从MinIO下载文件
	reader, size, err := h.storageService.Download(c.Request.Context(), attachment.MinioKey)
	if err != nil {
		if err == minio.ErrFileNotFound {
			models.JSONNotFound(c, "文件不存在")
			return
		}
		logger.Error("下载文件失败", "error", err, "attachment_id", attachmentID, "minio_key", attachment.MinioKey)
		models.JSONInternalError(c, "下载文件失败")
		return
	}
	defer reader.Close()

	// 设置响应头
	contentType := attachment.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 处理文件名，确保中文文件名正确显示
	filename := attachment.Filename
	if needsURLEncoding(filename) {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", urlEncode(filename)))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(size, 10))
	c.Header("Cache-Control", "public, max-age=31536000")

	// 流式传输文件内容
	c.Status(http.StatusOK)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		logger.Error("传输文件内容失败", "error", err, "attachment_id", attachmentID)
		return
	}

	logger.Info("文件下载成功",
		"attachment_id", attachmentID,
		"filename", attachment.Filename,
		"size", size,
		"user_agent", c.Request.UserAgent(),
	)
}

// GetAttachmentInfo 获取附件信息（不下载文件）
func (h *DownloadHandler) GetAttachmentInfo(c *gin.Context) {
	attachmentIDStr := c.Param("id")
	if attachmentIDStr == "" {
		models.JSONBadRequest(c, "缺少附件ID")
		return
	}

	attachmentID, err := strconv.ParseInt(attachmentIDStr, 10, 64)
	if err != nil {
		models.JSONBadRequest(c, "无效的附件ID")
		return
	}

	attachment, err := h.attachmentRepo.GetByID(c.Request.Context(), attachmentID)
	if err != nil {
		if err == query.ErrAttachmentNotFound {
			models.JSONNotFound(c, "附件不存在")
			return
		}
		logger.Error("查询附件信息失败", "error", err, "attachment_id", attachmentID)
		models.JSONInternalError(c, "获取附件信息失败")
		return
	}

	models.JSONSuccess(c, gin.H{
		"id":           attachment.ID,
		"filename":     attachment.Filename,
		"size":         attachment.Size,
		"content_type": attachment.ContentType,
		"file_type":    attachment.FileType,
		"created_at":   attachment.CreatedAt,
		"download_url": "/api/download/" + fmt.Sprintf("%d", attachment.ID),
	})
}

// needsURLEncoding 检查文件名是否需要URL编码
func needsURLEncoding(s string) bool {
	for _, r := range s {
		if r > 127 || r == '%' || r == ' ' {
			return true
		}
	}
	return false
}

// urlEncode 对字符串进行URL编码（简化版）
func urlEncode(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r > 127 || r == '%' || r == ' ' || r == '"' || r == '\\' {
			result.WriteString(fmt.Sprintf("%%%02X", r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
