package message

// === 便捷函数 ===

// CreateTextMessage 创建文本消息
func CreateTextMessage(text string) MessageBuilder {
	return NewTextMessageBuilder(text)
}

// CreateRichTextMessage 创建富文本消息（便捷方法）
func CreateRichTextMessage(title string, lines ...[]RichTextElement) MessageBuilder {
	builder := NewRichTextMessageBuilder()
	if title != "" {
		builder.SetTitle(title)
	}
	for _, line := range lines {
		builder.AddLine(line...)
	}
	return builder
}

// CreateImageMessage 创建图片消息
func CreateImageMessage(imageKey string) MessageBuilder {
	return NewImageMessageBuilder(imageKey)
}

// CreateFileMessage 创建文件消息
func CreateFileMessage(fileKey, fileName string) MessageBuilder {
	return NewFileMessageBuilder(fileKey, fileName)
}

// CreateSimpleRichTextMessage 创建简单富文本消息（包含一行文本）
func CreateSimpleRichTextMessage(text string) MessageBuilder {
	return NewRichTextMessageBuilder().AddText(text)
}

// CreateRichTextWithImage 创建包含图片的富文本消息
func CreateRichTextWithImage(text, imageKey string) MessageBuilder {
	return NewRichTextMessageBuilder().
		AddText(text).
		NewLine().
		AddImage(imageKey)
}

// CreateTextWithLink 创建带链接的富文本消息
func CreateTextWithLink(linkText, linkHref string) MessageBuilder {
	return NewRichTextMessageBuilder().
		AddLink(linkText, linkHref)
}

// CreateTextWithAt 创建带@用户的富文本消息
func CreateTextWithAt(text string, userID, userName string) MessageBuilder {
	return NewRichTextMessageBuilder().
		AddText(text).
		AddAt(userID, userName)
}

// CreateStyledRichText 创建带样式的富文本消息示例
func CreateStyledRichText(title string) MessageBuilder {
	return NewRichTextMessageBuilder().
		SetTitle(title).
		AddBoldText("这是加粗文本").
		AddItalicText(" 这是斜体文本").
		NewLine().
		AddUnderlineText("这是下划线文本").
		AddStrikethroughText(" 这是删除线文本")
}
