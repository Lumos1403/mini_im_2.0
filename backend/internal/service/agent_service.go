package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"mini_im/backend/internal/model"
	"mini_im/backend/internal/pkg/agentclient"
	"mini_im/backend/internal/pkg/logger"
	"mini_im/backend/internal/pkg/password"
	"mini_im/backend/internal/pkg/snowflake"
	mysqlrepo "mini_im/backend/internal/repository/mysql"

	"go.uber.org/zap"
)

const defaultAgentFailureMessage = "抱歉，我暂时无法回复，请稍后再试。"

var ErrDefaultAgentUsernameTaken = errors.New("default agent username is used by non-agent user")

var markdownImagePattern = regexp.MustCompile(`!\[([^\]\n]*)\]\((https?://[^\s)]+)\)`)
var mermaidFencePattern = regexp.MustCompile("(?s)```mermaid\\s*\\r?\\n(.*?)\\r?\\n```")

type AgentOptions struct {
	Enabled          bool
	APIBaseURL       string
	APITimeout       time.Duration
	DefaultUsername  string
	DefaultNickname  string
	DefaultAvatarURL string
	FailureMessage   string
}

type AgentMessageNotifier interface {
	NotifyAgentMessage(ctx context.Context, userID int64, data MessageReceiveOutput) error
	NotifyAgentMessageStreamStart(ctx context.Context, userID int64, data AgentMessageStartOutput) error
	NotifyAgentMessageStreamChunk(ctx context.Context, userID int64, data AgentMessageChunkOutput) error
	NotifyAgentMessageStreamDone(ctx context.Context, userID int64, data AgentMessageDoneOutput) error
	NotifyAgentMessageStreamError(ctx context.Context, userID int64, data AgentMessageErrorOutput) error
}

type AgentReplyTask struct {
	ConversationID int64
	UserID         int64
	AgentUserID    int64
	Message        string
}

type agentMarkdownImage struct {
	Alt string `json:"alt"`
	URL string `json:"url"`
}

type agentTextExtra struct {
	MarkdownImages []agentMarkdownImage `json:"markdown_images,omitempty"`
}

type AgentService struct {
	userRepo         *mysqlrepo.UserRepository
	friendRepo       *mysqlrepo.FriendRepository
	conversationRepo *mysqlrepo.ConversationRepository
	messageRepo      *mysqlrepo.MessageRepository
	idGenerator      *snowflake.Node
	client           *agentclient.Client
	notifier         AgentMessageNotifier
	options          AgentOptions
	mu               sync.RWMutex
	defaultAgentID   int64
}

func NewAgentService(
	userRepo *mysqlrepo.UserRepository,
	friendRepo *mysqlrepo.FriendRepository,
	conversationRepo *mysqlrepo.ConversationRepository,
	messageRepo *mysqlrepo.MessageRepository,
	idGenerator *snowflake.Node,
	options AgentOptions,
	client *agentclient.Client,
) *AgentService {
	options.DefaultUsername = strings.TrimSpace(options.DefaultUsername)
	if options.DefaultUsername == "" {
		options.DefaultUsername = "default_agent"
	}
	options.DefaultNickname = strings.TrimSpace(options.DefaultNickname)
	if options.DefaultNickname == "" {
		options.DefaultNickname = "IM Agent"
	}
	if options.APITimeout <= 0 {
		options.APITimeout = 30 * time.Second
	}
	options.FailureMessage = strings.TrimSpace(options.FailureMessage)
	if options.FailureMessage == "" {
		options.FailureMessage = defaultAgentFailureMessage
	}

	return &AgentService{
		userRepo:         userRepo,
		friendRepo:       friendRepo,
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		idGenerator:      idGenerator,
		client:           client,
		options:          options,
	}
}

func (s *AgentService) SetMessageNotifier(notifier AgentMessageNotifier) {
	s.notifier = notifier
}

