package converter

import (
	"fmt"
	"strings"

	"github.com/dfface/feishu-bot/pkg/message"
	"github.com/dfface/feishu-bot/third_party/memos"
)

// MemosConverter Memos 转换器接口
// 负责将 MessageContent 转换为 Memos 所需的格式
//
// Example:
//
// converter := NewMemosConverter()
// content, filePaths, err := converter.ConvertMessageContent(ctx, msgContent)
//
//	if err != nil {
//	    // 处理错误
//	}
//
// memo, attachments, err := memosClient.CreateMemoWithResources(ctx, content, memos.VisibilityPrivate, filePaths)
type MemosConverter interface {
	// ConvertMessageContent 将 MessageContent 转换为 Memos 内容和资源文件路径
	//
	// 参数:
	//   msgContent - 解析后的消息内容
	//
	// 返回:
	//   string - 转换后的 Memos 内容（Markdown 格式）
	//   []string - 资源文件路径列表
	//   error - 转换失败时返回错误
	ConvertMessageContent(msgContent *message.MessageContent) (string, []string, error)

	// ConvertRichTextToMarkdown 将富文本转换为 Markdown
	//
	// 参数:
	//   richText - 富文本内容
	//
	// 返回:
	//   string - 转换后的 Markdown 内容
	ConvertRichTextToMarkdown(richText *message.RichTextContent) string

	// ExtractResourcePaths 从 MessageContent 中提取资源文件路径
	//
	// 参数:
	//   msgContent - 消息内容
	//
	// 返回:
	//   []string - 资源文件路径列表
	ExtractResourcePaths(msgContent *message.MessageContent) []string
}

// memosConverter Memos 转换器实现
type memosConverter struct{}

// NewMemosConverter 创建 Memos 转换器
//
// 返回:
//
//	MemosConverter - 初始化好的 Memos 转换器实例
func NewMemosConverter() MemosConverter {
	return &memosConverter{}
}

// ConvertMessageContent 将 MessageContent 转换为 Memos 内容和资源文件路径
func (c *memosConverter) ConvertMessageContent(msgContent *message.MessageContent) (string, []string, error) {
	if msgContent == nil {
		return "", nil, fmt.Errorf("message content is nil")
	}

	// 构建 Memos 内容
	var contentBuilder strings.Builder

	// 根据消息类型处理
	switch msgContent.Type {
	case message.MessageTypePost:
		// 富文本消息
		if msgContent.RichText != nil {
			contentBuilder.WriteString(c.ConvertRichTextToMarkdown(msgContent.RichText))
		} else {
			contentBuilder.WriteString(msgContent.Text)
		}
	default:
		// 其他消息类型，直接使用文本内容
		contentBuilder.WriteString(msgContent.Text)
	}

	// 提取资源文件路径
	filePaths := c.ExtractResourcePaths(msgContent)

	return contentBuilder.String(), filePaths, nil
}

// ConvertRichTextToMarkdown 将富文本转换为 Markdown
func (c *memosConverter) ConvertRichTextToMarkdown(richText *message.RichTextContent) string {
	if richText == nil {
		return ""
	}

	var builder strings.Builder

	// 处理标题
	if richText.Title != "" {
		builder.WriteString(fmt.Sprintf("# %s\n\n", richText.Title))
	}

	// 处理内容行
	for _, line := range richText.Content {
		for _, elem := range line {
			c.convertElementToMarkdown(&builder, elem)
		}
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}

// convertElementToMarkdown 将单个富文本元素转换为 Markdown
func (c *memosConverter) convertElementToMarkdown(builder *strings.Builder, elem *message.RichTextElement) {
	if elem == nil {
		return
	}

	// 处理样式
	stylePrefix := ""
	styleSuffix := ""

	for _, style := range elem.Style {
		switch style {
		case string(message.TextStyleBold):
			stylePrefix += "**"
			styleSuffix = "**" + styleSuffix
		case string(message.TextStyleItalic):
			stylePrefix += "*"
			styleSuffix = "*" + styleSuffix
		case string(message.TextStyleUnderline):
			// Markdown 不支持下划线，使用 HTML
			stylePrefix += "<u>"
			styleSuffix = "</u>" + styleSuffix
		case string(message.TextStyleLineThrough):
			stylePrefix += "~~"
			styleSuffix = "~~" + styleSuffix
		}
	}

	// 根据元素类型转换
	switch elem.Tag {
	case string(message.RichTextTagText):
		builder.WriteString(stylePrefix)
		builder.WriteString(elem.Text)
		builder.WriteString(styleSuffix)

	case string(message.RichTextTagA):
		builder.WriteString(fmt.Sprintf("%s[%s](%s)%s", stylePrefix, elem.Text, elem.Href, styleSuffix))

	case string(message.RichTextTagAt):
		if elem.UserId == "all" {
			builder.WriteString("@所有人")
		} else if elem.UserName != "" {
			builder.WriteString(fmt.Sprintf("@%s", elem.UserName))
		} else {
			builder.WriteString(fmt.Sprintf("@用户(%s)", elem.UserId))
		}

	case string(message.RichTextTagImg):
		// 图片会上传，这里无需渲染
		// builder.WriteString("![图片](图片)")

	case string(message.RichTextTagMedia):
		// 媒体文件会上传，这里无需渲染
		// builder.WriteString(fmt.Sprintf("[文件: %s](文件)", elem.FileName))

	case string(message.RichTextTagEmotion):
		builder.WriteString(fmt.Sprintf(":%s:", elem.EmojiType))

	case string(message.RichTextTagHr):
		builder.WriteString("\n---\n")

	case string(message.RichTextTagCodeBlock):
		builder.WriteString(fmt.Sprintf("```%s\n%s\n```", elem.Language, elem.Text))

	case string(message.RichTextTagMd):
		builder.WriteString(elem.Text)

	default:
		// 未知元素类型，直接输出文本
		builder.WriteString(elem.Text)
	}
}

// ExtractResourcePaths 从 MessageContent 中提取资源文件路径
func (c *memosConverter) ExtractResourcePaths(msgContent *message.MessageContent) []string {
	if msgContent == nil {
		return nil
	}

	var filePaths []string

	// 从 Resources 中提取已下载的资源
	for _, resource := range msgContent.Resources {
		if resource.Downloaded && resource.LocalPath != "" {
			filePaths = append(filePaths, resource.LocalPath)
		}
	}

	return filePaths
}

// ConvertMessageContentToMemoRequest 便捷函数：直接转换为 Memos 创建请求
//
// 参数:
//
//	msgContent - 消息内容
//	visibility - Memo 可见性
//
// 返回:
//
//	string - Memo 内容
//	memos.Visibility - 可见性
//	[]string - 资源文件路径
//	error - 转换失败时返回错误
func ConvertMessageContentToMemoRequest(msgContent *message.MessageContent, visibility memos.Visibility) (string, memos.Visibility, []string, error) {
	converter := NewMemosConverter()
	content, filePaths, err := converter.ConvertMessageContent(msgContent)
	if err != nil {
		return "", "", nil, err
	}

	// 如果没有指定可见性，使用默认值
	if visibility == "" {
		visibility = memos.VisibilityPrivate
	}

	return content, visibility, filePaths, nil
}
