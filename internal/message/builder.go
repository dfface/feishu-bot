package message

import (
	"encoding/json"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// === 基础 Builder 实现 ===

// TextMessageBuilder 文本消息构建器
// 用于构建简单的文本消息内容
type TextMessageBuilder struct {
	text string
}

// NewTextMessageBuilder 创建文本消息构建器
//
// 参数:
//
//	text - 要发送的文本内容
//
// 返回:
//
//	*TextMessageBuilder - 初始化好的文本消息构建器
func NewTextMessageBuilder(text string) *TextMessageBuilder {
	return &TextMessageBuilder{text: text}
}

// Build 构建文本消息
// 将文本内容序列化为 JSON 格式的消息内容
//
// 返回:
//
//	string - JSON 格式的文本消息内容
//	error - 序列化失败时返回错误
func (b *TextMessageBuilder) Build() (string, error) {
	content := map[string]string{"text": b.text}
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(jsonContent), nil
}

// MessageType 返回文本消息类型
//
// 返回:
//
//	string - 固定返回 "text"
func (b *TextMessageBuilder) MessageType() string {
	return larkim.MsgTypeText
}

type RichTextParagraph []*RichTextElement

// RichTextMessageBuilder 富文本消息构建器
// 用于构建复杂的富文本消息，支持文本、链接、@用户、图片、媒体、表情等多种元素
type RichTextMessageBuilder struct {
	title   string
	content []RichTextParagraph
}

// NewRichTextMessageBuilder 创建富文本消息构建器
//
// 返回:
//
//	*RichTextMessageBuilder - 初始化好的富文本消息构建器
func NewRichTextMessageBuilder() *RichTextMessageBuilder {
	return &RichTextMessageBuilder{
		content: make([]RichTextParagraph, 0),
	}
}

// SetTitle 设置富文本消息的标题
//
// 参数:
//
//	title - 富文本标题
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) SetTitle(title string) *RichTextMessageBuilder {
	b.title = title
	return b
}

// AddText 添加文本元素到当前行
//
// 参数:
//
//	text - 文本内容
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddText(text string) *RichTextMessageBuilder {
	if text == "" {
		// text 必填
		text = " "
	}
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagText), Text: text, UnEscape: false})
}

// AddTextWithStyle 添加带样式的文本元素到当前行
//
// 参数:
//
//	text - 文本内容
//	styles - 样式列表，支持：bold（加粗）、italic（斜体）、underline（下划线）、lineThrough（删除线）
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddTextWithStyle(text string, styles ...string) *RichTextMessageBuilder {
	if text == "" {
		// text 必填
		text = " "
	}
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagText), Text: text, UnEscape: false, Style: styles})
}

// AddBoldText 添加加粗文本元素到当前行
//
// 参数:
//
//	text - 文本内容
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddBoldText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, string(TextStyleBold))
}

// AddItalicText 添加斜体文本元素到当前行
//
// 参数:
//
//	text - 文本内容
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddItalicText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, string(TextStyleItalic))
}

// AddUnderlineText 添加下划线文本元素到当前行
//
// 参数:
//
//	text - 文本内容
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddUnderlineText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, string(TextStyleUnderline))
}

// AddStrikethroughText 添加删除线文本元素到当前行
//
// 参数:
//
//	text - 文本内容
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddStrikethroughText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, string(TextStyleLineThrough))
}

// AddLink 添加链接元素到当前行
//
// 参数:
//
//	text - 链接显示文本
//	href - 链接地址
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddLink(text, href string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagA), Text: text, Href: href, UnEscape: false})
}

// AddLinkWithStyle 添加带样式的链接元素到当前行
//
// 参数:
//
//	text - 链接显示文本
//	href - 链接地址
//	styles - 样式列表，支持：bold（加粗）、italic（斜体）、underline（下划线）、lineThrough（删除线）
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddLinkWithStyle(text, href string, styles ...string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagA), Text: text, Href: href, UnEscape: false, Style: styles})
}

// AddAt 添加@用户元素到当前行
//
// 参数:
//
//	userID - 用户 ID（open_id、user_id、union_id 等）
//	userName - 用户名称（可选，用于显示）
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddAt(userID, userName string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagAt), UserId: userID, UserName: userName})
}

// AddAtWithStyle 添加带样式的@用户元素到当前行
//
// 参数:
//
//	userID - 用户 ID（open_id、user_id、union_id 等）
//	userName - 用户名称（可选，用于显示）
//	styles - 样式列表，支持：bold（加粗）、italic（斜体）、underline（下划线）、lineThrough（删除线）
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddAtWithStyle(userID, userName string, styles ...string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagAt), UserId: userID, UserName: userName, Style: styles})
}

// AddAtAll 添加@所有人元素到当前行
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddAtAll() *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagAt), UserId: "all"})
}

// AddImage 添加图片元素到当前行
//
// 参数:
//
//	imageKey - 图片 key（通过 UploadImage 获取）
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddImage(imageKey string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagImg), ImageKey: imageKey})
}

// AddMedia 添加视频/媒体元素到当前行
//
// 参数:
//
//	fileKey - 文件 key（通过 UploadFile 获取）
//	imageKey - 可选，封面图片 key
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddMedia(fileKey string, imageKey ...string) *RichTextMessageBuilder {
	elem := &RichTextElement{Tag: string(RichTextTagMedia), FileKey: fileKey}
	if len(imageKey) > 0 {
		elem.ImageKey = imageKey[0]
	}
	return b.AddElement(elem)
}