func (s *AgentService) EnsureDefaultAgentUser(ctx context.Context) (*model.UserWithProfile, error) {
	if s == nil || s.userRepo == nil || s.idGenerator == nil {
		return nil, errors.New("agent service unavailable")
	}

	username := s.options.DefaultUsername
	existing, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil {
		if existing.User.UserType != model.UserTypeAgent {
			return nil, fmt.Errorf("%w: %s", ErrDefaultAgentUsernameTaken, username)
		}
		if existing.User.Status != model.UserStatusNormal {
			return nil, errors.New("default agent user is not active")
		}
		s.setDefaultAgentID(existing.User.UserID)
		return existing, nil
	}
	if !errors.Is(err, mysqlrepo.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := randomAgentPasswordHash()
	if err != nil {
		return nil, err
	}

	agentUserID := s.idGenerator.NextID()
	user := &model.User{
		UserID:       agentUserID,
		Username:     username,
		PasswordHash: passwordHash,
		UserType:     model.UserTypeAgent,
		Status:       model.UserStatusNormal,
	}
	profile := &model.UserProfile{
		UserID:        agentUserID,
		Nickname:      s.options.DefaultNickname,
		AvatarURL:     sql.NullString{String: strings.TrimSpace(s.options.DefaultAvatarURL), Valid: strings.TrimSpace(s.options.DefaultAvatarURL) != ""},
		ProfileStatus: model.UserStatusNormal,
	}

	if err := s.userRepo.CreateUserWithProfile(ctx, user, profile); err != nil {
		if errors.Is(err, mysqlrepo.ErrDuplicateUser) {
			existing, findErr := s.userRepo.FindByUsername(ctx, username)
			if findErr != nil {
				return nil, findErr
			}
			if existing.User.UserType != model.UserTypeAgent {
				return nil, fmt.Errorf("%w: %s", ErrDefaultAgentUsernameTaken, username)
			}
			s.setDefaultAgentID(existing.User.UserID)
			return existing, nil
		}
		return nil, err
	}

	created := &model.UserWithProfile{User: *user, Profile: *profile}
	s.setDefaultAgentID(agentUserID)
	return created, nil
}

func (s *AgentService) EnsureDefaultAgentFriend(ctx context.Context, userID int64) error {
	if s == nil || userID <= 0 {
		return errors.New("invalid default agent friend request")
	}

	agent, err := s.EnsureDefaultAgentUser(ctx)
	if err != nil {
		return err
	}
	agentUserID := agent.User.UserID
	if agentUserID == userID {
		return nil
	}

	conversationID := s.idGenerator.NextID()
	return s.userRepo.WithTx(ctx, func(ctx context.Context, exec mysqlrepo.Executor) error {
		return s.EnsureDefaultAgentFriendInTx(ctx, exec, userID, agentUserID, conversationID)
	})
}

func (s *AgentService) EnsureDefaultAgentFriendInTx(ctx context.Context, exec mysqlrepo.Executor, userID int64, agentUserID int64, conversationID int64) error {
	if s == nil || s.friendRepo == nil || s.conversationRepo == nil {
		return errors.New("agent service unavailable")
	}
	if userID <= 0 || agentUserID <= 0 || userID == agentUserID {
		return errors.New("invalid default agent friendship")
	}

	if err := s.friendRepo.EnsureFriendshipInTx(ctx, exec, userID, agentUserID); err != nil {
		return err
	}
	_, err := s.conversationRepo.EnsurePrivateConversationInTx(ctx, exec, conversationID, userID, agentUserID)
	return err
}

func (s *AgentService) IsDefaultAgentUser(ctx context.Context, userID int64) (bool, error) {
	if s == nil || userID <= 0 {
		return false, nil
	}

	agentID := s.cachedDefaultAgentID()
	if agentID <= 0 {
		agent, err := s.EnsureDefaultAgentUser(ctx)
		if err != nil {
			return false, err
		}
		agentID = agent.User.UserID
	}
	return userID == agentID, nil
}

func (s *AgentService) HandleUserMessageAsync(task AgentReplyTask) {
	if s == nil || task.ConversationID <= 0 || task.UserID <= 0 || task.AgentUserID <= 0 || strings.TrimSpace(task.Message) == "" {
		return
	}

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.L().Error("agent reply worker panic", zap.Any("panic", recovered))
			}
		}()
		s.handleUserMessage(context.Background(), task)
	}()
}

