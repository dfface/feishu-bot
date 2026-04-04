package message

import (
	"encoding/json"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// === 基础 Builder 实现 ===

// TextMessageBuilder 文本消息构建器
type TextMessageBuilder struct {
	text string
}

// NewTextMessageBuilder 创建文本消息构建器
func NewTextMessageBuilder(text string) *TextMessageBuilder {
	return &TextMessageBuilder{text: text}
}

// Build 构建文本消息
func (b *TextMessageBuilder) Build() (string, error) {
	content := map[string]string{"text": b.text}
	jsonContent, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(jsonContent), nil
}

// MessageType 返回文本消息类型
func (b *TextMessageBuilder) MessageType() string {
	return larkim.MsgTypeText
}

// RichTextMessageBuilder 富文本消息构建器
type RichTextMessageBuilder struct {
	title   string
	content [][]RichTextElement
}

// NewRichTextMessageBuilder 创建富文本消息构建器
func NewRichTextMessageBuilder() *RichTextMessageBuilder {
	return &RichTextMessageBuilder{
		content: make([][]RichTextElement, 0),
	}
}

// SetTitle 设置标题
func (b *RichTextMessageBuilder) SetTitle(title string) *RichTextMessageBuilder {
	b.title = title
	return b
}

// AddText 添加文本元素到当前行
func (b *RichTextMessageBuilder) AddText(text string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextText{Text: text, UnEscape: false})
}

// AddTextWithStyle 添加带样式的文本元素到当前行
func (b *RichTextMessageBuilder) AddTextWithStyle(text string, styles ...string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextText{Text: text, UnEscape: false, Style: styles})
}

// AddBoldText 添加加粗文本元素到当前行
func (b *RichTextMessageBuilder) AddBoldText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, "bold")
}

// AddItalicText 添加斜体文本元素到当前行
func (b *RichTextMessageBuilder) AddItalicText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, "italic")
}

// AddUnderlineText 添加下划线文本元素到当前行
func (b *RichTextMessageBuilder) AddUnderlineText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, "underline")
}

// AddStrikethroughText 添加删除线文本元素到当前行
func (b *RichTextMessageBuilder) AddStrikethroughText(text string) *RichTextMessageBuilder {
	return b.AddTextWithStyle(text, "lineThrough")
}

// AddLink 添加链接元素到当前行
func (b *RichTextMessageBuilder) AddLink(text, href string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextA{Text: text, Href: href, UnEscape: false})
}

// AddLinkWithStyle 添加带样式的链接元素到当前行
func (b *RichTextMessageBuilder) AddLinkWithStyle(text, href string, styles ...string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextA{Text: text, Href: href, UnEscape: false, Style: styles})
}

// AddAt 添加@用户元素到当前行
func (b *RichTextMessageBuilder) AddAt(userID, userName string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextAt{UserId: userID, UserName: userName})
}

// AddAtWithStyle 添加带样式的@用户元素到当前行
func (b *RichTextMessageBuilder) AddAtWithStyle(userID, userName string, styles ...string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextAt{UserId: userID, UserName: userName, Style: styles})
}

// AddAtAll 添加@所有人元素到当前行
func (b *RichTextMessageBuilder) AddAtAll() *RichTextMessageBuilder {
	return b.AddElement(&RichTextAt{UserId: "all"})
}

// AddImage 添加图片元素到当前行
func (b *RichTextMessageBuilder) AddImage(imageKey string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextImg{ImageKey: imageKey})
}

// AddMedia 添加视频元素到当前行
func (b *RichTextMessageBuilder) AddMedia(fileKey string, imageKey ...string) *RichTextMessageBuilder {
	media := &RichTextMedia{FileKey: fileKey}
	if len(imageKey) > 0 {
		media.ImageKey = imageKey[0]
	}
	return b.AddElement(media)
}

// AddEmotion 添加表情元素到当前行
func (b *RichTextMessageBuilder) AddEmotion(emojiType EmojiType) *RichTextMessageBuilder {
	return b.AddElement(&RichTextEmotion{EmojiType: string(emojiType)})
}

// AddHr 添加分割线元素到当前行
func (b *RichTextMessageBuilder) AddHr() *RichTextMessageBuilder {
	return b.AddElement(&RichTextHr{})
}

// AddCodeBlock 添加代码块元素到当前行
func (b *RichTextMessageBuilder) AddCodeBlock(text string, language ...string) *RichTextMessageBuilder {
	codeBlock := &RichTextCodeBlock{Text: text}
	if len(language) > 0 {
		codeBlock.Language = language[0]
	}
	return b.AddElement(codeBlock)
}

// AddMd 添加Markdown元素到当前行
func (b *RichTextMessageBuilder) AddMd(text string) *RichTextMessageBuilder {
	return b.AddElement(&RichTextMd{Text: text})
}