// AddEmotion 添加表情元素到当前行
//
// 参数:
//
//	emojiType - 表情类型，使用预定义的 EmojiType 常量
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddEmotion(emojiType EmojiType) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagEmotion), EmojiType: string(emojiType)})
}

// AddHr 添加分割线元素到当前行
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddHr() *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagHr)})
}

// AddCodeBlock 添加代码块元素到当前行
//
// 参数:
//
//	text - 代码内容
//	language - 可选，代码语言（如 "go"、"python"、"javascript" 等）
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddCodeBlock(text string, language ...string) *RichTextMessageBuilder {
	elem := &RichTextElement{Tag: string(RichTextTagCodeBlock), Text: text}
	if len(language) > 0 {
		elem.Language = language[0]
	}
	return b.AddElement(elem)
}

// AddMd 添加Markdown元素到当前行
//
// 参数:
//
//	text - Markdown 文本内容
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddMd(text string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextElement{Tag: string(RichTextTagMd), Text: text})
}

// AddElement 添加元素到当前段落
// 如果当前没有段落，会自动创建一个段落
//
// 参数:
//
//	element - 要添加的富文本元素
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddElement(element *RichTextElement) *RichTextMessageBuilder {
	if len(b.content) == 0 {
		b.content = append(b.content, make(RichTextParagraph, 0))
	}
	lastParagraph := len(b.content) - 1
	b.content[lastParagraph] = append(b.content[lastParagraph], element)
	return b
}

// NewLine 在当前段落中添加一个换行符
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) NewLine() *RichTextMessageBuilder {
	return b.AddText("\n")
}

// NewParagraph 开始新的一个段落
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) NewParagraph() *RichTextMessageBuilder {
	b.content = append(b.content, make(RichTextParagraph, 0))
	return b
}

// AddParagraph 添加一个段落内容
//
// 参数:
//
//	elements - 该段落包含的富文本元素列表
//
// 返回:
//
//	*RichTextMessageBuilder - 构建器自身，支持链式调用
func (b *RichTextMessageBuilder) AddParagraph(elements ...*RichTextElement) *RichTextMessageBuilder {
	b.content = append(b.content, elements)
	return b
}

// Build 构建富文本消息
// 将富文本内容序列化为 JSON 格式
//
// 返回:
//
//	string - JSON 格式的富文本消息内容
//	error - 序列化失败时返回错误
func (b *RichTextMessageBuilder) Build() (string, error) {
	post := make(map[string]interface{})
	zhCN := make(map[string]interface{})

	if b.title != "" {
		zhCN["title"] = b.title
	}

	zhCN["content"] = b.content
	post["zh_cn"] = zhCN

	jsonContent, err := json.Marshal(post)
	if err != nil {
		return "", err
	}
	return string(jsonContent), nil
}

// MessageType 返回富文本消息类型
//
// 返回:
//
//	string - 固定返回 "post"
func (b *RichTextMessageBuilder) MessageType() string {
	return larkim.MsgTypePost
}

// ImageMessageBuilder 图片消息构建器
// 用于构建单独的图片消息
type ImageMessageBuilder struct {
	imageKey string
}

// NewImageMessageBuilder 创建图片消息构建器
//
// 参数:
//
//	imageKey - 图片 key（通过 UploadImage 获取）
//
// 返回:
//
//	*ImageMessageBuilder - 初始化好的图片消息构建器
func NewImageMessageBuilder(imageKey string) *ImageMessageBuilder {
	return &ImageMessageBuilder{imageKey: imageKey}
}

// Build 构建图片消息
//
// 返回:
//
//	string - JSON 格式的图片消息内容
//	error - 序列化失败时返回错误
func (b *ImageMessageBuilder) Build() (string, error) {
	msgImage := larkim.MessageImage{ImageKey: b.imageKey}
	return msgImage.String()
}

// MessageType 返回图片消息类型
//
// 返回:
//
//	string - 固定返回 "image"
func (b *ImageMessageBuilder) MessageType() string {
	return larkim.MsgTypeImage
}

// FileMessageBuilder 文件消息构建器
// 用于构建单独的文件消息
type FileMessageBuilder struct {
	fileKey  string
	fileName string
}

// NewFileMessageBuilder 创建文件消息构建器
//
// 参数:
//
//	fileKey - 文件 key（通过 UploadFile 获取）
//	fileName - 文件名
//
// 返回:
//
//	*FileMessageBuilder - 初始化好的文件消息构建器
func NewFileMessageBuilder(fileKey, fileName string) *FileMessageBuilder {
	return &FileMessageBuilder{
		fileKey:  fileKey,
		fileName: fileName,
	}
}

// Build 构建文件消息
//
// 返回:
//
//	string - JSON 格式的文件消息内容
//	error - 序列化失败时返回错误
func (b *FileMessageBuilder) Build() (string, error) {
	msgFile := larkim.MessageFile{FileKey: b.fileKey}
	return msgFile.String()
}

// MessageType 返回文件消息类型
//
// 返回:
//
//	string - 固定返回 "file"
func (b *FileMessageBuilder) MessageType() string {
	return larkim.MsgTypeFile
}

// === 接口验证 ===

var (
	_ MessageBuilder = (*TextMessageBuilder)(nil)
	_ MessageBuilder = (*RichTextMessageBuilder)(nil)
	_ MessageBuilder = (*ImageMessageBuilder)(nil)
	_ MessageBuilder = (*FileMessageBuilder)(nil)
)