func (s *AgentService) handleUserMessage(parentCtx context.Context, task AgentReplyTask) {
	if s == nil || s.idGenerator == nil {
		return
	}

	streamID := "agent-stream-" + strconv.FormatInt(s.idGenerator.NextID(), 10)
	replyClientMsgID := "agent-" + strconv.FormatInt(s.idGenerator.NextID(), 10)
	conversationID := formatID(task.ConversationID)
	userID := formatID(task.UserID)

	s.notifyAgentStreamStart(parentCtx, task.UserID, AgentMessageStartOutput{
		StreamID:       streamID,
		ConversationID: conversationID,
		ClientMsgID:    replyClientMsgID,
		SenderID:       formatID(task.AgentUserID),
		MessageType:    model.MessageTypeText,
		CreatedAt:      formatTime(now()),
	})

	candidates := make([]string, 0, 8)
	finalCandidate := ""
	chunkIndex := 0
	streamCompleted := false
	var streamErr error

	if s.options.Enabled && s.client != nil {
		callCtx, cancel := context.WithTimeout(parentCtx, s.options.APITimeout)
		err := s.client.ChatStream(callCtx, agentclient.ChatRequest{
			Message:        task.Message,
			ConversationID: conversationID,
			UserID:         userID,
		}, func(event agentclient.StreamEvent) error {
			if event.Done {
				streamCompleted = true
				return nil
			}
			if strings.TrimSpace(event.Error) != "" {
				streamErr = errors.New(strings.TrimSpace(event.Error))
				return nil
			}

			item := normalizeAgentStreamItem(event.Content)
			if shouldSkipAgentStreamItem(item, task.Message) {
				return nil
			}

			snapshot, mermaidPending := buildAgentDisplaySnapshot(finalCandidate, item)
			if strings.TrimSpace(snapshot) == "" && !mermaidPending {
				return nil
			}

			finalCandidate = snapshot
			candidates = append(candidates, snapshot)
			chunkIndex++
			s.notifyAgentStreamChunk(parentCtx, task.UserID, AgentMessageChunkOutput{
				StreamID:       streamID,
				ConversationID: conversationID,
				ClientMsgID:    replyClientMsgID,
				Content:        snapshot,
				ChunkIndex:     chunkIndex,
				Mode:           "replace",
				MermaidPending: mermaidPending,
			})
			return nil
		})
		cancel()
		if err != nil {
			streamErr = err
		}
	} else {
		streamErr = errors.New("agent integration is disabled")
		logger.L().Warn("agent chat stream skipped because agent integration is disabled",
			zap.Int64("conversation_id", task.ConversationID),
			zap.Int64("user_id", task.UserID),
		)
	}

	reply := chooseFinalAgentContent(candidates)
	streamFailed := streamErr != nil || !streamCompleted || strings.TrimSpace(reply) == ""
	if streamFailed {
		if streamErr != nil {
			logger.L().Warn("agent chat stream failed",
				zap.Int64("conversation_id", task.ConversationID),
				zap.Int64("user_id", task.UserID),
				zap.Error(streamErr),
			)
		}
		if strings.TrimSpace(reply) == "" {
			reply = s.options.FailureMessage
		} else {
			reply = strings.TrimSpace(reply) + "\n\n（生成中断，请稍后重试。）"
		}
		s.notifyAgentStreamError(parentCtx, task.UserID, AgentMessageErrorOutput{
			StreamID:       streamID,
			ConversationID: conversationID,
			ClientMsgID:    replyClientMsgID,
			Code:           "agent_stream_failed",
			Message:        s.options.FailureMessage,
		})
	}

	persistCtx, cancel := context.WithTimeout(parentCtx, s.options.APITimeout)
	defer cancel()

	message, err := s.createAgentTextMessageWithClientMsgID(persistCtx, task.ConversationID, task.AgentUserID, task.UserID, reply, replyClientMsgID)
	if err != nil {
		logger.L().Warn("agent reply persist failed",
			zap.Int64("conversation_id", task.ConversationID),
			zap.Int64("user_id", task.UserID),
			zap.Error(err),
		)
		return
	}

	if s.notifier != nil {
		receive := *toReceiveOutput(message)
		if !streamFailed {
			s.notifyAgentStreamDone(persistCtx, task.UserID, AgentMessageDoneOutput{
				StreamID:       streamID,
				ConversationID: conversationID,
				ClientMsgID:    replyClientMsgID,
				Message:        receive,
			})
		}
		if err := s.notifier.NotifyAgentMessage(persistCtx, task.UserID, receive); err != nil {
			logger.L().Warn("agent reply websocket push failed",
				zap.Int64("conversation_id", task.ConversationID),
				zap.Int64("user_id", task.UserID),
				zap.Int64("message_id", message.MessageID),
				zap.Error(err),
			)
		}
	}
}

