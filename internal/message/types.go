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
	Process(ctx context.Context, msg *larkim.P2MessageReceiveV1) (*MessageContent, error)
}

// RichTextElement 富文本元素
// 统一的富文本元素结构，用于发送和接收富文本消息
// 支持所有富文本元素类型：text、a、at、img、media、emotion、hr、code_block、md
type RichTextElement struct {
	Tag       string   `json:"tag"`                  // 元素类型：text、a、at、img、media、emotion、hr、code_block、md
	Text      string   `json:"text,omitempty"`       // 文本内容，用于 text、a、code_block、md 元素
	Href      string   `json:"href,omitempty"`       // 链接地址，用于 a 元素
	UserId    string   `json:"user_id,omitempty"`    // 用户 ID，用于 at 元素（可设置为 "all" 表示 @所有人）
	UserName  string   `json:"user_name,omitempty"`  // 用户名称，用于 at 元素
	ImageKey  string   `json:"image_key,omitempty"`  // 图片 key，用于 img 元素
	Width     int      `json:"width,omitempty"`      // 图片宽度，用于 img 元素
	Height    int      `json:"height,omitempty"`     // 图片高度，用于 img 元素
	FileType  string   `json:"file_type,omitempty"`  // 文件类型，用于 media 元素
	FileKey   string   `json:"file_key,omitempty"`   // 文件 key，用于 media 元素
	FileName  string   `json:"file_name,omitempty"`  // 文件名称，用于 media 元素
	Duration  int      `json:"duration,omitempty"`   // 媒体时长，用于 media 元素
	EmojiType string   `json:"emoji_type,omitempty"` // 表情类型，用于 emotion 元素，如 "Lark_Emoji_Facepalm_0"
	Style     []string `json:"style,omitempty"`      // 样式列表，支持：bold（加粗）、italic（斜体）、underline（下划线）、lineThrough（删除线）
	Content   string   `json:"content,omitempty"`    // 内容，用于 code_block、md 等元素
	Language  string   `json:"language,omitempty"`   // 代码语言，用于 code_block 元素，如 "go"、"python"、"javascript" 等
	UnEscape  bool     `json:"un_escape,omitempty"`  // 是否不转义 HTML 特殊字符，默认为 false（即转义），用于 text、a 元素
}

// RichTextContent 富文本内容
// 存储完整的富文本消息结构，包括标题和所有元素
type RichTextContent struct {
	Title   string               `json:"title"`   // 富文本标题
	Content [][]*RichTextElement `json:"content"` // 内容行列表，每行包含多个元素
}

// MessageContent 消息内容
// 存储解析后的结构化消息内容，包含各种类型消息的详细信息
type MessageContent struct {
	ID            string                 // 消息 ID，用于唯一标识消息
	SenderID      string                 // 发送者 ID，用于关联消息和用户
	Type          MessageType            // 消息类型：text、post、image、file、audio、media、sticker、interactive、share_chat、share_user、system、todo、vote
	Text          string                 // 消息文本内容，对于文本消息是完整内容，对于其他消息是摘要
	RawContent    *string                // 原始消息内容 JSON 字符串
	RichText      *RichTextContent       // 解析后的富文本结构，仅当消息类型为 post 时有效
	Resources     []Resource             // 消息中包含的资源列表（图片、文件、音频、视频等）
	Location      *Location              // 位置信息，仅当消息类型为 location 时有效
	Sticker       *Sticker               // 表情贴纸信息，仅当消息类型为 sticker 时有效
	Interactive   map[string]interface{} // 交互式卡片内容，仅当消息类型为 interactive 时有效
	ShareChat     *ShareChat             // 分享的群聊信息，仅当消息类型为 share_chat 时有效
	ShareUser     *ShareUser             // 分享的用户信息，仅当消息类型为 share_user 时有效
	SystemMessage *SystemMessage         // 系统消息内容，仅当消息类型为 system 时有效
	Todo          *Todo                  // 待办事项内容，仅当消息类型为 todo 时有效
	Vote          *Vote                  // 投票内容，仅当消息类型为 vote 时有效
}

// Resource 资源信息
// 存储消息中包含的资源（图片、文件、音频、视频等）的详细信息
type Resource struct {
	Type       ResourceType // 资源类型：image、file、audio、media
	FileKey    string       // 文件 key，用于下载和重新上传
	FileName   string       // 文件名称
	ImageKey   string       // 图片 key，用于图片资源
	MessageID  string       // 所属消息的 ID，用于下载资源
	Duration   int          // 媒体时长（秒），用于音频和视频资源
	LocalPath  string       // 本地文件路径，下载后存储的位置
	Downloaded bool         // 是否已下载到本地
}

// Location 位置信息
// 存储位置消息的详细信息
type Location struct {
	Name      string // 位置名称
	Longitude string // 经度
	Latitude  string // 纬度
}

// Sticker 表情包信息
// 存储表情包消息的详细信息
type Sticker struct {
	FileKey string // 表情文件 key
}

// ShareChat 群组名片信息
// 存储分享群组消息的详细信息
type ShareChat struct {
	ChatID string // 群组 ID
}

// ShareUser 人名片信息
// 存储分享用户消息的详细信息
type ShareUser struct {
	UserID string // 用户 ID
}

// SystemMessage 系统消息信息
// 存储系统消息的详细信息
type SystemMessage struct {
	Template string            // 消息模板
	Params   map[string]string // 模板参数
}

// Todo 待办任务信息
// 存储待办任务消息的详细信息
type Todo struct {
	TaskID string // 待办任务 ID
}

// Vote 投票信息
// 存储投票消息的详细信息
type Vote struct {
	VoteID string // 投票 ID
}
