package message

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// ParsedRichTextElement 解析后的富文本元素
// 用于存储从飞书富文本消息中解析出的结构化元素
// 支持所有富文本元素类型：text、a、at、img、media、emotion、hr、code_block、md

type ParsedRichTextElement struct {
	Tag       string   `json:"tag"`                  // 元素类型：text、a、at、img、media、emotion、hr、code_block、md
	Text      string   `json:"text,omitempty"`       // 文本内容，用于 text、a、code_block、md 元素
	Href      string   `json:"href,omitempty"`       // 链接地址，用于 a 元素
	UserId    string   `json:"user_id,omitempty"`    // 用户 ID，用于 at 元素
	UserName  string   `json:"user_name,omitempty"`  // 用户名称，用于 at 元素
	ImageKey  string   `json:"image_key,omitempty"`  // 图片 key，用于 img 元素
	Width     int      `json:"width,omitempty"`      // 图片宽度，用于 img 元素
	Height    int      `json:"height,omitempty"`     // 图片高度，用于 img 元素
	FileType  string   `json:"file_type,omitempty"`  // 文件类型，用于 media 元素
	FileKey   string   `json:"file_key,omitempty"`   // 文件 key，用于 media 元素
	FileName  string   `json:"file_name,omitempty"`  // 文件名称，用于 media 元素
	Duration  int      `json:"duration,omitempty"`   // 媒体时长，用于 media 元素
	EmojiType string   `json:"emoji_type,omitempty"` // 表情类型，用于 emotion 元素
	Style     []string `json:"style,omitempty"`      // 样式列表，支持：bold、italic、underline、lineThrough
	Content   string   `json:"content,omitempty"`    // 内容，用于 code_block、md 等元素
	Language  string   `json:"language,omitempty"`   // 代码语言，用于 code_block 元素
}

// ParsedRichTextContent 解析后的富文本内容
// 存储完整的富文本消息结构，包括标题和所有元素

type ParsedRichTextContent struct {
	Title   string                    `json:"title"`   // 富文本标题
	Content [][]ParsedRichTextElement `json:"content"` // 内容行列表，每行包含多个元素
}

// MessageContent 消息内容
// 存储解析后的结构化消息内容，包含各种类型消息的详细信息

type MessageContent struct {
	Type          MessageType            // 消息类型：text、post、image、file、audio、media、sticker、interactive、share_chat、share_user、system、todo、vote
	Text          string                 // 消息文本内容，对于文本消息是完整内容，对于其他消息是摘要
	RawContent    *string                // 原始消息内容 JSON 字符串
	RichText      *ParsedRichTextContent // 解析后的富文本结构，仅当消息类型为 post 时有效
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
	Type       string // 资源类型：image、file、audio、media
	FileKey    string // 文件 key，用于下载和重新上传
	FileName   string // 文件名称
	ImageKey   string // 图片 key，用于图片资源
	MessageID  string // 所属消息的 ID，用于下载资源
	Duration   int    // 媒体时长（秒），用于音频和视频资源
	LocalPath  string // 本地文件路径，下载后存储的位置
	Downloaded bool   // 是否已下载到本地
}

// Location 位置信息
type Location struct {
	Name      string
	Longitude string
	Latitude  string
}

// Sticker 表情包信息
type Sticker struct {
	FileKey string
}

// ShareChat 群组名片信息
type ShareChat struct {
	ChatID string
}

// ShareUser 人名片信息
type ShareUser struct {
	UserID string
}

// SystemMessage 系统消息信息
type SystemMessage struct {
	Template string
	Params   map[string]string
}

// Todo 待办任务信息
type Todo struct {
	TaskID string
}

// Vote 投票信息
type Vote struct {
	VoteID string
}

// Processor 消息处理器
type Processor struct {
	client  *lark.Client
	logger  *zap.Logger
	tempDir string
}

// NewProcessor 创建消息处理器
func NewProcessor(client *lark.Client, logger *zap.Logger) *Processor {
	tempDir := filepath.Join(os.TempDir(), "feishu-bot-messages")
	_ = os.MkdirAll(tempDir, 0755)

	return &Processor{
		client:  client,
		logger:  logger,
		tempDir: tempDir,
	}
}

