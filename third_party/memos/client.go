package memos

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/dfface/feishu-bot/internal/logger"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
	"github.com/usememos/memos/proto/gen/api/v1/apiv1connect"
	"go.uber.org/zap"
)

// Visibility Memo 的可见性选项
type Visibility string

const (
	// VisibilityPrivate 私有可见性，只有创建者可以访问
	VisibilityPrivate Visibility = "PRIVATE"
	// VisibilityProtected 受保护的可见性
	VisibilityProtected Visibility = "PROTECTED"
	// VisibilityPublic 公共可见性，所有人都可以访问
	VisibilityPublic Visibility = "PUBLIC"
)

// MIME 类型映射表，用于根据文件扩展名推断 MIME 类型
var mimeTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".flac": "audio/flac",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain",
	".md":   "text/markdown",
	".json": "application/json",
	".xml":  "application/xml",
	".yml":  "text/yaml",
	".yaml": "text/yaml",
	".csv":  "text/csv",
	".zip":  "application/zip",
	".rar":  "application/x-rar-compressed",
	".tar":  "application/x-tar",
	".gz":   "application/gzip",
}

// ClientInterface 定义了 Memos 客户端的接口
type ClientInterface interface {
	// CreateMemo 创建一个新的 Memo
	// 参数:
	//   ctx - 上下文
	//   content - Memo 的内容，支持 Markdown 格式
	//   visibility - Memo 的可见性，默认为 Private
	// 返回值:
	//   *v1pb.Memo - 创建的 Memo 对象
	//   error - 错误信息
	CreateMemo(ctx context.Context, content string, visibility Visibility) (*v1pb.Memo, error)

	// GetMemo 根据名称获取单个 Memo
	// 参数:
	//   ctx - 上下文
	//   memoName - Memo 的名称，格式为 "memos/{memo_id}"
	// 返回值:
	//   *v1pb.Memo - 获取的 Memo 对象
	//   error - 错误信息
	GetMemo(ctx context.Context, memoName string) (*v1pb.Memo, error)

	// UpdateMemo 更新现有的 Memo
	// 参数:
	//   ctx - 上下文
	//   memo - 包含更新内容的 Memo 对象
	// 返回值:
	//   *v1pb.Memo - 更新后的 Memo 对象
	//   error - 错误信息
	UpdateMemo(ctx context.Context, memo *v1pb.Memo) (*v1pb.Memo, error)

	// DeleteMemo 删除指定的 Memo
	// 参数:
	//   ctx - 上下文
	//   memoName - Memo 的名称，格式为 "memos/{memo_id}"
	// 返回值:
	//   error - 错误信息
	DeleteMemo(ctx context.Context, memoName string) error

	// ListMemos 列出 Memos，支持分页
	// 参数:
	//   ctx - 上下文
	//   pageSize - 每页的数量，默认为 50
	//   pageToken - 分页令牌，用于获取下一页
	// 返回值:
	//   *v1pb.ListMemosResponse - 包含 Memos 列表和分页信息的响应
	//   error - 错误信息
	ListMemos(ctx context.Context, pageSize int32, pageToken string) (*v1pb.ListMemosResponse, error)

	// UploadResource 从文件上传资源到 Memos
	// 参数:
	//   ctx - 上下文
	//   filePath - 要上传的文件路径
	//   memoName - 要关联的 Memo 名称，可选，格式为 "memos/{memo_id}"
	// 返回值:
	//   *v1pb.Attachment - 上传的附件对象
	//   error - 错误信息
	UploadResource(ctx context.Context, filePath string, memoName string) (*v1pb.Attachment, error)

	// UploadResourceFromBytes 从字节流上传资源
	// 参数:
	//   ctx - 上下文
	//   data - 文件的字节数据
	//   filename - 文件名
	//   memoName - 要关联的 Memo 名称，可选，格式为 "memos/{memo_id}"
	// 返回值:
	//   *v1pb.Attachment - 上传的附件对象
	//   error - 错误信息
	UploadResourceFromBytes(ctx context.Context, data []byte, filename string, memoName string) (*v1pb.Attachment, error)

	// UploadResourceFromReader 从 io.Reader 上传资源
	// 参数:
	//   ctx - 上下文
	//   reader - 包含文件数据的 io.Reader
	//   filename - 文件名
	//   memoName - 要关联的 Memo 名称，可选，格式为 "memos/{memo_id}"
	// 返回值:
	//   *v1pb.Attachment - 上传的附件对象
	//   error - 错误信息
	UploadResourceFromReader(ctx context.Context, reader io.Reader, filename string, memoName string) (*v1pb.Attachment, error)

	// UploadResources 批量上传多个资源
	// 参数:
	//   ctx - 上下文
	//   filePaths - 要上传的文件路径列表
	//   memoName - 要关联的 Memo 名称，可选，格式为 "memos/{memo_id}"
	// 返回值:
	//   []*v1pb.Attachment - 成功上传的附件对象列表
	//   error - 如果有上传失败，返回包含错误信息的 error
	UploadResources(ctx context.Context, filePaths []string, memoName string) ([]*v1pb.Attachment, error)

	// GetAttachment 获取单个附件
	// 参数:
	//   ctx - 上下文
	//   attachmentName - 附件的名称，格式为 "attachments/{attachment_id}"
	// 返回值:
	//   *v1pb.Attachment - 获取的附件对象
	//   error - 错误信息
	GetAttachment(ctx context.Context, attachmentName string) (*v1pb.Attachment, error)

	// DeleteAttachment 删除指定的附件
	// 参数:
	//   ctx - 上下文
	//   attachmentName - 附件的名称，格式为 "attachments/{attachment_id}"
	// 返回值:
	//   error - 错误信息
	DeleteAttachment(ctx context.Context, attachmentName string) error

	// CreateMemoWithResources 创建带资源的 Memo
	// 这是一个便捷方法，会先创建 Memo 然后上传并关联资源
	// 参数:
	//   ctx - 上下文
	//   content - Memo 的内容，支持 Markdown 格式
	//   visibility - Memo 的可见性，默认为 Private
	//   filePaths - 要上传并关联的文件路径列表
	// 返回值:
	//   *v1pb.Memo - 创建的 Memo 对象
	//   []*v1pb.Attachment - 成功上传的附件对象列表
	//   error - 错误信息
	CreateMemoWithResources(ctx context.Context, content string, visibility Visibility, filePaths []string) (*v1pb.Memo, []*v1pb.Attachment, error)
}

