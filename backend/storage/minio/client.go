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

	"io.lazydoge/aclove/jsonutil"
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
	Success bool           `json:"success"`
	URL     string         `json:"url"`
	Key     string         `json:"key"`
	Size    jsonutil.Int64 `json:"size"`
	Error   string         `json:"error,omitempty"`
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string         `json:"key"`
	Size         jsonutil.Int64 `json:"size"`
	ContentType  string         `json:"content_type"`
	LastModified time.Time      `json:"last_modified"`
	URL          string         `json:"url"`
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

	// 解析 endpoint
	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "http://localhost:9000"
	}

	// 创建 S3 客户端（兼容 MinIO）
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // MinIO 需要使用 path-style
	})

	publicURL := cfg.PublicURL
	if publicURL == "" {
		publicURL = endpoint
	}

	client := &Client{
		s3Client:  s3Client,
		bucket:    cfg.Bucket,
		baseURL:   endpoint,
		publicURL: publicURL,
	}

	// 确保 bucket 存在
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" {
				// 创建 bucket
				_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
					Bucket: aws.String(cfg.Bucket),
				})
				if err != nil {
					return nil, fmt.Errorf("创建 bucket 失败: %w", err)
				}
				logger.Info("创建 bucket 成功", "bucket", cfg.Bucket)
			} else {
				return nil, fmt.Errorf("检查 bucket 失败: %w", err)
			}
		}
	}

	return client, nil
}

// Upload 上传文件
func (c *Client) Upload(ctx context.Context, key string, content io.Reader, size int64, contentType string) (*UploadResponse, error) {
	if contentType == "" {
		contentType = defaultContentType
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          content,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return &UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("上传失败: %v", err),
		}, fmt.Errorf("上传文件失败: %w", err)
	}

	return &UploadResponse{
		Success: true,
		URL:     c.GetPublicURL(key),
		Key:     key,
		Size:    jsonutil.Int64(size),
	}, nil
}

// GetPublicURL 获取文件的公开访问 URL
func (c *Client) GetPublicURL(key string) string {
	u, _ := url.Parse(c.publicURL)
	u.Path = path.Join(u.Path, c.bucket, key)
	return u.String()
}

// Delete 删除文件
func (c *Client) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// GetFileInfo 获取文件信息
func (c *Client) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
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

	size := int64(0)
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return &FileInfo{
		Key:          key,
		Size:         jsonutil.Int64(size),
		ContentType:  contentType,
		LastModified: lastModified,
		URL:          c.GetPublicURL(key),
	}, nil
}

// GenerateKey 生成文件存储 key
func (c *Client) GenerateKey(filename string) string {
	ext := path.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	// 清理文件名中的特殊字符
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// 添加时间戳避免冲突
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}

// GetPresignedURL 获取预签名 URL（用于临时访问）
func (c *Client) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3Client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("生成预签名 URL 失败: %w", err)
	}

	return req.URL, nil
}

// Download 下载文件
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey" {
			return nil, 0, ErrFileNotFound
		}
		return nil, 0, fmt.Errorf("下载文件失败: %w", err)
	}

	size := int64(0)
	if result.ContentLength != nil {
		size = *result.ContentLength
	}

	return result.Body, size, nil
}

// Copy 复制文件
func (c *Client) Copy(ctx context.Context, sourceKey, destKey string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	source := fmt.Sprintf("%s/%s", c.bucket, sourceKey)
	_, err := c.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(c.bucket),
		CopySource: aws.String(url.PathEscape(source)),
		Key:        aws.String(destKey),
	})
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	return nil
}

// ListFiles 列出文件
func (c *Client) ListFiles(ctx context.Context, prefix string, maxKeys int32) ([]FileInfo, error) {
	if maxKeys <= 0 {
		maxKeys = 1000
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(c.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(maxKeys),
	})
	if err != nil {
		return nil, fmt.Errorf("列出文件失败: %w", err)
	}

	files := make([]FileInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		key := ""
		if obj.Key != nil {
			key = *obj.Key
		}

		size := int64(0)
		if obj.Size != nil {
			size = *obj.Size
		}

		lastModified := time.Time{}
		if obj.LastModified != nil {
			lastModified = *obj.LastModified
		}

		files = append(files, FileInfo{
			Key:          key,
			Size:         jsonutil.Int64(size),
			LastModified: lastModified,
			URL:          c.GetPublicURL(key),
		})
	}

	return files, nil
}

// IsExist 检查文件是否存在
func (c *Client) IsExist(ctx context.Context, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
			return false, nil
		}
		return false, fmt.Errorf("检查文件存在性失败: %w", err)
	}

	return true, nil
}

// ReadFile 读取文件内容到内存
func (c *Client) ReadFile(ctx context.Context, key string) ([]byte, error) {
	reader, size, err := c.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 限制最大读取 100MB
	const maxSize = 100 * 1024 * 1024
	if size > maxSize {
		return nil, fmt.Errorf("文件太大，无法读取到内存: %d bytes", size)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	return buf.Bytes(), nil
}

// ExtractKeyFromURL 从 URL 提取 key
func (c *Client) ExtractKeyFromURL(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("解析 URL 失败: %w", err)
	}

	// 移除 bucket 前缀
	path := u.Path
	prefix := "/" + c.bucket + "/"
	if strings.HasPrefix(path, prefix) {
		return path[len(prefix):], nil
	}

	// 如果没有 bucket 前缀，直接返回路径（去掉开头的 /）
	if strings.HasPrefix(path, "/") {
		return path[1:], nil
	}

	return path, nil
}

// IsHealthy 检查客户端是否健康
func (c *Client) IsHealthy(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := c.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})

	return err == nil
}

// DeleteByURL 通过 URL 删除文件
func (c *Client) DeleteByURL(ctx context.Context, fileURL string) error {
	key, err := c.ExtractKeyFromURL(fileURL)
	if err != nil {
		return err
	}
	return c.Delete(ctx, key)
}