// Process 处理消息
func (p *Processor) Process(ctx context.Context, msg *larkim.EventMessage) (*MessageContent, error) {
	if msg == nil || msg.MessageType == nil {
		return nil, fmt.Errorf("invalid message")
	}

	content := &MessageContent{
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
				p.logger.Warn("Failed to download resource from rich text", zap.Error(err))
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
			p.logger.Warn("Failed to download image", zap.Error(err))
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
			p.logger.Warn("Failed to download file", zap.Error(err))
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
			p.logger.Warn("Failed to download audio", zap.Error(err))
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
			p.logger.Warn("Failed to download media", zap.Error(err))
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
func (p *Processor) parsePost(content *string) (string, []Resource, *ParsedRichTextContent, error) {
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
	richText := &ParsedRichTextContent{
		Title:   post.Title,
		Content: make([][]ParsedRichTextElement, len(post.Content)),
	}

	// 添加标题
	if post.Title != "" {
		result.WriteString(fmt.Sprintf("# %s\n\n", post.Title))
	}

	// 解析内容
	for rowIndex, row := range post.Content {
		richText.Content[rowIndex] = make([]ParsedRichTextElement, len(row))
		for colIndex, segment := range row {
			// 构建富文本元素
			element := ParsedRichTextElement{
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
			case "text":
				result.WriteString(segment.Text)
			case "a":
				result.WriteString(fmt.Sprintf("[%s](%s)", segment.Text, segment.Href))
			case "at":
				if segment.UserName != "" {
					result.WriteString(fmt.Sprintf("@%s", segment.UserName))
				} else {
					result.WriteString(fmt.Sprintf("@%s", segment.UserId))
				}
			case "img":
				result.WriteString(fmt.Sprintf("[图片:%s]", segment.ImageKey))
				// 收集图片资源
				if segment.ImageKey != "" && !seenFileKeys[segment.ImageKey] {
					resources = append(resources, Resource{
						Type:     "image",
						FileKey:  segment.ImageKey,
						ImageKey: segment.ImageKey,
					})
					seenFileKeys[segment.ImageKey] = true
				}
			case "media":
				result.WriteString(fmt.Sprintf("[媒体:%s]", segment.FileKey))
				// 收集媒体资源
				if segment.FileKey != "" && !seenFileKeys[segment.FileKey] {
					resources = append(resources, Resource{
						Type:     "media",
						FileKey:  segment.FileKey,
						ImageKey: segment.ImageKey,
					})
					seenFileKeys[segment.FileKey] = true
				}
			case "emotion":
				result.WriteString(fmt.Sprintf("[表情:%s]", segment.EmojiType))
			case "hr":
				result.WriteString("\n---\n")
			case "code_block":
				result.WriteString(fmt.Sprintf("\n```%s\n%s\n```\n", segment.Language, segment.Text))
			case "md":
				result.WriteString(segment.Text)
			}
		}
		result.WriteString("\n")
	}

	return result.String(), resources, richText, nil
}

// parseImage 解析图片消息
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
		Type:     "image",
		FileKey:  imageContent.ImageKey,
		ImageKey: imageContent.ImageKey,
	}, nil
}

// parseFile 解析文件消息
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
		Type:     "file",
		FileKey:  fileContent.FileKey,
		FileName: fileContent.FileName,
	}, nil
}

// parseAudio 解析音频消息
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
		Type:     "audio",
		FileKey:  audioContent.FileKey,
		Duration: audioContent.Duration,
	}, nil
}

// parseMedia 解析视频消息
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
		Type:     "media",
		FileKey:  mediaContent.FileKey,
		ImageKey: mediaContent.ImageKey,
		FileName: mediaContent.FileName,
		Duration: mediaContent.Duration,
	}, nil
}

// parseLocation 解析位置消息
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
func (p *Processor) downloadResource(ctx context.Context, resource *Resource) error {
	if resource.MessageID == "" {
		return fmt.Errorf("no message id")
	}

	// 确定要下载的文件 key 和类型
	var fileKey string
	var resourceType string
	switch resource.Type {
	case "image":
		if resource.ImageKey != "" {
			fileKey = resource.ImageKey
		} else {
			fileKey = resource.FileKey
		}
		resourceType = "image"
	default:
		fileKey = resource.FileKey
		resourceType = "file"
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

	// 构建本地路径
	localPath := filepath.Join(p.tempDir, resp.FileName)

	if !resp.Success() {
		return fmt.Errorf("failed to get message resource: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	// 使用 SDK 提供的方法保存文件
	if err := resp.WriteFile(localPath); err != nil {
		return err
	}

	// 更新资源信息
	resource.LocalPath = localPath
	resource.Downloaded = true

	p.logger.Info("Resource downloaded",
		zap.String("type", resource.Type),
		zap.String("message_id", resource.MessageID),
		zap.String("file_key", fileKey),
		zap.String("path", localPath))
	return nil
}

// parseSticker 解析表情包消息
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