func (s *AgentService) createAgentTextMessage(ctx context.Context, conversationID int64, agentUserID int64, userID int64, content string) (*model.Message, error) {
	return s.createAgentTextMessageWithClientMsgID(ctx, conversationID, agentUserID, userID, content, "")
}

func (s *AgentService) createAgentTextMessageWithClientMsgID(ctx context.Context, conversationID int64, agentUserID int64, userID int64, content string, clientMsgID string) (*model.Message, error) {
	if s.messageRepo == nil || s.idGenerator == nil {
		return nil, errors.New("agent message dependencies unavailable")
	}
	content = normalizeAgentReplyContent(content)
	if content == "" {
		content = s.options.FailureMessage
	}

	messageID := s.idGenerator.NextID()
	clientMsgID = strings.TrimSpace(clientMsgID)
	if clientMsgID == "" {
		clientMsgID = "agent-" + strconv.FormatInt(messageID, 10)
	}
	message := &model.Message{
		MessageID:      messageID,
		ConversationID: conversationID,
		SenderID:       agentUserID,
		ClientMsgID:    clientMsgID,
		MessageType:    model.MessageTypeText,
		Content:        sql.NullString{String: content, Valid: true},
		ExtraJSON:      sql.NullString{String: buildAgentTextExtraJSON(content), Valid: true},
		SendStatus:     model.MessageSendStatusSent,
		CreatedAt:      now(),
	}

	if err := s.messageRepo.CreatePrivateMessage(ctx, message, userID); err != nil {
		return nil, err
	}
	return message, nil
}

