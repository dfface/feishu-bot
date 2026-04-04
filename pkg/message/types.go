package message

import (
	"context"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// MessageBuilder 消息构建器接口，用于构建各种类型的消息内容。
// 所有具体的消息构建器都必须实现这个接口。
//
// Example:
//
//	builder := NewTextMessageBuilder("Hello, World!")
//	content, err := builder.Build()
//	msgType := builder.MessageType()
type MessageBuilder interface {
	// Build 构建消息内容，返回 JSON 字符串
	Build() (string, error)
	// MessageType 获取消息类型
	MessageType() string
}

// FileUploader 文件上传器接口，提供图片和文件上传功能。
//
// 该接口专注于文件上传，与消息发送功能解耦。
type FileUploader interface {
	// UploadImage 上传图片
	// imagePath: 图片文件路径
	// imageType: 图片类型（message 或 avatar）
	UploadImage(ctx context.Context, imagePath string, imageType ImageType) (string, error)

	// UploadFile 上传文件
	// filePath: 文件路径
	// fileType: 文件类型
	// fileName: 文件名（可选，不传则使用文件路径中的文件名）
	UploadFile(ctx context.Context, filePath string, fileType FileType, fileName string) (string, error)
}

// MessageSender 消息发送器接口，专注于消息发送和表情反应功能。
//
// 该接口设计遵循单一职责原则，只关注消息发送，文件上传由 FileUploader 独立处理。
//
// Example:
//
//	builder := NewImageMessageBuilder(imageKey)
//	sender.SendMessage(ctx, ReceiveIDTypeOpenID, userID, builder)
type MessageSender interface {
	// SendMessage 发送消息
	// receiveIDType: 接收者 ID 类型（open_id, user_id, union_id, email, chat_id）
	// receiveID: 接收者 ID
	// builder: 消息构建器
	SendMessage(ctx context.Context, receiveIDType ReceiveIDType, receiveID string, builder MessageBuilder) (*larkim.CreateMessageResp, error)

	// ReplyMessage 回复消息
	// messageID: 被回复的消息 ID
	// builder: 消息构建器
	ReplyMessage(ctx context.Context, messageID string, builder MessageBuilder) (*larkim.ReplyMessageResp, error)

	// AddReaction 添加表情回复
	// messageID: 消息 ID
	// emojiType: 表情类型
	AddReaction(ctx context.Context, messageID string, emojiType EmojiType) (*larkim.CreateMessageReactionResp, error)
}

// MessageReceiver 消息接收器接口，用于处理接收到的消息。
//
// 该接口负责解析飞书发送的原始事件消息，提取结构化的内容并下载相关资源。
//
// Example:
//
//	processor := NewProcessor(client, logger)
//	msgContent, err := processor.Process(ctx, eventMessage)
//	if err != nil {
//	    // 处理错误
//	}
//	fmt.Println("Message text:", msgContent.Text)
//	for _, resource := range msgContent.Resources {
//	    fmt.Println("Resource:", resource.Type, resource.LocalPath)
//	}
type MessageReceiver interface {
	// Process 处理接收到的飞书事件消息，解析并提取内容。
	// 对于包含资源的消息（图片、文件、音频、视频等），会自动下载到本地。
	// 对于富文本消息，会解析其中的所有元素，包括图片、媒体等资源并下载。
	//
	// 参数:
	//   ctx - 上下文，用于取消操作
	//   msg - 飞书原始事件消息
	//
	// 返回:
	//   *MessageContent - 解析后的结构化消息内容
	//   error - 如果处理失败返回错误
	Process(ctx context.Context, msg *larkim.EventMessage) (*MessageContent, error)
}

// RichTextElement 富文本元素接口
type RichTextElement interface {
	Tag() string
}

// RichTextText 文本元素
type RichTextText struct {
	Text     string   `json:"text"`
	UnEscape bool     `json:"un_escape,omitempty"`
	Style    []string `json:"style,omitempty"`
}

func (e *RichTextText) Tag() string { return "text" }

// RichTextA 链接元素
type RichTextA struct {
	Text     string   `json:"text"`
	Href     string   `json:"href"`
	UnEscape bool     `json:"un_escape,omitempty"`
	Style    []string `json:"style,omitempty"`
}

func (e *RichTextA) Tag() string { return "a" }

// RichTextAt @元素
type RichTextAt struct {
	UserId   string   `json:"user_id"`
	UserName string   `json:"user_name,omitempty"`
	Style    []string `json:"style,omitempty"`
}

func (e *RichTextAt) Tag() string { return "at" }

// RichTextImg 图片元素
type RichTextImg struct {
	ImageKey string `json:"image_key"`
}

func (e *RichTextImg) Tag() string { return "img" }

// RichTextMedia 视频元素
type RichTextMedia struct {
	FileKey  string `json:"file_key"`
	ImageKey string `json:"image_key,omitempty"`
}

func (e *RichTextMedia) Tag() string { return "media" }

// RichTextEmotion 表情元素
type RichTextEmotion struct {
	EmojiType string `json:"emoji_type"`
}

func (e *RichTextEmotion) Tag() string { return "emotion" }

// RichTextHr 分割线元素
type RichTextHr struct{}

func (e *RichTextHr) Tag() string { return "hr" }

// RichTextCodeBlock 代码块元素
type RichTextCodeBlock struct {
	Language string `json:"language,omitempty"`
	Text     string `json:"text"`
}

func (e *RichTextCodeBlock) Tag() string { return "code_block" }

// RichTextMd Markdown元素
type RichTextMd struct {
	Text string `json:"text"`
}

func (e *RichTextMd) Tag() string { return "md" }