// Client Memos API 客户端，实现了 ClientInterface 接口
type Client struct {
	baseURL     string
	accessToken string
	httpClient  *http.Client

	InstanceService   apiv1connect.InstanceServiceClient
	AuthService       apiv1connect.AuthServiceClient
	UserService       apiv1connect.UserServiceClient
	MemoService       apiv1connect.MemoServiceClient
	AttachmentService apiv1connect.AttachmentServiceClient
}

// 确保 Client 实现了 ClientInterface 接口
var _ ClientInterface = (*Client)(nil)

// NewClient 创建 Memos 客户端
// 参数:
//
//	baseURL - Memos 服务器的基础 URL，例如 "https://memos.example.com"
//	accessToken - 访问令牌，用于认证
//
// 返回值:
//
//	*Client - Memos 客户端实例
func NewClient(baseURL, accessToken string) *Client {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	httpClient := &http.Client{
		Transport: &authTransport{
			token:     accessToken,
			transport: http.DefaultTransport,
		},
	}

	return &Client{
		baseURL:     baseURL,
		accessToken: accessToken,
		httpClient:  httpClient,

		InstanceService:   apiv1connect.NewInstanceServiceClient(httpClient, baseURL),
		AuthService:       apiv1connect.NewAuthServiceClient(httpClient, baseURL),
		UserService:       apiv1connect.NewUserServiceClient(httpClient, baseURL),
		MemoService:       apiv1connect.NewMemoServiceClient(httpClient, baseURL),
		AttachmentService: apiv1connect.NewAttachmentServiceClient(httpClient, baseURL),
	}
}