func (s *AgentService) notifyAgentStreamStart(ctx context.Context, userID int64, data AgentMessageStartOutput) {
	if s == nil || s.notifier == nil {
		return
	}
	if err := s.notifier.NotifyAgentMessageStreamStart(ctx, userID, data); err != nil {
		logger.L().Warn("agent stream start websocket push failed",
			zap.String("stream_id", data.StreamID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
	}
}

func (s *AgentService) notifyAgentStreamChunk(ctx context.Context, userID int64, data AgentMessageChunkOutput) {
	if s == nil || s.notifier == nil {
		return
	}
	if err := s.notifier.NotifyAgentMessageStreamChunk(ctx, userID, data); err != nil {
		logger.L().Warn("agent stream chunk websocket push failed",
			zap.String("stream_id", data.StreamID),
			zap.Int("chunk_index", data.ChunkIndex),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
	}
}

func (s *AgentService) notifyAgentStreamDone(ctx context.Context, userID int64, data AgentMessageDoneOutput) {
	if s == nil || s.notifier == nil {
		return
	}
	if err := s.notifier.NotifyAgentMessageStreamDone(ctx, userID, data); err != nil {
		logger.L().Warn("agent stream done websocket push failed",
			zap.String("stream_id", data.StreamID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
	}
}

func (s *AgentService) notifyAgentStreamError(ctx context.Context, userID int64, data AgentMessageErrorOutput) {
	if s == nil || s.notifier == nil {
		return
	}
	if err := s.notifier.NotifyAgentMessageStreamError(ctx, userID, data); err != nil {
		logger.L().Warn("agent stream error websocket push failed",
			zap.String("stream_id", data.StreamID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
	}
}

func normalizeAgentStreamItem(raw string) string {
	value := strings.ReplaceAll(raw, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.TrimSpace(value)
}

func shouldSkipAgentStreamItem(item string, userPrompt string) bool {
	trimmed := strings.TrimSpace(item)
	if trimmed == "" || trimmed == "[DONE]" {
		return true
	}
	if trimmed == strings.TrimSpace(userPrompt) {
		return true
	}

	lower := strings.ToLower(trimmed)
	skippedPrefixes := []string{
		"human:",
		"user:",
		"input:",
		"observation:",
		"tool observation:",
	}
	for _, prefix := range skippedPrefixes {
		if strings.HasPrefix(lower, prefix) && !containsMermaidReference(trimmed) {
			return true
		}
	}
	return false
}

func buildAgentDisplaySnapshot(currentCandidate string, item string) (string, bool) {
	item = strings.TrimSpace(item)
	if item == "" {
		return "", false
	}

	current := strings.TrimSpace(currentCandidate)
	if current == "" {
		return item, hasMermaidPending(item)
	}

	snapshot := mergeAgentStreamSnapshot(current, item)
	return snapshot, hasMermaidPending(snapshot)
}

func mergeAgentStreamSnapshot(current string, item string) string {
	if current == "" {
		return item
	}
	if item == "" {
		return current
	}

	normalizedCurrent := normalizeStreamComparable(current)
	normalizedItem := normalizeStreamComparable(item)
	if normalizedItem == "" {
		return current
	}
	if normalizedCurrent == normalizedItem || strings.Contains(normalizedCurrent, normalizedItem) {
		return current
	}
	if strings.Contains(normalizedItem, normalizedCurrent) {
		return item
	}
	if containsMermaidReference(item) && isMermaidOnlyItem(item) {
		return current + "\n\n" + item
	}
	return current + "\n\n" + item
}

func chooseFinalAgentContent(candidates []string) string {
	for i := len(candidates) - 1; i >= 0; i-- {
		if strings.TrimSpace(candidates[i]) != "" {
			return strings.TrimSpace(candidates[i])
		}
	}
	return ""
}

func normalizeStreamComparable(value string) string {
	fields := strings.Fields(strings.TrimSpace(value))
	return strings.Join(fields, " ")
}

func containsMermaidReference(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "mermaid.ink/img/") ||
		strings.Contains(lower, "mermaid.ink/svg/") ||
		strings.Contains(lower, "```mermaid")
}

func isMermaidOnlyItem(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !containsMermaidReference(trimmed) {
		return false
	}

	withoutImages := markdownImagePattern.ReplaceAllString(trimmed, "")
	if strings.TrimSpace(withoutImages) == "" {
		return true
	}

	lower := strings.ToLower(trimmed)
	return (strings.HasPrefix(lower, "https://mermaid.ink/img/") ||
		strings.HasPrefix(lower, "https://mermaid.ink/svg/")) &&
		!strings.Contains(trimmed, " ")
}

func hasMermaidPending(value string) bool {
	trimmed := strings.TrimSpace(value)
	lower := strings.ToLower(trimmed)

	if index := strings.LastIndex(lower, "```mermaid"); index >= 0 {
		afterFenceStart := lower[index+len("```mermaid"):]
		if !strings.Contains(afterFenceStart, "```") {
			return true
		}
	}

	if strings.Contains(lower, "![") && strings.Contains(lower, "(https://mermaid.ink") {
		lastImageStart := strings.LastIndex(lower, "![")
		if lastImageStart >= 0 && !strings.Contains(lower[lastImageStart:], ")") {
			return true
		}
	}

	return strings.HasSuffix(lower, "https://mermaid.ink") ||
		strings.HasSuffix(lower, "https://mermaid.ink/") ||
		strings.HasSuffix(lower, "https://mermaid.ink/img") ||
		strings.HasSuffix(lower, "https://mermaid.ink/img/") ||
		strings.HasSuffix(lower, "https://mermaid.ink/svg") ||
		strings.HasSuffix(lower, "https://mermaid.ink/svg/")
}

func normalizeAgentReplyContent(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	return mermaidFencePattern.ReplaceAllStringFunc(content, func(block string) string {
		matches := mermaidFencePattern.FindStringSubmatch(block)
		if len(matches) < 2 {
			return block
		}

		diagram := strings.TrimSpace(matches[1])
		if diagram == "" {
			return block
		}

		encoded := base64.RawURLEncoding.EncodeToString([]byte(diagram))
		return "![Mermaid diagram](https://mermaid.ink/img/" + encoded + ")"
	})
}

func buildAgentTextExtraJSON(content string) string {
	images := extractMarkdownImages(content)
	if len(images) == 0 {
		return "{}"
	}

	raw, err := json.Marshal(agentTextExtra{MarkdownImages: images})
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func extractMarkdownImages(content string) []agentMarkdownImage {
	matches := markdownImagePattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return nil
	}

	images := make([]agentMarkdownImage, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if len(match) < 3 || !isHTTPURL(match[2]) {
			continue
		}
		key := match[1] + "\x00" + match[2]
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		images = append(images, agentMarkdownImage{
			Alt: strings.TrimSpace(match[1]),
			URL: match[2],
		})
	}
	return images
}

func isHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func (s *AgentService) cachedDefaultAgentID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.defaultAgentID
}

func (s *AgentService) setDefaultAgentID(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultAgentID = userID
}

func randomAgentPasswordHash() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return password.Hash(hex.EncodeToString(raw))
}
