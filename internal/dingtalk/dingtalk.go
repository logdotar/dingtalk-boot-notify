// Package dingtalk 提供钉钉 webhook 消息发送客户端。
package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 消息类型常量
const (
	MsgTypeText       = "text"
	MsgTypeMarkdown   = "markdown"
	MsgTypeLink       = "link"
	MsgTypeActionCard = "actionCard"
	MsgTypeFeedCard   = "feedCard"
)

// HTTPClient 是发送 HTTP 请求的接口，用于在测试中 mock。
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client 是钉钉 webhook 客户端，通过自定义机器人 webhook API 向钉钉群发送消息。
type Client struct {
	webhookURL string
	secret     string
	httpClient HTTPClient
}

// Response 表示钉钉 API 响应。
type Response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// TextMessage 文本消息结构。
type TextMessage struct {
	Content string `json:"content"`
}

// MarkdownMessage Markdown 消息结构。
type MarkdownMessage struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// LinkMessage 链接消息结构。
type LinkMessage struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageUrl"`
	PicURL     string `json:"picUrl,omitempty"`
}

// ActionCardMessage 动作卡片消息结构。
type ActionCardMessage struct {
	Title          string `json:"title"`
	Text           string `json:"text"`
	BtnOrientation string `json:"btnOrientation,omitempty"`
	SingleTitle    string `json:"singleTitle,omitempty"`
	SingleURL      string `json:"singleURL,omitempty"`
}

// FeedCardMessage FeedCard 消息结构。
type FeedCardMessage struct {
	Links []FeedCardLink `json:"links"`
}

// FeedCardLink FeedCard 链接项。
type FeedCardLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageUrl"`
	PicURL     string `json:"picUrl"`
}

// AtInfo @ 信息结构。
type AtInfo struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	AtUserIds []string `json:"atUserIds,omitempty"`
	IsAtAll   bool     `json:"isAtAll"`
}

// MessagePayload 消息负载结构。
type MessagePayload struct {
	MsgType    string      `json:"msgtype"`
	Text       interface{} `json:"text,omitempty"`
	Markdown   interface{} `json:"markdown,omitempty"`
	Link       interface{} `json:"link,omitempty"`
	ActionCard interface{} `json:"actionCard,omitempty"`
	FeedCard   interface{} `json:"feedCard,omitempty"`
	At         *AtInfo     `json:"at,omitempty"`
}

