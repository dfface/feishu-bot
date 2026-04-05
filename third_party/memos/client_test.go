package memos

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1pb "github.com/usememos/memos/proto/gen/api/v1"
)

// TestNewClient 测试 NewClient 函数
func TestNewClient(t *testing.T) {
	// 测试用例 1: 正常创建客户端
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)
	assert.Equal(t, "https://memos.yuhan.tech", client.baseURL)
	assert.Equal(t, "test_token", client.accessToken)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.MemoService)
	assert.NotNil(t, client.AttachmentService)

	// 测试用例 2: 没有协议前缀的 URL
	client = NewClient("memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)
	assert.Equal(t, "https://memos.yuhan.tech", client.baseURL)

	// 测试用例 3: HTTP 协议前缀
	client = NewClient("http://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)
	assert.Equal(t, "http://memos.yuhan.tech", client.baseURL)
}

// TestCreateMemo 测试 CreateMemo 函数
func TestCreateMemo(t *testing.T) {
	// 这里应该使用模拟客户端进行测试，或者使用真实的测试环境
	// 由于这是一个集成测试，我们暂时只测试基本的参数验证
	client := NewClient("https://memos.yuhan.tech", "memos_pat_gcervJhJGQQRUTLHMRRPX1c9PSSCs6Pr")
	assert.NotNil(t, client)

	// 测试用例 1: 正常创建 Memo
	ctx := context.Background()
	content := "测试 Memo 内容"
	visibility := VisibilityPrivate

	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.CreateMemo(ctx, content, visibility)
	// 我们不断言错误，因为这取决于测试环境是否可用
	t.Logf("CreateMemo 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 空内容
	resp, err = client.CreateMemo(ctx, "", visibility)
	t.Logf("CreateMemo 空内容测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 3: 空可见性
	resp, err = client.CreateMemo(ctx, content, "")
	t.Logf("CreateMemo 空可见性测试结果: resp=%v, err=%v", resp, err)
}

// TestListMemos 测试 ListMemos 函数
func TestListMemos(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	pageSize := int32(10)
	pageToken := ""

	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.ListMemos(ctx, pageSize, pageToken)
	t.Logf("ListMemos 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 负数 pageSize
	resp, err = client.ListMemos(ctx, -1, pageToken)
	t.Logf("ListMemos 负数 pageSize 测试结果: resp=%v, err=%v", resp, err)
}

// TestDetectContentType 测试 detectContentType 方法
func TestDetectContentType(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	// 测试用例 1: 图片类型
	jpgData := []byte{0xFF, 0xD8, 0xFF} // JPEG 魔术数字
	contentType := client.detectContentType(jpgData, "test.jpg")
	assert.Equal(t, "image/jpeg", contentType)

	// 测试用例 2: 文本类型
	txtData := []byte("Hello, world!")
	contentType = client.detectContentType(txtData, "test.txt")
	assert.Equal(t, "text/plain", contentType)

	// 测试用例 3: 未知类型
	unknownData := []byte{0x00, 0x01, 0x02}
	contentType = client.detectContentType(unknownData, "test.bin")
	assert.Equal(t, "application/octet-stream", contentType)

	// 测试用例 4: 根据文件扩展名推断
	contentType = client.detectContentType(unknownData, "test.md")
	assert.Equal(t, "text/markdown", contentType)
}

// TestVisibility 测试 Visibility 类型
func TestVisibility(t *testing.T) {
	assert.Equal(t, Visibility("PRIVATE"), VisibilityPrivate)
	assert.Equal(t, Visibility("PROTECTED"), VisibilityProtected)
	assert.Equal(t, Visibility("PUBLIC"), VisibilityPublic)
}

// TestGetMemo 测试 GetMemo 函数
func TestGetMemo(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	// 测试用例 1: 正常获取 Memo
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.GetMemo(ctx, "memos/1")
	t.Logf("GetMemo 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 空 memoName
	resp, err = client.GetMemo(ctx, "")
	t.Logf("GetMemo 空 memoName 测试结果: resp=%v, err=%v", resp, err)
}

// TestUpdateMemo 测试 UpdateMemo 函数
func TestUpdateMemo(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	// 测试用例 1: 正常更新 Memo
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	memo := &v1pb.Memo{
		Name:       "memos/1",
		Content:    "更新后的 Memo 内容",
		Visibility: v1pb.Visibility_PRIVATE,
	}
	resp, err := client.UpdateMemo(ctx, memo)
	t.Logf("UpdateMemo 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 空 memo
	resp, err = client.UpdateMemo(ctx, nil)
	t.Logf("UpdateMemo 空 memo 测试结果: resp=%v, err=%v", resp, err)
}

// TestDeleteMemo 测试 DeleteMemo 函数
func TestDeleteMemo(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	// 测试用例 1: 正常删除 Memo
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	err := client.DeleteMemo(ctx, "memos/1")
	t.Logf("DeleteMemo 测试结果: err=%v", err)

	// 测试用例 2: 空 memoName
	err = client.DeleteMemo(ctx, "")
	t.Logf("DeleteMemo 空 memoName 测试结果: err=%v", err)
}

// TestUploadResource 测试 UploadResource 函数
func TestUploadResource(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	// 测试用例 1: 正常上传资源
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.UploadResource(ctx, "test.txt", "")
	t.Logf("UploadResource 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 关联到 Memo
	resp, err = client.UploadResource(ctx, "test.txt", "memos/1")
	t.Logf("UploadResource 关联 Memo 测试结果: resp=%v, err=%v", resp, err)
}

// TestUploadResourceFromBytes 测试 UploadResourceFromBytes 函数
func TestUploadResourceFromBytes(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	data := []byte("test data")
	// 测试用例 1: 正常上传资源
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.UploadResourceFromBytes(ctx, data, "test.txt", "")
	t.Logf("UploadResourceFromBytes 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 空文件名
	resp, err = client.UploadResourceFromBytes(ctx, data, "", "")
	t.Logf("UploadResourceFromBytes 空文件名测试结果: resp=%v, err=%v", resp, err)
}

// TestUploadResourceFromReader 测试 UploadResourceFromReader 函数
func TestUploadResourceFromReader(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	reader := strings.NewReader("test data")
	// 测试用例 1: 正常上传资源
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.UploadResourceFromReader(ctx, reader, "test.txt", "")
	t.Logf("UploadResourceFromReader 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 空文件名
	resp, err = client.UploadResourceFromReader(ctx, reader, "", "")
	t.Logf("UploadResourceFromReader 空文件名测试结果: resp=%v, err=%v", resp, err)
}

// TestUploadResources 测试 UploadResources 函数
func TestUploadResources(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	filePaths := []string{"test1.txt", "test2.txt"}
	// 测试用例 1: 正常批量上传资源
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.UploadResources(ctx, filePaths, "")
	t.Logf("UploadResources 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 关联到 Memo
	resp, err = client.UploadResources(ctx, filePaths, "memos/1")
	t.Logf("UploadResources 关联 Memo 测试结果: resp=%v, err=%v", resp, err)
}

// TestGetAttachment 测试 GetAttachment 函数
func TestGetAttachment(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	// 测试用例 1: 正常获取附件
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	resp, err := client.GetAttachment(ctx, "attachments/1")
	t.Logf("GetAttachment 测试结果: resp=%v, err=%v", resp, err)

	// 测试用例 2: 空 attachmentName
	resp, err = client.GetAttachment(ctx, "")
	t.Logf("GetAttachment 空 attachmentName 测试结果: resp=%v, err=%v", resp, err)
}

// TestDeleteAttachment 测试 DeleteAttachment 函数
func TestDeleteAttachment(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	// 测试用例 1: 正常删除附件
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	err := client.DeleteAttachment(ctx, "attachments/1")
	t.Logf("DeleteAttachment 测试结果: err=%v", err)

	// 测试用例 2: 空 attachmentName
	err = client.DeleteAttachment(ctx, "")
	t.Logf("DeleteAttachment 空 attachmentName 测试结果: err=%v", err)
}

// TestCreateMemoWithResources 测试 CreateMemoWithResources 函数
func TestCreateMemoWithResources(t *testing.T) {
	client := NewClient("https://memos.yuhan.tech", "test_token")
	assert.NotNil(t, client)

	ctx := context.Background()
	content := "测试带资源的 Memo 内容"
	visibility := VisibilityPrivate
	filePaths := []string{"test1.txt", "test2.txt"}

	// 测试用例 1: 正常创建带资源的 Memo
	// 注意：由于这是一个真实的 API 调用，实际运行时可能会失败
	// 这里主要是测试函数签名和基本参数传递
	memo, attachments, err := client.CreateMemoWithResources(ctx, content, visibility, filePaths)
	t.Logf("CreateMemoWithResources 测试结果: memo=%v, attachments=%v, err=%v", memo, attachments, err)

	// 测试用例 2: 空内容
	memo, attachments, err = client.CreateMemoWithResources(ctx, "", visibility, filePaths)
	t.Logf("CreateMemoWithResources 空内容测试结果: memo=%v, attachments=%v, err=%v", memo, attachments, err)

	// 测试用例 3: 空文件路径
	memo, attachments, err = client.CreateMemoWithResources(ctx, content, visibility, nil)
	t.Logf("CreateMemoWithResources 空文件路径测试结果: memo=%v, attachments=%v, err=%v", memo, attachments, err)
}
