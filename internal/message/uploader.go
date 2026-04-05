package message

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dfface/feishu-bot/internal/logger"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// DefaultFileUploader 默认的文件上传器实现
// 实现了 FileUploader 接口，提供图片和文件上传功能
type DefaultFileUploader struct {
	client *lark.Client
}

// NewFileUploader 创建文件上传器
//
// 参数:
//
//	client - 飞书 API 客户端
//
// 返回:
//
//	FileUploader - 初始化好的文件上传器实例
func NewFileUploader(client *lark.Client) FileUploader {
	return &DefaultFileUploader{
		client: client,
	}
}

// UploadImage 上传图片
// 将本地图片文件上传到飞书服务器，返回图片 key
//
// 参数:
//
//	ctx - 上下文，用于取消操作
//	imagePath - 本地图片文件路径
//	imageType - 图片类型（message 用于消息图片，avatar 用于头像）
//
// 返回:
//
//	string - 上传成功后返回的图片 key
//	error - 上传失败时返回错误
func (u *DefaultFileUploader) UploadImage(ctx context.Context, imagePath string, imageType ImageType) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	var larkImageType string
	switch imageType {
	case ImageTypeMessage:
		larkImageType = larkim.ImageTypeMessage
	case ImageTypeAvatar:
		larkImageType = larkim.ImageTypeAvatar
	default:
		larkImageType = larkim.ImageTypeMessage
	}

	resp, err := u.client.Im.V1.Image.Create(ctx,
		larkim.NewCreateImageReqBuilder().
			Body(larkim.NewCreateImageReqBodyBuilder().
				ImageType(larkImageType).
				Image(file).
				Build()).
			Build())

	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("failed to upload image: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	logger.Info("Image uploaded successfully",
		zap.String("image_key", *resp.Data.ImageKey),
		zap.String("image_path", filepath.Base(imagePath)),
	)
	return *resp.Data.ImageKey, nil
}

// UploadFile 上传文件
// 将本地文件上传到飞书服务器，返回文件 key
//
// 参数:
//
//	ctx - 上下文，用于取消操作
//	filePath - 本地文件路径
//	fileType - 文件类型（opus、mp4、pdf、doc、xls、stream）
//	fileName - 文件名，可选，不传则使用文件路径中的文件名
//
// 返回:
//
//	string - 上传成功后返回的文件 key
//	error - 上传失败时返回错误
func (u *DefaultFileUploader) UploadFile(ctx context.Context, filePath string, fileType FileType, fileName string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if fileName == "" {
		fileName = filepath.Base(filePath)
	}

	var larkFileType string
	switch fileType {
	case FileTypeOpus:
		larkFileType = larkim.FileTypeOpus
	case FileTypeMp4:
		larkFileType = larkim.FileTypeMp4
	case FileTypePdf:
		larkFileType = larkim.FileTypePdf
	case FileTypeDoc:
		larkFileType = larkim.FileTypeDoc
	case FileTypeXls:
		larkFileType = larkim.FileTypeXls
	case FileTypeStream:
		larkFileType = larkim.FileTypeStream
	default:
		larkFileType = larkim.FileTypeStream
	}

	resp, err := u.client.Im.V1.File.Create(ctx,
		larkim.NewCreateFileReqBuilder().
			Body(larkim.NewCreateFileReqBodyBuilder().
				FileType(larkFileType).
				FileName(fileName).
				File(file).
				Build()).
			Build())

	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	if !resp.Success() {
		return "", fmt.Errorf("failed to upload file: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	logger.Info("File uploaded successfully",
		zap.String("file_key", *resp.Data.FileKey),
		zap.String("file_name", fileName),
	)
	return *resp.Data.FileKey, nil
}

// 确保 DefaultFileUploader 实现了 FileUploader 接口
var _ FileUploader = (*DefaultFileUploader)(nil)