// NewClient 创建钉钉客户端。
//
//	webhookURL: 钉钉机器人完整 webhook URL（必填）
//	secret:     机器人加签秘钥（可选，可为空）
//
// 返回配置好的 Client 实例。
func NewClient(webhookURL, secret string) *Client {
	return &Client{
		webhookURL: webhookURL,
		secret:     secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendText 发送文本消息。
//
//	ctx:     用于取消和超时控制的上下文
//	content: 消息内容
//	atInfo:  @ 信息（可选，传 nil 表示不 @ 任何人）
//
// 错误返回条件：
//   - webhook URL 为空
//   - HTTP 请求创建或执行失败
//   - 响应状态码非 200
//   - 钉钉 API 返回非零错误码
func (c *Client) SendText(ctx context.Context, content string, atInfo *AtInfo) error {
	payload := MessagePayload{
		MsgType: MsgTypeText,
		Text:    TextMessage{Content: content},
		At:      atInfo,
	}
	return c.send(ctx, payload)
}

// SendMarkdown 发送 Markdown 消息。
//
//	ctx:     用于取消和超时控制的上下文
//	title:   消息标题（首屏会话透出）
//	text:    Markdown 格式的消息内容
//	atInfo:  @ 信息（可选，传 nil 表示不 @ 任何人）
//
// 支持的 Markdown 语法：
//   - 标题（一级～六级）
//   - 引用
//   - 加粗、斜体
//   - 链接
//   - 图片
//   - 无序列表、有序列表
func (c *Client) SendMarkdown(ctx context.Context, title, text string, atInfo *AtInfo) error {
	payload := MessagePayload{
		MsgType: MsgTypeMarkdown,
		Markdown: MarkdownMessage{
			Title: title,
			Text:  text,
		},
		At: atInfo,
	}
	return c.send(ctx, payload)
}

// SendLink 发送链接消息。
//
//	ctx:        用于取消和超时控制的上下文
//	title:      消息标题
//	text:       消息内容（太长只会部分展示）
//	messageURL: 点击消息跳转的 URL
//	picURL:     图片 URL（可选）
func (c *Client) SendLink(ctx context.Context, title, text, messageURL, picURL string) error {
	payload := MessagePayload{
		MsgType: MsgTypeLink,
		Link: LinkMessage{
			Title:      title,
			Text:       text,
			MessageURL: messageURL,
			PicURL:     picURL,
		},
	}
	return c.send(ctx, payload)
}

// SendActionCard 发送单个按钮的动作卡片消息。
//
//	ctx:            用于取消和超时控制的上下文
//	title:          消息标题
//	text:           消息内容（支持 Markdown）
//	singleTitle:    按钮标题
//	singleURL:      点击按钮跳转的 URL
//	btnOrientation: 按钮排列方向（"0"竖直 "1"横向）
func (c *Client) SendActionCard(ctx context.Context, title, text, singleTitle, singleURL, btnOrientation string) error {
	payload := MessagePayload{
		MsgType: MsgTypeActionCard,
		ActionCard: ActionCardMessage{
			Title:          title,
			Text:           text,
			BtnOrientation: btnOrientation,
			SingleTitle:    singleTitle,
			SingleURL:      singleURL,
		},
	}
	return c.send(ctx, payload)
}

// SendFeedCard 发送 FeedCard 消息。
//
//	ctx:   用于取消和超时控制的上下文
//	links: 链接列表（最多 10 个）
func (c *Client) SendFeedCard(ctx context.Context, links []FeedCardLink) error {
	payload := MessagePayload{
		MsgType: MsgTypeFeedCard,
		FeedCard: FeedCardMessage{
			Links: links,
		},
	}
	return c.send(ctx, payload)
}

// Send 发送文本消息（兼容旧接口）。
//
//	ctx:     用于取消和超时控制的上下文
//	message: 要发送的文本消息内容
//
// 错误返回条件：
//   - webhook URL 为空
//   - HTTP 请求创建或执行失败
//   - 响应状态码非 200
//   - 钉钉 API 返回非零错误码
func (c *Client) Send(ctx context.Context, message string) error {
	return c.SendText(ctx, message, nil)
}

// send 发送消息负载到钉钉 webhook。
func (c *Client) send(ctx context.Context, payload MessagePayload) error {
	if c.webhookURL == "" {
		return fmt.Errorf("webhook URL 不能为空")
	}

	fullURL := c.webhookURL

	if c.secret != "" {
		timestamp := time.Now().UnixMilli()
		signature := c.generateSignature(timestamp)
		separator := "&"
		if !strings.Contains(c.webhookURL, "?") {
			separator = "?"
		}
		fullURL = fmt.Sprintf("%s%stimestamp=%d&sign=%s", c.webhookURL, separator, timestamp, signature)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化消息内容失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var dingResp Response
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if dingResp.ErrCode != 0 {
		return fmt.Errorf("钉钉错误码=%d, 错误信息=%s", dingResp.ErrCode, dingResp.ErrMsg)
	}

	return nil
}

// generateSignature 生成钉钉 HMAC-SHA256 签名。
//
//	timestamp: 当前 Unix 时间戳（毫秒）
//
// 签名算法：timestamp + "\n" + secret 作为签名字符串，
// 使用 HmacSHA256 算法计算签名，然后进行 Base64 encode。
func (c *Client) generateSignature(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, c.secret)
	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(stringToSign))
	signBytes := mac.Sum(nil)
	return url.QueryEscape(base64.StdEncoding.EncodeToString(signBytes))
}
