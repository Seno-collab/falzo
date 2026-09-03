package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"be/internal/alerting"
)

const telegramMessageLimit = 4_096

type Client struct {
	httpClient *http.Client
	endpoint   string
	chatID     string
	token      string
}

func NewClient(baseURL, token, chatID string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(token) == "" || strings.TrimSpace(chatID) == "" {
		return nil, fmt.Errorf("Telegram bot token and chat ID are required")
	}
	parsedBaseURL, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("invalid Telegram API base URL")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		endpoint:   parsedBaseURL.String() + "/bot" + token + "/sendMessage",
		chatID:     chatID,
		token:      token,
	}, nil
}

func (c *Client) SendAlert(ctx context.Context, event alerting.Event) error {
	payload, err := json.Marshal(map[string]any{
		"chat_id":                     c.chatID,
		"text":                        FormatAlert(event),
		"disable_notification":        false,
		"protect_content":             false,
		"allow_sending_without_reply": true,
	})
	if err != nil {
		return fmt.Errorf("encode Telegram message: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create Telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Telegram message: %s", c.redact(err.Error()))
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(response.Body, 2_048))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Telegram API returned %d: %s", response.StatusCode, c.redact(strings.TrimSpace(string(body))))
	}
	var result struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal(body, &result); err != nil || !result.OK {
		return fmt.Errorf("Telegram API returned an invalid response")
	}
	return nil
}

func (c *Client) redact(value string) string {
	return strings.ReplaceAll(value, c.token, "[REDACTED]")
}

func FormatAlert(event alerting.Event) string {
	var builder strings.Builder
	if event.Fields["event_type"] == "account_locked" {
		builder.WriteString("🔒 Falzo account locked\n")
	} else {
		builder.WriteString("🚨 Falzo error\n")
	}
	writeLine(&builder, "Service", event.Service)
	writeLine(&builder, "Environment", event.Environment)
	writeLine(&builder, "Time", event.OccurredAt.UTC().Format(time.RFC3339))
	writeLine(&builder, "Message", event.Message)
	if event.Source != "" {
		writeLine(&builder, "Source", event.Source)
	}
	if event.ID != "" {
		writeLine(&builder, "Event ID", event.ID)
	}

	keys := make([]string, 0, len(event.Fields))
	for key := range event.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if key == "event_type" {
			continue
		}
		writeLine(&builder, key, fmt.Sprint(event.Fields[key]))
	}

	message := builder.String()
	if len([]rune(message)) <= telegramMessageLimit {
		return message
	}
	return string([]rune(message)[:telegramMessageLimit-1]) + "…"
}

func writeLine(builder *strings.Builder, label, value string) {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\x00", "")
	fmt.Fprintf(builder, "%s: %s\n", label, value)
}
