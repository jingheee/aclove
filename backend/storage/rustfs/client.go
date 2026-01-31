package rustfs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"io.lazydoge/aclove/logger"
)

const (
	defaultTimeout     = 30 * time.Second
	maxRetries         = 3
	retryDelay         = 500 * time.Millisecond
	defaultContentType = "application/octet-stream"
)

// Client RustFS 存储客户端
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// Config 客户端配置
type Config struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

// UploadResponse 上传响应
type UploadResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Key     string `json:"key"`
	Size    int64  `json:"size"`
	Error   string `json:"error,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type"`
	LastModified time.Time `json:"last_modified"`
	URL          string    `json:"url"`
}

// NewClient 创建 RustFS 客户端
func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return &Client{
		baseURL:  strings.TrimSuffix(cfg.BaseURL, "/"),
		username: cfg.Username,
		password: cfg.Password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Upload 上传文件
func (c *Client) Upload(ctx context.Context, filename string, content io.Reader, contentType string) (*UploadResponse, error) {
	if contentType == "" {
		contentType = defaultContentType
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("创建表单文件失败: %w", err)
	}

	if _, err := io.Copy(part, content); err != nil {
		return nil, fmt.Errorf("写入文件内容失败: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭表单写入器失败: %w", err)
	}

	uploadURL := fmt.Sprintf("%s/api/upload", c.baseURL)

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay * time.Duration(i))
			logger.Warn("RustFS 上传重试", "attempt", i+1, "filename", filename)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(body.Bytes()))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", c.basicAuth())

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
			continue
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("上传失败: HTTP %d, %s", resp.StatusCode, string(respBody))
			if resp.StatusCode >= 500 {
				continue // 服务器错误，重试
			}
			return nil, lastErr // 客户端错误，不重试
		}

		var result UploadResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		if !result.Success {
			return nil, fmt.Errorf("上传失败: %s", result.Error)
		}

		logger.Info("RustFS 上传成功", "filename", filename, "key", result.Key, "size", result.Size)
		return &result, nil
	}

	return nil, fmt.Errorf("上传失败，已重试 %d 次: %w", maxRetries, lastErr)
}

// UploadBytes 上传字节数据
func (c *Client) UploadBytes(ctx context.Context, filename string, data []byte, contentType string) (*UploadResponse, error) {
	return c.Upload(ctx, filename, bytes.NewReader(data), contentType)
}

// Delete 删除文件
func (c *Client) Delete(ctx context.Context, key string) error {
	deleteURL := fmt.Sprintf("%s/api/files/%s", c.baseURL, url.PathEscape(key))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除失败: HTTP %d, %s", resp.StatusCode, string(body))
	}

	logger.Info("RustFS 删除成功", "key", key)
	return nil
}

// GetFileInfo 获取文件信息
func (c *Client) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	infoURL := fmt.Sprintf("%s/api/files/%s/info", c.baseURL, url.PathEscape(key))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, infoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrFileNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("获取文件信息失败: HTTP %d, %s", resp.StatusCode, string(body))
	}

	var info FileInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &info, nil
}

// GetPublicURL 获取公开访问 URL
func (c *Client) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/files/%s", c.baseURL, url.PathEscape(key))
}

// IsHealthy 检查服务健康状态
func (c *Client) IsHealthy(ctx context.Context) bool {
	healthURL := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ExtractKeyFromURL 从 URL 中提取文件 key
func (c *Client) ExtractKeyFromURL(fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("解析 URL 失败: %w", err)
	}

	baseParsed, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("解析 base URL 失败: %w", err)
	}

	if parsedURL.Host != baseParsed.Host {
		return "", fmt.Errorf("URL 不属于当前 RustFS 实例")
	}

	// 期望路径格式: /files/{key}
	const prefix = "/files/"
	if !strings.HasPrefix(parsedURL.Path, prefix) {
		return "", fmt.Errorf("无效的 RustFS URL 格式")
	}

	key := strings.TrimPrefix(parsedURL.Path, prefix)
	key, err = url.PathUnescape(key)
	if err != nil {
		return "", fmt.Errorf("解码 key 失败: %w", err)
	}

	return key, nil
}

// GenerateKey 生成存储 key
func (c *Client) GenerateKey(originalName string) string {
	ext := path.Ext(originalName)
	timestamp := time.Now().UnixNano()
	random := generateRandomString(8)

	// 清理文件名，移除特殊字符
	name := strings.TrimSuffix(originalName, ext)
	name = sanitizeFilename(name)

	return fmt.Sprintf("%s/%d_%s%s", time.Now().Format("2006/01/02"), timestamp, random, ext)
}

func (c *Client) basicAuth() string {
	auth := c.username + ":" + c.password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func sanitizeFilename(name string) string {
	// 移除或替换不安全的字符
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}
