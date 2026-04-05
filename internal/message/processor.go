package message

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/dfface/feishu-bot/internal/logger"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// Processor 消息处理器
// 负责解析飞书 WebSocket 接收到的各种类型消息
// 支持文本、富文本、图片、文件、音频、视频、表情包、位置、群组名片、人名片、系统消息等多种类型
type Processor struct {
	client  *lark.Client // 飞书 API 客户端
	tempDir string       // 临时文件存储目录，用于保存下载的资源
}

// NewProcessor 创建消息处理器
//
// 参数:
//
//	client - 飞书 API 客户端
//
// 返回:
//
//	*Processor - 初始化好的消息处理器实例
func NewProcessor(client *lark.Client) *Processor {
	tempDir := filepath.Join(os.TempDir(), "feishu-bot-messages")
	_ = os.MkdirAll(tempDir, 0755)

	return &Processor{
		client:  client,
		tempDir: tempDir,
	}
}

// Process 处理消息
// 将飞书原始消息事件解析为结构化的 MessageContent
//
// 参数:
//
//	ctx - 上下文
//	msg - 飞书原始消息事件
//
// 返回:
//
//	*MessageContent - 解析后的结构化消息内容
//	error - 解析失败时返回错误
func (p *Processor) Process(ctx context.Context, event *larkim.P2MessageReceiveV1) (*MessageContent, error) {
	if event == nil || event.Event == nil || event.Event.Message == nil || event.Event.Message.MessageType == nil || event.Event.Sender == nil || event.Event.Sender.SenderId == nil || event.Event.Sender.SenderId.OpenId == nil {
		logger.Warn("Invalid message", zap.Any("event", event))
		return nil, fmt.Errorf("invalid message")
	}

	msg := event.Event.Message

	content := &MessageContent{
		ID:         *msg.MessageId,
		SenderID:   *event.Event.Sender.SenderId.OpenId,
		Type:       MessageType(*msg.MessageType),
		RawContent: msg.Content,
		Resources:  make([]Resource, 0),
	}

	switch content.Type {
	case MessageTypeText:
		text, err := p.parseText(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = text

	case MessageTypePost:
		post, resources, richText, err := p.parsePost(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = post
		content.RichText = richText
		// 下载富文本中的所有资源
		for i := range resources {
			resources[i].MessageID = *msg.MessageId
			if err := p.downloadResource(ctx, &resources[i]); err != nil {
				logger.Warn("Failed to download resource from rich text", zap.Error(err))
			} else {
				content.Resources = append(content.Resources, resources[i])
			}
		}

	case MessageTypeImage:
		resource, err := p.parseImage(msg.Content)
		if err != nil {
			return nil, err
		}
		resource.MessageID = *msg.MessageId
		// 下载图片
		if err := p.downloadResource(ctx, &resource); err != nil {
			logger.Warn("Failed to download image", zap.Error(err))
		}
		content.Resources = append(content.Resources, resource)

	case MessageTypeFile:
		resource, err := p.parseFile(msg.Content)
		if err != nil {
			return nil, err
		}
		resource.MessageID = *msg.MessageId
		// 下载文件
		if err := p.downloadResource(ctx, &resource); err != nil {
			logger.Warn("Failed to download file", zap.Error(err))
		}
		content.Resources = append(content.Resources, resource)

	case MessageTypeAudio:
		resource, err := p.parseAudio(msg.Content)
		if err != nil {
			return nil, err
		}
		resource.MessageID = *msg.MessageId
		// 下载音频
		if err := p.downloadResource(ctx, &resource); err != nil {
			logger.Warn("Failed to download audio", zap.Error(err))
		}
		content.Resources = append(content.Resources, resource)

	case MessageTypeMedia:
		resource, err := p.parseMedia(msg.Content)
		if err != nil {
			return nil, err
		}
		resource.MessageID = *msg.MessageId
		// 下载视频
		if err := p.downloadResource(ctx, &resource); err != nil {
			logger.Warn("Failed to download media", zap.Error(err))
		}
		content.Resources = append(content.Resources, resource)

	case MessageTypeLocation:
		location, err := p.parseLocation(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Location = location
		content.Text = fmt.Sprintf("📍 %s\n经度: %s\n纬度: %s", location.Name, location.Longitude, location.Latitude)

	case MessageTypeMergeForward:
		text, err := p.parseMergeForward(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = text

	case MessageTypeSticker:
		sticker, err := p.parseSticker(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Sticker = sticker
		content.Text = "🎨 表情包"

	case MessageTypeInteractive:
		interactive, err := p.parseInteractive(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Interactive = interactive
		content.Text = "📦 交互式消息"

	case MessageTypeRedPacket:
		text, err := p.parseRedPacket(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = text

	case MessageTypeShareChat:
		shareChat, err := p.parseShareChat(msg.Content)
		if err != nil {
			return nil, err
		}
		content.ShareChat = shareChat
		content.Text = "👥 群组名片"

	case MessageTypeShareUser:
		shareUser, err := p.parseShareUser(msg.Content)
		if err != nil {
			return nil, err
		}
		content.ShareUser = shareUser
		content.Text = "👤 人名片"

	case MessageTypeSystem:
		systemMsg, err := p.parseSystemMessage(msg.Content)
		if err != nil {
			return nil, err
		}
		content.SystemMessage = systemMsg
		content.Text = "📢 系统消息"

	case MessageTypeTodo:
		todo, err := p.parseTodo(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Todo = todo
		content.Text = "✅ 待办任务"

	case MessageTypeVote:
		vote, err := p.parseVote(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Vote = vote
		content.Text = "🗳️ 投票消息"

	case MessageTypeFolder:
		text, err := p.parseFolder(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = text

	case MessageTypeCalendar:
		text, err := p.parseCalendar(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = text

	case MessageTypeVideoChat:
		text, err := p.parseVideoChat(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = text
	}

	return content, nil
}

// parseText 解析文本消息
//
// 参数:
//
//	content - 原始文本消息内容 JSON 字符串
//
// 返回:
//
//	string - 解析后的文本内容
//	error - 解析失败时返回错误
func (p *Processor) parseText(content *string) (string, error) {
	if content == nil {
		return "", nil
	}

	var textContent struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(*content), &textContent); err != nil {
		return "", err
	}

	return textContent.Text, nil
}

// parsePost 解析富文本消息，返回文本、资源列表和解析后的富文本结构
//
// 参数:
//
//	content - 原始富文本消息内容 JSON 字符串
//
// 返回:
//
//	string - 纯文本摘要
//	[]Resource - 收集到的资源列表（图片、媒体等）
//	*RichTextContent - 解析后的结构化富文本内容
//	error - 解析失败时返回错误
func (p *Processor) parsePost(content *string) (string, []Resource, *RichTextContent, error) {
	if content == nil {
		return "", nil, nil, nil
	}

	var post struct {
		Title   string `json:"title"`
		Content [][]struct {
			Tag       string   `json:"tag"`
			Text      string   `json:"text,omitempty"`
			Href      string   `json:"href,omitempty"`
			Style     []string `json:"style,omitempty"`
			Language  string   `json:"language,omitempty"`
			EmojiType string   `json:"emoji_type,omitempty"`
			ImageKey  string   `json:"image_key,omitempty"`
			FileKey   string   `json:"file_key,omitempty"`
			UserId    string   `json:"user_id,omitempty"`
			UserName  string   `json:"user_name,omitempty"`
			Width     int      `json:"width,omitempty"`
			Height    int      `json:"height,omitempty"`
		}
	}

	if err := json.Unmarshal([]byte(*content), &post); err != nil {
		return "", nil, nil, err
	}

	var result strings.Builder
	resources := make([]Resource, 0)
	seenFileKeys := make(map[string]bool)

	// 构建富文本结构
	richText := &RichTextContent{
		Title:   post.Title,
		Content: make([][]*RichTextElement, len(post.Content)),
	}

	// 添加标题
	if post.Title != "" {
		result.WriteString(fmt.Sprintf("# %s\n\n", post.Title))
	}

	// 解析内容
	for rowIndex, row := range post.Content {
		richText.Content[rowIndex] = make([]*RichTextElement, len(row))
		for colIndex, segment := range row {
			// 构建富文本元素
			element := &RichTextElement{
				Tag:       segment.Tag,
				Text:      segment.Text,
				Href:      segment.Href,
				Style:     segment.Style,
				ImageKey:  segment.ImageKey,
				FileKey:   segment.FileKey,
				UserId:    segment.UserId,
				UserName:  segment.UserName,
				EmojiType: segment.EmojiType,
				Width:     segment.Width,
				Height:    segment.Height,
				Language:  segment.Language,
			}

			richText.Content[rowIndex][colIndex] = element

			switch segment.Tag {
			case string(RichTextTagText):
				result.WriteString(segment.Text)
			case string(RichTextTagA):
				result.WriteString(fmt.Sprintf("[%s](%s)", segment.Text, segment.Href))
			case string(RichTextTagAt):
				if segment.UserName != "" {
					result.WriteString(fmt.Sprintf("@%s", segment.UserName))
				} else {
					result.WriteString(fmt.Sprintf("@%s", segment.UserId))
				}
			case string(RichTextTagImg):
				result.WriteString(fmt.Sprintf("[图片:%s]", segment.ImageKey))
				// 收集图片资源
				if segment.ImageKey != "" && !seenFileKeys[segment.ImageKey] {
					resources = append(resources, Resource{
						Type:     ResourceTypeImage,
						FileKey:  segment.ImageKey,
						ImageKey: segment.ImageKey,
					})
					seenFileKeys[segment.ImageKey] = true
				}
			case string(RichTextTagMedia):
				result.WriteString(fmt.Sprintf("[媒体:%s]", segment.FileKey))
				// 收集媒体资源
				if segment.FileKey != "" && !seenFileKeys[segment.FileKey] {
					resources = append(resources, Resource{
						Type:     ResourceTypeMedia,
						FileKey:  segment.FileKey,
						ImageKey: segment.ImageKey,
					})
					seenFileKeys[segment.FileKey] = true
				}
			case string(RichTextTagEmotion):
				result.WriteString(fmt.Sprintf("[表情:%s]", segment.EmojiType))
			case string(RichTextTagHr):
				result.WriteString("\n---\n")
			case string(RichTextTagCodeBlock):
				result.WriteString(fmt.Sprintf("\n```%s\n%s\n```\n", segment.Language, segment.Text))
			case string(RichTextTagMd):
				result.WriteString(segment.Text)
			}
		}
		result.WriteString("\n")
	}

	return result.String(), resources, richText, nil
}

// parseImage 解析图片消息
//
// 参数:
//
//	content - 原始图片消息内容 JSON 字符串
//
// 返回:
//
//	Resource - 解析后的图片资源信息
//	error - 解析失败时返回错误
func (p *Processor) parseImage(content *string) (Resource, error) {
	if content == nil {
		return Resource{}, fmt.Errorf("empty content")
	}

	var imageContent struct {
		ImageKey string `json:"image_key"`
	}
	if err := json.Unmarshal([]byte(*content), &imageContent); err != nil {
		return Resource{}, err
	}

	return Resource{
		Type:     ResourceTypeImage,
		FileKey:  imageContent.ImageKey,
		ImageKey: imageContent.ImageKey,
	}, nil
}

// parseFile 解析文件消息
//
// 参数:
//
//	content - 原始文件消息内容 JSON 字符串
//
// 返回:
//
//	Resource - 解析后的文件资源信息
//	error - 解析失败时返回错误
func (p *Processor) parseFile(content *string) (Resource, error) {
	if content == nil {
		return Resource{}, fmt.Errorf("empty content")
	}

	var fileContent struct {
		FileKey  string `json:"file_key"`
		FileName string `json:"file_name"`
	}
	if err := json.Unmarshal([]byte(*content), &fileContent); err != nil {
		return Resource{}, err
	}

	return Resource{
		Type:     ResourceTypeFile,
		FileKey:  fileContent.FileKey,
		FileName: fileContent.FileName,
	}, nil
}

// parseAudio 解析音频消息
//
// 参数:
//
//	content - 原始音频消息内容 JSON 字符串
//
// 返回:
//
//	Resource - 解析后的音频资源信息
//	error - 解析失败时返回错误
func (p *Processor) parseAudio(content *string) (Resource, error) {
	if content == nil {
		return Resource{}, fmt.Errorf("empty content")
	}

	var audioContent struct {
		FileKey  string `json:"file_key"`
		Duration int    `json:"duration"`
	}
	if err := json.Unmarshal([]byte(*content), &audioContent); err != nil {
		return Resource{}, err
	}

	return Resource{
		Type:     ResourceTypeAudio,
		FileKey:  audioContent.FileKey,
		Duration: audioContent.Duration,
	}, nil
}

// parseMedia 解析视频消息
//
// 参数:
//
//	content - 原始视频消息内容 JSON 字符串
//
// 返回:
//
//	Resource - 解析后的视频资源信息
//	error - 解析失败时返回错误
func (p *Processor) parseMedia(content *string) (Resource, error) {
	if content == nil {
		return Resource{}, fmt.Errorf("empty content")
	}

	var mediaContent struct {
		FileKey  string `json:"file_key"`
		ImageKey string `json:"image_key"`
		FileName string `json:"file_name"`
		Duration int    `json:"duration"`
	}
	if err := json.Unmarshal([]byte(*content), &mediaContent); err != nil {
		return Resource{}, err
	}

	return Resource{
		Type:     ResourceTypeMedia,
		FileKey:  mediaContent.FileKey,
		ImageKey: mediaContent.ImageKey,
		FileName: mediaContent.FileName,
		Duration: mediaContent.Duration,
	}, nil
}

// parseLocation 解析位置消息
//
// 参数:
//
//	content - 原始位置消息内容 JSON 字符串
//
// 返回:
//
//	*Location - 解析后的位置信息
//	error - 解析失败时返回错误
func (p *Processor) parseLocation(content *string) (*Location, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var locationContent struct {
		Name      string `json:"name"`
		Longitude string `json:"longitude"`
		Latitude  string `json:"latitude"`
	}
	if err := json.Unmarshal([]byte(*content), &locationContent); err != nil {
		return nil, err
	}

	return &Location{
		Name:      locationContent.Name,
		Longitude: locationContent.Longitude,
		Latitude:  locationContent.Latitude,
	}, nil
}

// parseMergeForward 解析合并转发消息
//
// 参数:
//
//	content - 原始合并转发消息内容 JSON 字符串
//
// 返回:
//
//	string - 解析后的文本摘要
//	error - 解析失败时返回错误
func (p *Processor) parseMergeForward(content *string) (string, error) {
	if content == nil {
		return "", fmt.Errorf("empty content")
	}

	var mergeContent struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(*content), &mergeContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("📦 %s", mergeContent.Content), nil
}

// downloadResource 下载资源到本地
// 使用飞书 API 的 MessageResource.Get() 接口下载消息中的资源文件
//
// 参数:
//
//	ctx - 上下文
//	resource - 要下载的资源信息（包含 MessageID、FileKey 等）
//
// 返回:
//
//	error - 下载失败时返回错误
func (p *Processor) downloadResource(ctx context.Context, resource *Resource) error {
	if resource.MessageID == "" {
		return fmt.Errorf("no message id")
	}

	// 确定要下载的文件 key 和类型
	var fileKey string
	var resourceType string
	switch resource.Type {
	case ResourceTypeImage:
		if resource.ImageKey != "" {
			fileKey = resource.ImageKey
		} else {
			fileKey = resource.FileKey
		}
		resourceType = string(ResourceTypeImage)
	default:
		fileKey = resource.FileKey
		resourceType = string(ResourceTypeFile)
	}

	if fileKey == "" {
		return fmt.Errorf("no file key")
	}

	// 使用 MessageResource.Get() 下载资源
	req := larkim.NewGetMessageResourceReqBuilder().
		MessageId(resource.MessageID).
		FileKey(fileKey).
		Type(resourceType).
		Build()

	resp, err := p.client.Im.V1.MessageResource.Get(ctx, req)
	if err != nil {
		return err
	}

	if !resp.Success() {
		return fmt.Errorf("failed to get message resource: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	// 构建唯一的本地路径，避免文件名冲突
	fileName := resp.FileName
	if fileName == "" {
		// 如果没有文件名，使用 UUID 生成
		fileName = uuid.New().String()
		// 从 Content-Type 头获取 MIME 类型来确定扩展名
		contentType := resp.Header.Get("Content-Type")
		ext := p.getExtensionFromContentType(contentType, resource.Type)
		fileName += ext
	} else {
		// 如果有文件名，还是添加 UUID 前缀防止冲突
		ext := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, ext)
		fileName = fmt.Sprintf("%s_%s%s", baseName, uuid.New().String(), ext)
	}

	localPath := filepath.Join(p.tempDir, fileName)

	// 使用 SDK 提供的方法保存文件
	if err := resp.WriteFile(localPath); err != nil {
		return err
	}

	// 更新资源信息
	resource.LocalPath = localPath
	resource.Downloaded = true
	// 也更新 FileName 为我们生成的文件名
	resource.FileName = fileName

	logger.Info("Resource downloaded",
		zap.String("type", string(resource.Type)),
		zap.String("message_id", resource.MessageID),
		zap.String("file_key", fileKey),
		zap.String("path", localPath))
	return nil
}

// getExtensionFromContentType 根据 MIME 类型获取文件扩展名
// 使用标准库 mime 来处理，更全面可靠
func (p *Processor) getExtensionFromContentType(contentType string, resourceType ResourceType) string {
	// 首先尝试从标准库获取扩展名
	if contentType != "" {
		// 解析 Content-Type，去除参数部分（如 "image/jpeg; charset=utf-8"）
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err == nil {
			extensions, err := mime.ExtensionsByType(mediaType)
			if err == nil && len(extensions) > 0 {
				return extensions[0]
			}
		}
	}

	// 如果标准库没有找到，根据资源类型返回默认扩展名
	switch resourceType {
	case ResourceTypeImage:
		return ".jpg"
	case ResourceTypeAudio:
		return ".mp3"
	case ResourceTypeMedia:
		return ".mp4"
	case ResourceTypeFile:
		return ".bin"
	default:
		return ".bin"
	}
}

// parseSticker 解析表情包消息
//
// 参数:
//
//	content - 原始表情包消息内容 JSON 字符串
//
// 返回:
//
//	*Sticker - 解析后的表情包信息
//	error - 解析失败时返回错误
func (p *Processor) parseSticker(content *string) (*Sticker, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var stickerContent struct {
		FileKey string `json:"file_key"`
	}
	if err := json.Unmarshal([]byte(*content), &stickerContent); err != nil {
		return nil, err
	}

	return &Sticker{
		FileKey: stickerContent.FileKey,
	}, nil
}

// parseInteractive 解析交互式消息
//
// 参数:
//
//	content - 原始交互式消息内容 JSON 字符串
//
// 返回:
//
//	map[string]interface{} - 解析后的交互式卡片内容
//	error - 解析失败时返回错误
func (p *Processor) parseInteractive(content *string) (map[string]interface{}, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var interactive map[string]interface{}
	if err := json.Unmarshal([]byte(*content), &interactive); err != nil {
		return nil, err
	}

	return interactive, nil
}

// parseRedPacket 解析红包消息
//
// 参数:
//
//	content - 原始红包消息内容 JSON 字符串
//
// 返回:
//
//	string - 解析后的文本摘要
//	error - 解析失败时返回错误
func (p *Processor) parseRedPacket(content *string) (string, error) {
	if content == nil {
		return "", fmt.Errorf("empty content")
	}

	var redPacketContent struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(*content), &redPacketContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("🧧 %s", redPacketContent.Text), nil
}

// parseShareChat 解析群组名片
//
// 参数:
//
//	content - 原始群组名片消息内容 JSON 字符串
//
// 返回:
//
//	*ShareChat - 解析后的群组名片信息
//	error - 解析失败时返回错误
func (p *Processor) parseShareChat(content *string) (*ShareChat, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var shareChatContent struct {
		ChatID string `json:"chat_id"`
	}
	if err := json.Unmarshal([]byte(*content), &shareChatContent); err != nil {
		return nil, err
	}

	return &ShareChat{
		ChatID: shareChatContent.ChatID,
	}, nil
}

// parseShareUser 解析人名片
//
// 参数:
//
//	content - 原始人名片消息内容 JSON 字符串
//
// 返回:
//
//	*ShareUser - 解析后的人名片信息
//	error - 解析失败时返回错误
func (p *Processor) parseShareUser(content *string) (*ShareUser, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var shareUserContent struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal([]byte(*content), &shareUserContent); err != nil {
		return nil, err
	}

	return &ShareUser{
		UserID: shareUserContent.UserID,
	}, nil
}

// parseSystemMessage 解析系统消息
//
// 参数:
//
//	content - 原始系统消息内容 JSON 字符串
//
// 返回:
//
//	*SystemMessage - 解析后的系统消息内容
//	error - 解析失败时返回错误
func (p *Processor) parseSystemMessage(content *string) (*SystemMessage, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var systemContent struct {
		Template string                 `json:"template"`
		Params   map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal([]byte(*content), &systemContent); err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for k, v := range systemContent.Params {
		params[k] = fmt.Sprintf("%v", v)
	}

	return &SystemMessage{
		Template: systemContent.Template,
		Params:   params,
	}, nil
}

// parseTodo 解析待办任务消息
//
// 参数:
//
//	content - 原始待办任务消息内容 JSON 字符串
//
// 返回:
//
//	*Todo - 解析后的待办任务信息
//	error - 解析失败时返回错误
func (p *Processor) parseTodo(content *string) (*Todo, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var todoContent struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal([]byte(*content), &todoContent); err != nil {
		return nil, err
	}

	return &Todo{
		TaskID: todoContent.TaskID,
	}, nil
}

// parseVote 解析投票消息
//
// 参数:
//
//	content - 原始投票消息内容 JSON 字符串
//
// 返回:
//
//	*Vote - 解析后的投票信息
//	error - 解析失败时返回错误
func (p *Processor) parseVote(content *string) (*Vote, error) {
	if content == nil {
		return nil, fmt.Errorf("empty content")
	}

	var voteContent struct {
		VoteID string `json:"vote_id"`
	}
	if err := json.Unmarshal([]byte(*content), &voteContent); err != nil {
		return nil, err
	}

	return &Vote{
		VoteID: voteContent.VoteID,
	}, nil
}

// parseFolder 解析文件夹消息
//
// 参数:
//
//	content - 原始文件夹消息内容 JSON 字符串
//
// 返回:
//
//	string - 解析后的文本摘要
//	error - 解析失败时返回错误
func (p *Processor) parseFolder(content *string) (string, error) {
	if content == nil {
		return "", fmt.Errorf("empty content")
	}

	var folderContent struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(*content), &folderContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("📁 %s", folderContent.Name), nil
}

// parseCalendar 解析日历消息
//
// 参数:
//
//	content - 原始日历消息内容 JSON 字符串
//
// 返回:
//
//	string - 解析后的文本摘要
//	error - 解析失败时返回错误
func (p *Processor) parseCalendar(content *string) (string, error) {
	if content == nil {
		return "", fmt.Errorf("empty content")
	}

	var calendarContent struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(*content), &calendarContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("📅 %s", calendarContent.Title), nil
}

// parseVideoChat 解析视频会议消息
//
// 参数:
//
//	content - 原始视频会议消息内容 JSON 字符串
//
// 返回:
//
//	string - 解析后的文本摘要
//	error - 解析失败时返回错误
func (p *Processor) parseVideoChat(content *string) (string, error) {
	if content == nil {
		return "", fmt.Errorf("empty content")
	}

	var videoChatContent struct {
		Topic string `json:"topic"`
	}
	if err := json.Unmarshal([]byte(*content), &videoChatContent); err != nil {
		return "", err
	}

	return fmt.Sprintf("🎥 %s", videoChatContent.Topic), nil
}

// 确保 Processor 实现了 MessageReceiver 接口
var _ MessageReceiver = (*Processor)(nil)
