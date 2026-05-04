package agentclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "http://127.0.0.1:8100"

var (
	ErrInvalidRequest  = errors.New("invalid agent request")
	ErrInvalidResponse = errors.New("invalid agent response")
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type ChatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	UserID         string `json:"user_id,omitempty"`
}

type ChatResponse struct {
	Reply          string `json:"reply"`
	ConversationID string `json:"conversation_id"`
	Status         string `json:"status"`
}

type StreamEvent struct {
	Content string
	Done    bool
	Error   string
}

func New(baseURL string, timeout time.Duration) *Client {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	if c == nil || c.httpClient == nil || strings.TrimSpace(c.baseURL) == "" {
		return nil, ErrInvalidRequest
	}
	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		return nil, ErrInvalidRequest
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("agent chat failed with status %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}
	chatResp.Reply = strings.TrimSpace(chatResp.Reply)
	if chatResp.Reply == "" || strings.TrimSpace(chatResp.Status) != "success" {
		return nil, ErrInvalidResponse
	}

	return &chatResp, nil
}

func (c *Client) ChatStream(ctx context.Context, request ChatRequest, onEvent func(StreamEvent) error) error {
	if c == nil || c.httpClient == nil || strings.TrimSpace(c.baseURL) == "" || onEvent == nil {
		return ErrInvalidRequest
	}
	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		return ErrInvalidRequest
	}

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat/stream", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("agent chat stream failed with status %d", resp.StatusCode)
	}

	return parseTolerantSSE(resp.Body, onEvent)
}

func parseTolerantSSE(reader io.Reader, onEvent func(StreamEvent) error) error {
	buffered := bufio.NewReader(reader)
	var current strings.Builder
	haveEvent := false
	pendingBlankLines := 0
	sawDone := false

	emit := func() error {
		if !haveEvent {
			return nil
		}
		raw := current.String()
		current.Reset()
		haveEvent = false
		pendingBlankLines = 0

		if strings.TrimSpace(raw) == "" {
			return nil
		}
		control := strings.TrimSpace(raw)
		if control == "[DONE]" {
			sawDone = true
			return onEvent(StreamEvent{Done: true})
		}
		if strings.HasPrefix(control, "[ERROR] ") {
			sawDone = true
			return onEvent(StreamEvent{Error: strings.TrimSpace(strings.TrimPrefix(control, "[ERROR] "))})
		}
		return onEvent(StreamEvent{Content: raw})
	}

	for {
		line, err := buffered.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			switch {
			case strings.HasPrefix(line, "data:"):
				if emitErr := emit(); emitErr != nil {
					return emitErr
				}
				haveEvent = true
				value := strings.TrimPrefix(line, "data:")
				if strings.HasPrefix(value, " ") {
					value = value[1:]
				}
				current.WriteString(value)
			case line == "":
				if haveEvent {
					pendingBlankLines++
				}
			case haveEvent:
				for i := 0; i <= pendingBlankLines; i++ {
					current.WriteByte('\n')
				}
				pendingBlankLines = 0
				current.WriteString(line)
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				if emitErr := emit(); emitErr != nil {
					return emitErr
				}
				if !sawDone {
					return io.ErrUnexpectedEOF
				}
				return nil
			}
			return err
		}
	}
}
