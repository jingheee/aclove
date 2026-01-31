package rustfs

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"io.lazydoge/aclove/logger"
)

// Service 文件存储服务
type Service struct {
	client      *Client
	maxFileSize int64
	allowedTypes map[string]bool
	uploadSem   chan struct{}
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	Client       *Client
	MaxFileSize  int64             // 最大文件大小（字节）
	MaxConcurrent int              // 最大并发上传数
	AllowedTypes []string          // 允许的文件类型（MIME type 前缀）
}

// UploadResult 上传结果
type UploadResult struct {
	URL         string `json:"url"`
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
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

	// 默认允许的图片类型
	if len(allowedTypes) == 0 {
		allowedTypes = map[string]bool{
			"image/":    true,
			"video/":    true,
			"audio/":    true,
			"text/":     true,
			"application/pdf": true,
		}
	}

	maxFileSize := cfg.MaxFileSize
	if maxFileSize <= 0 {
		maxFileSize = 50 * 1024 * 1024 // 默认 50MB
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
	// 验证文件大小
	if size > s.maxFileSize {
		return nil, fmt.Errorf("%w: 最大允许 %d MB", ErrFileTooLarge, s.maxFileSize/1024/1024)
	}

	// 验证文件类型
	contentType := detectContentType(filename, content)
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("%w: 只允许上传图片文件", ErrInvalidFileType)
	}

	return s.upload(ctx, filename, content, size, contentType)
}

// UploadFile 上传通用文件
func (s *Service) UploadFile(ctx context.Context, filename string, content io.Reader, size int64) (*UploadResult, error) {
	// 验证文件大小
	if size > s.maxFileSize {
		return nil, fmt.Errorf("%w: 最大允许 %d MB", ErrFileTooLarge, s.maxFileSize/1024/1024)
	}

	// 验证文件类型
	contentType := detectContentType(filename, content)
	if !s.isAllowedType(contentType) {
		return nil, fmt.Errorf("%w: 不支持的文件类型 %s", ErrInvalidFileType, contentType)
	}

	return s.upload(ctx, filename, content, size, contentType)
}

// UploadFromURL 从 URL 下载并上传文件
func (s *Service) UploadFromURL(ctx context.Context, fileURL string) (*UploadResult, error) {
	// 使用 semaphore 限制并发
	select {
	case s.uploadSem <- struct{}{}:
		defer func() { <-s.uploadSem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 下载文件
	resp, err := s.client.httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("下载文件失败: HTTP %d", resp.StatusCode)
	}

	size := resp.ContentLength
	if size > s.maxFileSize {
		return nil, fmt.Errorf("%w: 文件大小 %d MB 超过限制 %d MB", 
			ErrFileTooLarge, size/1024/1024, s.maxFileSize/1024/1024)
	}

	filename := filepath.Base(fileURL)
	if filename == "" || filename == "." {
		filename = "download"
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(filename, resp.Body)
	}

	return s.upload(ctx, filename, resp.Body, size, contentType)
}

// DeleteFile 删除文件
func (s *Service) DeleteFile(ctx context.Context, fileURL string) error {
	key, err := s.client.ExtractKeyFromURL(fileURL)
	if err != nil {
		// 如果不是我们的 URL，静默忽略
		logger.Warn("无法从 URL 提取 key", "url", fileURL, "error", err)
		return nil
	}

	if err := s.client.Delete(ctx, key); err != nil {
		if err == ErrFileNotFound {
			return nil // 文件已不存在，视为成功
		}
		return err
	}

	return nil
}

// DeleteFiles 批量删除文件
func (s *Service) DeleteFiles(ctx context.Context, urls []string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))

	for _, fileURL := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			if err := s.DeleteFile(ctx, url); err != nil {
				errChan <- fmt.Errorf("删除 %s 失败: %w", url, err)
			}
		}(fileURL)
	}

	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("批量删除失败: %d 个错误", len(errs))
	}

	return nil
}

// GetFileInfo 获取文件信息
func (s *Service) GetFileInfo(ctx context.Context, fileURL string) (*FileInfo, error) {
	key, err := s.client.ExtractKeyFromURL(fileURL)
	if err != nil {
		return nil, err
	}

	return s.client.GetFileInfo(ctx, key)
}

// IsHealthy 检查服务健康状态
func (s *Service) IsHealthy(ctx context.Context) bool {
	return s.client.IsHealthy(ctx)
}

// GetStats 获取服务统计
func (s *Service) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"max_file_size_mb":   s.maxFileSize / 1024 / 1024,
		"max_concurrent":     cap(s.uploadSem),
		"current_uploads":    len(s.uploadSem),
		"allowed_types":      s.allowedTypes,
	}
}

func (s *Service) upload(ctx context.Context, filename string, content io.Reader, size int64, contentType string) (*UploadResult, error) {
	// 使用 semaphore 限制并发
	select {
	case s.uploadSem <- struct{}{}:
		defer func() { <-s.uploadSem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 生成存储 key
	key := s.client.GenerateKey(filename)

	start := time.Now()
	resp, err := s.client.Upload(ctx, key, content, contentType)
	if err != nil {
		logger.Error("文件上传失败", "error", err, "filename", filename, "key", key)
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	duration := time.Since(start)
	logger.Info("文件上传成功",
		"filename", filename,
		"key", key,
		"size", size,
		"duration_ms", duration.Milliseconds(),
	)

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
	// 优先从文件扩展名检测
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mp3":
		return "audio/mpeg"
	case ".pdf":
		return "application/pdf"
	}

	// 使用 MIME 类型检测
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}

	// 尝试读取内容检测
	buf := make([]byte, 512)
	n, _ := content.Read(buf)
	if n > 0 {
		return http.DetectContentType(buf[:n])
	}

	return "application/octet-stream"
}


