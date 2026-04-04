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

// MessageContent 消息内容
type MessageContent struct {
	Type          MessageType
	Text          string
	RawContent    *string // 原始消息内容 JSON 字符串
	Resources     []Resource
	Location      *Location
	Sticker       *Sticker
	Interactive   map[string]interface{}
	ShareChat     *ShareChat
	ShareUser     *ShareUser
	SystemMessage *SystemMessage
	Todo          *Todo
	Vote          *Vote
}

// Resource 资源信息
type Resource struct {
	Type       string
	FileKey    string
	FileName   string
	ImageKey   string
	Duration   int
	LocalPath  string // 本地文件路径
	Downloaded bool   // 是否已下载
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
		post, resources, err := p.parsePost(msg.Content)
		if err != nil {
			return nil, err
		}
		content.Text = post
		// 下载富文本中的所有资源
		for i := range resources {
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

// parsePost 解析富文本消息，返回文本和资源列表
func (p *Processor) parsePost(content *string) (string, []Resource, error) {
	if content == nil {
		return "", nil, nil
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
		}
	}

	if err := json.Unmarshal([]byte(*content), &post); err != nil {
		return "", nil, err
	}

	var result strings.Builder
	resources := make([]Resource, 0)
	seenFileKeys := make(map[string]bool)

	// 添加标题
	if post.Title != "" {
		result.WriteString(fmt.Sprintf("# %s\n\n", post.Title))
	}

	// 解析内容
	for _, row := range post.Content {
		for _, segment := range row {
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

	return result.String(), resources, nil
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
	// 确定文件名
	var fileName string
	if resource.FileName != "" {
		fileName = resource.FileName
	} else {
		switch resource.Type {
		case "image":
			if resource.ImageKey != "" {
				fileName = fmt.Sprintf("%s.jpg", resource.ImageKey)
			} else {
				fileName = fmt.Sprintf("%s.jpg", resource.FileKey)
			}
		case "audio":
			fileName = fmt.Sprintf("%s.aac", resource.FileKey)
		case "media":
			fileName = fmt.Sprintf("%s.mp4", resource.FileKey)
		case "file":
			fileName = resource.FileKey
		default:
			fileName = resource.FileKey
		}
	}

	// 构建本地路径
	localPath := filepath.Join(p.tempDir, fileName)

	// 下载文件
	if resource.Type == "image" {
		var imageKey string
		if resource.ImageKey != "" {
			imageKey = resource.ImageKey
		} else {
			imageKey = resource.FileKey
		}

		if imageKey == "" {
			return fmt.Errorf("no image key")
		}

		req := larkim.NewGetImageReqBuilder().ImageKey(imageKey).Build()
		resp, err := p.client.Im.Image.Get(ctx, req)
		if err != nil {
			return err
		}

		if resp.Code != 0 {
			return fmt.Errorf("failed to get image resource: code=%d, msg=%s", resp.Code, resp.Msg)
		}

		// 使用 SDK 提供的方法保存文件
		if err := resp.WriteFile(localPath); err != nil {
			return err
		}
	} else {
		if resource.FileKey == "" {
			return fmt.Errorf("no file key")
		}

		req := larkim.NewGetFileReqBuilder().FileKey(resource.FileKey).Build()
		resp, err := p.client.Im.File.Get(ctx, req)
		if err != nil {
			return err
		}

		if !resp.Success() {
			return fmt.Errorf("failed to get file resource: %s", resp.Msg)
		}

		// 使用 SDK 提供的方法保存文件
		if err := resp.WriteFile(localPath); err != nil {
			return err
		}
	}

	// 更新资源信息
	resource.LocalPath = localPath
	resource.Downloaded = true

	p.logger.Info("Resource downloaded", zap.String("type", resource.Type), zap.String("path", localPath))
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