// AddElement 添加元素到当前行
func (b *RichTextMessageBuilder) AddElement(element RichTextElement) *RichTextMessageBuilder {
	if len(b.content) == 0 {
		b.content = append(b.content, make([]RichTextElement, 0))
	}
	lastLine := len(b.content) - 1
	b.content[lastLine] = append(b.content[lastLine], element)
	return b
}

// NewLine 开始新的一行
func (b *RichTextMessageBuilder) NewLine() *RichTextMessageBuilder {
	b.content = append(b.content, make([]RichTextElement, 0))
	return b
}

// AddLine 添加一行内容
func (b *RichTextMessageBuilder) AddLine(elements ...RichTextElement) *RichTextMessageBuilder {
	b.content = append(b.content, elements)
	return b
}

// Build 构建富文本消息
func (b *RichTextMessageBuilder) Build() (string, error) {
	post := make(map[string]interface{})
	zhCN := make(map[string]interface{})
	
	if b.title != "" {
		zhCN["title"] = b.title
	}
	
	content := make([][]map[string]interface{}, 0, len(b.content))
	
	for _, line := range b.content {
		lineElements := make([]map[string]interface{}, 0, len(line))
		for _, elem := range line {
			elemMap := make(map[string]interface{})
			
			switch e := elem.(type) {
			case *RichTextText:
				elemMap["tag"] = "text"
				elemMap["text"] = e.Text
				if e.UnEscape {
					elemMap["un_escape"] = e.UnEscape
				}
				if len(e.Style) > 0 {
					elemMap["style"] = e.Style
				}
			case *RichTextA:
				elemMap["tag"] = "a"
				elemMap["text"] = e.Text
				elemMap["href"] = e.Href
				if e.UnEscape {
					elemMap["un_escape"] = e.UnEscape
				}
				if len(e.Style) > 0 {
					elemMap["style"] = e.Style
				}
			case *RichTextAt:
				elemMap["tag"] = "at"
				elemMap["user_id"] = e.UserId
				if e.UserName != "" {
					elemMap["user_name"] = e.UserName
				}
				if len(e.Style) > 0 {
					elemMap["style"] = e.Style
				}
			case *RichTextImg:
				elemMap["tag"] = "img"
				elemMap["image_key"] = e.ImageKey
			case *RichTextMedia:
				elemMap["tag"] = "media"
				elemMap["file_key"] = e.FileKey
				if e.ImageKey != "" {
					elemMap["image_key"] = e.ImageKey
				}
			case *RichTextEmotion:
				elemMap["tag"] = "emotion"
				elemMap["emoji_type"] = e.EmojiType
			case *RichTextHr:
				elemMap["tag"] = "hr"
			case *RichTextCodeBlock:
				elemMap["tag"] = "code_block"
				elemMap["text"] = e.Text
				if e.Language != "" {
					elemMap["language"] = e.Language
				}
			case *RichTextMd:
				elemMap["tag"] = "md"
				elemMap["text"] = e.Text
			}
			
			if len(elemMap) > 0 {
				lineElements = append(lineElements, elemMap)
			}
		}
		if len(lineElements) > 0 {
			content = append(content, lineElements)
		}
	}
	
	zhCN["content"] = content
	post["zh_cn"] = zhCN
	
	jsonContent, err := json.Marshal(post)
	if err != nil {
		return "", err
	}
	return string(jsonContent), nil
}

// MessageType 返回富文本消息类型
func (b *RichTextMessageBuilder) MessageType() string {
	return larkim.MsgTypePost
}

// ImageMessageBuilder 图片消息构建器
type ImageMessageBuilder struct {
	imageKey string
}

// NewImageMessageBuilder 创建图片消息构建器
func NewImageMessageBuilder(imageKey string) *ImageMessageBuilder {
	return &ImageMessageBuilder{imageKey: imageKey}
}

// Build 构建图片消息
func (b *ImageMessageBuilder) Build() (string, error) {
	msgImage := larkim.MessageImage{ImageKey: b.imageKey}
	return msgImage.String()
}

// MessageType 返回图片消息类型
func (b *ImageMessageBuilder) MessageType() string {
	return larkim.MsgTypeImage
}

// FileMessageBuilder 文件消息构建器
type FileMessageBuilder struct {
	fileKey  string
	fileName string
}

// NewFileMessageBuilder 创建文件消息构建器
func NewFileMessageBuilder(fileKey, fileName string) *FileMessageBuilder {
	return &FileMessageBuilder{
		fileKey:  fileKey,
		fileName: fileName,
	}
}

// Build 构建文件消息
func (b *FileMessageBuilder) Build() (string, error) {
	msgFile := larkim.MessageFile{FileKey: b.fileKey}
	return msgFile.String()
}

// MessageType 返回文件消息类型
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
