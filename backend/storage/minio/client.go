package minio

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"io.lazydoge/aclove/logger"
)

const (
	defaultTimeout     = 30 * time.Second
	maxRetries         = 3
	defaultContentType = "application/octet-stream"
	defaultRegion      = "us-east-1"
)

// Client MinIO 存储客户端 (使用 AWS S3 SDK 兼容模式)
type Client struct {
	s3Client  *s3.Client
	bucket    string
	baseURL   string
	publicURL string
}

// Config 客户端配置
type Config struct {
	// Endpoint MinIO 服务地址 (如: http://localhost:9000)
	Endpoint string
	// Bucket 存储桶名称
	Bucket string
	// AccessKeyID 访问密钥 ID (默认: root)
	AccessKeyID string
	// SecretAccessKey 访问密钥 (默认: root)
	SecretAccessKey string
	// PublicURL 公开访问 URL (可选，默认使用 Endpoint)
	PublicURL string
	// Region 区域 (可选，默认为 us-east-1)
	Region string
	// Timeout 请求超时时间
	Timeout time.Duration
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

// NewClient 创建 MinIO 客户端
func NewClient(cfg Config) (*Client, error) {
	region := cfg.Region
	if region == "" {
		region = defaultRegion
	}

	accessKeyID := cfg.AccessKeyID
	if accessKeyID == "" {
		accessKeyID = "root"
	}

	secretAccessKey := cfg.SecretAccessKey
	if secretAccessKey == "" {
		secretAccessKey = "root"
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKeyID,
				secretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("加载 AWS 配置失败: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	publicURL := cfg.PublicURL
	if publicURL == "" {
		publicURL = cfg.Endpoint
	}

	return &Client{
		s3Client:  client,
		bucket:    cfg.Bucket,
		baseURL:   strings.TrimSuffix(cfg.Endpoint, "/"),
		publicURL: strings.TrimSuffix(publicURL, "/"),
	}, nil
}

// Upload 上传文件
func (c *Client) Upload(ctx context.Context, filename string, content io.Reader, contentType string) (*UploadResponse, error) {
	// 确保存储桶存在
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}

	key := c.GenerateKey(filename)

	if contentType == "" {
		contentType = defaultContentType
	}

	data, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		logger.Error("MinIO 上传失败", "key", key, "error", err)
		return nil, fmt.Errorf("上传失败: %w", err)
	}

	logger.Info("MinIO 上传成功", "filename", filename, "key", key, "size", len(data))

	return &UploadResponse{
		Success: true,
		URL:     c.GetPublicURL(key),
		Key:     key,
		Size:    int64(len(data)),
	}, nil
}

// ensureBucket 确保存储桶存在，不存在则自动创建
func (c *Client) ensureBucket(ctx context.Context) error {
	exists, err := c.BucketExists(ctx)
	if err != nil {
		return fmt.Errorf("检查存储桶失败: %w", err)
	}
	if !exists {
		if err := c.CreateBucket(ctx); err != nil {
			return fmt.Errorf("创建存储桶失败: %w", err)
		}
		logger.Info("MinIO 存储桶已自动创建", "bucket", c.bucket)
	}
	return nil
}

// UploadBytes 上传字节数据
func (c *Client) UploadBytes(ctx context.Context, filename string, data []byte, contentType string) (*UploadResponse, error) {
	return c.Upload(ctx, filename, bytes.NewReader(data), contentType)
}

// Delete 删除文件
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Error("MinIO 删除失败", "key", key, "error", err)
		return fmt.Errorf("删除失败: %w", err)
	}

	logger.Info("MinIO 删除成功", "key", key)
	return nil
}

// DeleteByURL 通过 URL 删除文件
func (c *Client) DeleteByURL(ctx context.Context, fileURL string) error {
	key, err := c.ExtractKeyFromURL(fileURL)
	if err != nil {
		return err
	}
	return c.Delete(ctx, key)
}

// GetFileInfo 获取文件信息
func (c *Client) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	result, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr *smithy.OperationError
		if errors.As(err, &apiErr) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	lastModified := time.Time{}
	if result.LastModified != nil {
		lastModified = *result.LastModified
	}

	return &FileInfo{
		Key:          key,
		Size:         aws.ToInt64(result.ContentLength),
		ContentType:  contentType,
		LastModified: lastModified,
		URL:          c.GetPublicURL(key),
	}, nil
}

// GetPublicURL 获取公开访问 URL
func (c *Client) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.publicURL, c.bucket, key)
}

// IsHealthy 检查服务健康状态
func (c *Client) IsHealthy(ctx context.Context) bool {
	_, err := c.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	return err == nil
}

// ExtractKeyFromURL 从 URL 中提取文件 key
func (c *Client) ExtractKeyFromURL(fileURL string) (string, error) {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("解析 URL 失败: %w", err)
	}

	expectedPrefix := "/" + c.bucket + "/"
	if !strings.HasPrefix(parsedURL.Path, expectedPrefix) {
		return "", fmt.Errorf("无效的 MinIO URL 格式")
	}

	key := strings.TrimPrefix(parsedURL.Path, expectedPrefix)
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

	name := strings.TrimSuffix(originalName, ext)
	name = sanitizeFilename(name)

	return fmt.Sprintf("%s/%d_%s%s", time.Now().Format("2006/01/02"), timestamp, random, ext)
}

// ListObjects 列出存储桶中的对象
func (c *Client) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]FileInfo, error) {
	if maxKeys <= 0 {
		maxKeys = 1000
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(c.bucket),
		MaxKeys: aws.Int32(maxKeys),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	result, err := c.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("列出对象失败: %w", err)
	}

	files := make([]FileInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		files = append(files, FileInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: aws.ToTime(obj.LastModified),
			URL:          c.GetPublicURL(aws.ToString(obj.Key)),
		})
	}

	return files, nil
}

// GetObject 获取对象内容
func (c *Client) GetObject(ctx context.Context, key string) (io.ReadCloser, *FileInfo, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr *smithy.OperationError
		if errors.As(err, &apiErr) {
			return nil, nil, ErrFileNotFound
		}
		return nil, nil, fmt.Errorf("获取对象失败: %w", err)
	}

	info := &FileInfo{
		Key:          key,
		Size:         aws.ToInt64(result.ContentLength),
		ContentType:  aws.ToString(result.ContentType),
		LastModified: aws.ToTime(result.LastModified),
		URL:          c.GetPublicURL(key),
	}

	return result.Body, info, nil
}

// CreateBucket 创建存储桶
func (c *Client) CreateBucket(ctx context.Context) error {
	_, err := c.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		var apiErr *smithy.OperationError
		if errors.As(err, &apiErr) {
			return nil
		}
		return fmt.Errorf("创建存储桶失败: %w", err)
	}
	return nil
}

// BucketExists 检查存储桶是否存在
func (c *Client) BucketExists(ctx context.Context) (bool, error) {
	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		var apiErr *smithy.OperationError
		if errors.As(err, &apiErr) {
			return false, nil
		}
		return false, fmt.Errorf("检查存储桶失败: %w", err)
	}
	return true, nil
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