// CreateMemo 创建一个新的 Memo
func (c *Client) CreateMemo(ctx context.Context, content string, visibility Visibility) (*v1pb.Memo, error) {
	if visibility == "" {
		visibility = VisibilityPrivate
	}

	var v v1pb.Visibility
	switch visibility {
	case VisibilityPrivate:
		v = v1pb.Visibility_PRIVATE
	case VisibilityProtected:
		v = v1pb.Visibility_PROTECTED
	case VisibilityPublic:
		v = v1pb.Visibility_PUBLIC
	default:
		return nil, fmt.Errorf("invalid visibility: %s", visibility)
	}

	resp, err := c.MemoService.CreateMemo(ctx, connect.NewRequest(&v1pb.CreateMemoRequest{
		Memo: &v1pb.Memo{
			Content:    content,
			Visibility: v,
		},
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to create memo: %w", err)
	}

	if resp == nil || resp.Msg == nil {
		logger.Error("invalid response", zap.Any("resp", resp))
		return nil, fmt.Errorf("invalid response: nil response or message")
	}

	logger.Info("Memo created successfully", zap.String("memo_name", resp.Msg.Name))
	return resp.Msg, nil
}

// GetMemo 根据名称获取单个 Memo
func (c *Client) GetMemo(ctx context.Context, memoName string) (*v1pb.Memo, error) {
	resp, err := c.MemoService.GetMemo(ctx, connect.NewRequest(&v1pb.GetMemoRequest{
		Name: memoName,
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to get memo: %w", err)
	}

	if resp == nil || resp.Msg == nil {
		logger.Error("invalid response", zap.Any("resp", resp))
		return nil, fmt.Errorf("invalid response: nil response or message")
	}

	return resp.Msg, nil
}

// UpdateMemo 更新现有的 Memo
func (c *Client) UpdateMemo(ctx context.Context, memo *v1pb.Memo) (*v1pb.Memo, error) {
	resp, err := c.MemoService.UpdateMemo(ctx, connect.NewRequest(&v1pb.UpdateMemoRequest{
		Memo: memo,
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to update memo: %w", err)
	}

	if resp == nil || resp.Msg == nil {
		logger.Error("invalid response", zap.Any("resp", resp))
		return nil, fmt.Errorf("invalid response: nil response or message")
	}

	logger.Info("Memo updated successfully", zap.String("memo_name", resp.Msg.Name))
	return resp.Msg, nil
}

// DeleteMemo 删除指定的 Memo
func (c *Client) DeleteMemo(ctx context.Context, memoName string) error {
	_, err := c.MemoService.DeleteMemo(ctx, connect.NewRequest(&v1pb.DeleteMemoRequest{
		Name: memoName,
	}))

	if err != nil {
		return fmt.Errorf("failed to delete memo: %w", err)
	}

	logger.Info("Memo deleted successfully", zap.String("memo_name", memoName))
	return nil
}

// ListMemos 列出 Memos，支持分页
func (c *Client) ListMemos(ctx context.Context, pageSize int32, pageToken string) (*v1pb.ListMemosResponse, error) {
	if pageSize <= 0 {
		pageSize = 50
	}

	resp, err := c.MemoService.ListMemos(ctx, connect.NewRequest(&v1pb.ListMemosRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to list memos: %w", err)
	}

	if resp == nil || resp.Msg == nil {
		logger.Error("invalid response", zap.Any("resp", resp))
		return nil, fmt.Errorf("invalid response: nil response or message")
	}

	return resp.Msg, nil
}

// UploadResource 从文件上传资源到 Memos
func (c *Client) UploadResource(ctx context.Context, filePath string, memoName string) (*v1pb.Attachment, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return c.UploadResourceFromBytes(ctx, bytes, filepath.Base(filePath), memoName)
}

// UploadResourceFromBytes 从字节流上传资源
func (c *Client) UploadResourceFromBytes(ctx context.Context, data []byte, filename string, memoName string) (*v1pb.Attachment, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	contentType := c.detectContentType(data, filename)

	attachment := &v1pb.Attachment{
		Filename: filename,
		Type:     contentType,
		Size:     int64(len(data)),
		Content:  data,
	}

	if memoName != "" {
		attachment.Memo = &memoName
	}

	resp, err := c.AttachmentService.CreateAttachment(ctx, connect.NewRequest(&v1pb.CreateAttachmentRequest{
		Attachment: attachment,
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to create attachment: %w", err)
	}

	if resp == nil || resp.Msg == nil {
		logger.Error("invalid response", zap.Any("resp", resp))
		return nil, fmt.Errorf("invalid response: nil response or message")
	}

	logger.Info("Resource uploaded successfully",
		zap.String("attachment_name", resp.Msg.Name),
		zap.String("filename", resp.Msg.Filename),
	)
	return resp.Msg, nil
}

// UploadResourceFromReader 从 io.Reader 上传资源
func (c *Client) UploadResourceFromReader(ctx context.Context, reader io.Reader, filename string, memoName string) (*v1pb.Attachment, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}

	return c.UploadResourceFromBytes(ctx, data, filename, memoName)
}

// UploadResources 批量上传多个资源
func (c *Client) UploadResources(ctx context.Context, filePaths []string, memoName string) ([]*v1pb.Attachment, error) {
	attachments := make([]*v1pb.Attachment, 0, len(filePaths))
	errors := make([]error, 0)

	for _, filePath := range filePaths {
		attachment, err := c.UploadResource(ctx, filePath, memoName)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to upload %s: %w", filePath, err))
			continue
		}
		attachments = append(attachments, attachment)
	}

	if len(errors) > 0 {
		return attachments, fmt.Errorf("some uploads failed: %v", errors)
	}

	return attachments, nil
}

// GetAttachment 获取单个附件
func (c *Client) GetAttachment(ctx context.Context, attachmentName string) (*v1pb.Attachment, error) {
	resp, err := c.AttachmentService.GetAttachment(ctx, connect.NewRequest(&v1pb.GetAttachmentRequest{
		Name: attachmentName,
	}))

	if err != nil {
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	if resp == nil || resp.Msg == nil {
		logger.Error("invalid response", zap.Any("resp", resp))
		return nil, fmt.Errorf("invalid response: nil response or message")
	}

	return resp.Msg, nil
}

// DeleteAttachment 删除指定的附件
func (c *Client) DeleteAttachment(ctx context.Context, attachmentName string) error {
	_, err := c.AttachmentService.DeleteAttachment(ctx, connect.NewRequest(&v1pb.DeleteAttachmentRequest{
		Name: attachmentName,
	}))

	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	logger.Info("Attachment deleted successfully", zap.String("attachment_name", attachmentName))
	return nil
}

// CreateMemoWithResources 创建带资源的 Memo
func (c *Client) CreateMemoWithResources(ctx context.Context, content string, visibility Visibility, filePaths []string) (*v1pb.Memo, []*v1pb.Attachment, error) {
	memo, err := c.CreateMemo(ctx, content, visibility)
	if err != nil {
		return nil, nil, err
	}

	attachments, err := c.UploadResources(ctx, filePaths, memo.Name)
	if err != nil {
		return memo, attachments, err
	}

	return memo, attachments, nil
}

// detectContentType 检测内容类型
// 优先使用 http.DetectContentType 检测，
// 如果检测失败，则根据文件扩展名从 mimeTypes 映射表中查找，
// 如果还是找不到，则使用 mime.TypeByExtension 尝试，
// 最后默认返回 "application/octet-stream"
func (c *Client) detectContentType(data []byte, filename string) string {
	contentType := http.DetectContentType(data)
	if semicolonIndex := strings.Index(contentType, ";"); semicolonIndex != -1 {
		contentType = strings.TrimSpace(contentType[:semicolonIndex])
	}

	if contentType == "" || contentType == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(filename))
		if mimeType, ok := mimeTypes[ext]; ok {
			return mimeType
		}

		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			if semicolonIndex := strings.Index(mimeType, ";"); semicolonIndex != -1 {
				mimeType = strings.TrimSpace(mimeType[:semicolonIndex])
			}
			return mimeType
		}

		return "application/octet-stream"
	}

	return contentType
}

// authTransport 添加 Authorization header 到所有 HTTP 请求
type authTransport struct {
	token     string
	transport http.RoundTripper
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != "" {
		req.Header.Set("Authorization", "Bearer "+t.token)
	}
	return t.transport.RoundTrip(req)
}
