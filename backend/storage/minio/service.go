package minio

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"io.lazydoge/aclove/jsonutil"
	"io.lazydoge/aclove/logger"
)

// Service 文件存储服务
type Service struct {
	client       *Client
	maxFileSize  int64
	allowedTypes map[string]bool
	uploadSem    chan struct{}
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Client        *Client
	MaxFileSize   int64    // 最大文件大小（字节）
	MaxConcurrent int      // 最大并发上传数
	AllowedTypes  []string // 允许的文件类型（MIME type 前缀）
}

// UploadResult 上传结果
type UploadResult struct {
	URL         string         `json:"url"`
	Key         string         `json:"key"`
	Size        jsonutil.Int64 `json:"size"`
	ContentType string         `json:"content_type"`
	Filename    string         `json:"filename"`
}

// NewService 创建文件存储服务
func NewService(cfg ServiceConfig) *Service {
	maxConcurrent := cfg.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	allowedTypes := make(map[string]bool)
	for _, t := range cfg.AllowedTypes {
		allowedTypes[t] = true
	}

	if len(allowedTypes) == 0 {
		allowedTypes = map[string]bool{
			"image/":          true,
			"video/":          true,
			"audio/":          true,
			"text/":           true,
			"application/pdf": true,
		}
	}

	maxFileSize := cfg.MaxFileSize
	if maxFileSize <= 0 {
		maxFileSize = 50 * 1024 * 1024
	}

	return &Service{
		client:       cfg.Client,
		maxFileSize:  maxFileSize,
		allowedTypes: allowedTypes,
		uploadSem:    make(chan struct{}, maxConcurrent),
	}
}

// UploadImage 上传图片
func (s *Service) UploadImage(ctx context.Context, filename string, content io.Reader, size int64) (*UploadResult, error) {
	if size > s.maxFileSize {
		return nil, fmt.Errorf("%w: 最大允许 %d MB", ErrFileTooLarge, s.maxFileSize/1024/1024)
	}

	contentType := detectContentType(filename, content)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("%w: 只允许上传图片文件", ErrInvalidFileType)
	}

	return s.upload(ctx, filename, content, size, contentType)
}

// UploadFile 上传通用文件
func (s *Service) UploadFile(ctx context.Context, filename string, content io.Reader, size int64) (*UploadResult, error) {
	if size > s.maxFileSize {
		return nil, fmt.Errorf("%w: 最大允许 %d MB", ErrFileTooLarge, s.maxFileSize/1024/1024)
	}

	contentType := detectContentType(filename, content)
	if !s.isAllowedType(contentType) {
		return nil, fmt.Errorf("%w: 不支持的文件类型 %s", ErrInvalidFileType, contentType)
	}

	return s.upload(ctx, filename, content, size, contentType)
}

// UploadFromURL 从 URL 上传文件
func (s *Service) UploadFromURL(ctx context.Context, fileURL string, filename string) (*UploadResult, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载文件失败，状态码: %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size > s.maxFileSize {
		return nil, fmt.Errorf("%w: 最大允许 %d MB", ErrFileTooLarge, s.maxFileSize/1024/1024)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(filename, resp.Body)
	}

	return s.upload(ctx, filename, resp.Body, size, contentType)
}

// Delete 删除文件
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.client.Delete(ctx, key)
}

// DeleteByURL 通过 URL 删除文件
func (s *Service) DeleteByURL(ctx context.Context, fileURL string) error {
	key, err := s.client.ExtractKeyFromURL(fileURL)
	if err != nil {
		return err
	}
	return s.client.Delete(ctx, key)
}

// GetFileInfo 获取文件信息
func (s *Service) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	return s.client.GetFileInfo(ctx, key)
}

// IsHealthy 检查服务健康状态
func (s *Service) IsHealthy(ctx context.Context) bool {
	return s.client.IsHealthy(ctx)
}

// DeleteFile 通过 URL 删除文件
func (s *Service) DeleteFile(ctx context.Context, fileURL string) error {
	return s.client.DeleteByURL(ctx, fileURL)
}

// Download 下载文件
func (s *Service) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	return s.client.Download(ctx, key)
}

// GetStats 获取存储统计
func (s *Service) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_file_size":  s.maxFileSize,
		"max_concurrent": cap(s.uploadSem),
		"allowed_types":  s.allowedTypes,
	}
}

func (s *Service) upload(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (*UploadResult, error) {
	s.uploadSem <- struct{}{}
	defer func() { <-s.uploadSem }()

	// 生成唯一的存储key
	key := s.client.GenerateKey(filename)

	resp, err := s.client.Upload(ctx, key, content, size, contentType)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL:         resp.URL,
		Key:         resp.Key,
		Size:        resp.Size,
		ContentType: contentType,
		Filename:    filename,
	}, nil
}

func (s *Service) isAllowedType(contentType string) bool {
	for prefix := range s.allowedTypes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}

func detectContentType(filename string, content io.Reader) string {
	if ext := filepath.Ext(filename); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}

	buf := make([]byte, 512)
	n, err := content.Read(buf)
	if err != nil && err != io.EOF {
		logger.Warn("检测文件类型失败", "error", err)
	}

	if seeker, ok := content.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return http.DetectContentType(buf[:n])
}
