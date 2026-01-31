package rustfs

import "errors"

var (
	// ErrFileNotFound 文件不存在
	ErrFileNotFound = errors.New("文件不存在")

	// ErrUploadFailed 上传失败
	ErrUploadFailed = errors.New("上传失败")

	// ErrDeleteFailed 删除失败
	ErrDeleteFailed = errors.New("删除失败")

	// ErrInvalidURL 无效的 URL
	ErrInvalidURL = errors.New("无效的 URL")

	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable = errors.New("RustFS 服务不可用")

	// ErrFileTooLarge 文件过大
	ErrFileTooLarge = errors.New("文件过大")

	// ErrInvalidFileType 无效的文件类型
	ErrInvalidFileType = errors.New("无效的文件类型")
)
