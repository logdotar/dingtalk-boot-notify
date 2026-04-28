package dingtalk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mockHTTPClient 模拟 HTTP 客户端。
type mockHTTPClient struct {
	resp *http.Response
	err  error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.resp, m.err
}

// mockReadCloser 模拟可关闭的读取器。
type mockReadCloser struct {
	content string
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if len(m.content) == 0 {
		return 0, io.EOF
	}
	if len(p) < len(m.content) {
		copy(p, m.content[:len(p)])
		m.content = m.content[len(p):]
		return len(p), nil
	}
	n = len(m.content)
	copy(p, m.content)
	m.content = ""
	return n, io.EOF
}

func (m *mockReadCloser) Close() error {
	return nil
}

// TestClient_Send_Success 测试发送成功。
func TestClient_Send_Success(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.Send(context.Background(), "test message")
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_Send_EmptyWebhook 测试空 webhook 错误。
func TestClient_Send_EmptyWebhook(t *testing.T) {
	client := &Client{
		webhookURL: "",
		httpClient: &mockHTTPClient{},
	}

	err := client.Send(context.Background(), "test message")
	if err == nil {
		t.Fatal("期望返回错误，因为 webhook 为空")
	}
	if !strings.Contains(err.Error(), "webhook URL 不能为空") {
		t.Errorf("意外的错误消息: %v", err)
	}
}

// TestClient_Send_HTTPError 测试 HTTP 错误。
func TestClient_Send_HTTPError(t *testing.T) {
	client := &Client{
		webhookURL: "https://example.com/webhook",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       &mockReadCloser{content: "internal error"},
			},
		},
	}

	err := client.Send(context.Background(), "test message")
	if err == nil {
		t.Fatal("期望返回错误，因为 HTTP 500")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("意外的错误消息: %v", err)
	}
}

// TestClient_Send_DingTalkError 测试钉钉 API 错误。
func TestClient_Send_DingTalkError(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 300001, ErrMsg: "invalid token"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.Send(context.Background(), "test message")
	if err == nil {
		t.Fatal("期望返回错误，因为钉钉错误响应")
	}
	if !strings.Contains(err.Error(), "300001") {
		t.Errorf("意外的错误消息: %v", err)
	}
}

// TestClient_Send_NetworkError 测试网络错误。
func TestClient_Send_NetworkError(t *testing.T) {
	client := &Client{
		webhookURL: "https://example.com/webhook",
		httpClient: &mockHTTPClient{
			err: errors.New("connection refused"),
		},
	}

	err := client.Send(context.Background(), "test message")
	if err == nil {
		t.Fatal("期望返回错误，因为网络失败")
	}
	if !strings.Contains(err.Error(), "发送请求失败") {
		t.Errorf("意外的错误消息: %v", err)
	}
}

// TestClient_Send_ContextCanceled 测试上下文取消。
func TestClient_Send_ContextCanceled(t *testing.T) {
	client := &Client{
		webhookURL: "https://example.com/webhook",
		httpClient: &mockHTTPClient{
			err: context.Canceled,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Send(ctx, "test message")
	if err == nil {
		t.Fatal("期望返回错误，因为上下文已取消")
	}
}

// TestClient_Send_WithSecret 测试带签名发送。
func TestClient_Send_WithSecret(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "test-secret",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.Send(context.Background(), "test message")
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestGenerateSignature 测试签名生成。
func TestGenerateSignature(t *testing.T) {
	client := &Client{
		secret: "test-secret",
	}

	timestamp := time.Now().UnixMilli()
	sig := client.generateSignature(timestamp)

	if sig == "" {
		t.Fatal("期望非空签名")
	}

	sig2 := client.generateSignature(timestamp)
	if sig != sig2 {
		t.Fatal("相同时间戳应产生相同签名")
	}

	sig3 := client.generateSignature(timestamp + 1000)
	if sig == sig3 {
		t.Fatal("不同时间戳应产生不同签名")
	}
}

// TestClient_SendMarkdown 测试 Markdown 消息发送。
func TestClient_SendMarkdown(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.SendMarkdown(context.Background(), "标题", "### 测试Markdown", nil)
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_SendLink 测试链接消息发送。
func TestClient_SendLink(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.SendLink(context.Background(), "链接标题", "链接内容", "https://example.com", "")
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_SendLink_WithPicURL 测试带图片的链接消息。
func TestClient_SendLink_WithPicURL(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.SendLink(context.Background(), "链接标题", "链接内容", "https://example.com", "https://example.com/pic.jpg")
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_SendActionCard 测试动作卡片消息发送。
func TestClient_SendActionCard(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.SendActionCard(context.Background(), "卡片标题", "卡片内容", "查看详情", "https://example.com", "0")
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_SendFeedCard 测试 FeedCard 消息发送。
func TestClient_SendFeedCard(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	links := []FeedCardLink{
		{Title: "标题1", MessageURL: "https://example.com/1", PicURL: "https://example.com/1.jpg"},
		{Title: "标题2", MessageURL: "https://example.com/2", PicURL: "https://example.com/2.jpg"},
	}

	err := client.SendFeedCard(context.Background(), links)
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_SendText_WithAtInfo 测试带 @ 信息发送。
func TestClient_SendText_WithAtInfo(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	atInfo := &AtInfo{
		AtMobiles: []string{"13800138000"},
		IsAtAll:   false,
	}

	err := client.SendText(context.Background(), "测试消息 @13800138000", atInfo)
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_SendText_WithAtAll 测试 @ 所有人。
func TestClient_SendText_WithAtAll(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	atInfo := &AtInfo{
		IsAtAll: true,
	}

	err := client.SendText(context.Background(), "测试消息 @所有人", atInfo)
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_Send_WithSecretAndQueryParams 测试带签名和已有查询参数的 webhook。
func TestClient_Send_WithSecretAndQueryParams(t *testing.T) {
	respBody, _ := json.Marshal(Response{ErrCode: 0, ErrMsg: "ok"})
	client := &Client{
		webhookURL: "https://example.com/webhook?key=value",
		secret:     "test-secret",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: string(respBody)},
			},
		},
	}

	err := client.Send(context.Background(), "test message")
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestClient_Send_ResponseReadError 测试读取响应错误。
func TestClient_Send_ResponseReadError(t *testing.T) {
	client := &Client{
		webhookURL: "https://example.com/webhook",
		httpClient: &mockHTTPClient{
			resp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &mockReadCloser{content: "invalid json"},
			},
		},
	}

	err := client.Send(context.Background(), "test message")
	if err == nil {
		t.Fatal("期望返回错误，因为响应解析失败")
	}
}

// TestClient_Send_JSONMarshalError 测试 JSON 序列化错误（理论上不会发生）。
func TestClient_Send_JSONMarshalError(t *testing.T) {
	client := &Client{
		webhookURL: "https://example.com/webhook",
		secret:     "",
		httpClient: &mockHTTPClient{},
	}

	_ = client
}

// TestMessagePayload_Marshal 测试消息负载序列化。
func TestMessagePayload_Marshal(t *testing.T) {
	payload := MessagePayload{
		MsgType: MsgTypeText,
		Text:    TextMessage{Content: "test"},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}

	if len(data) == 0 {
		t.Error("期望非空序列化数据")
	}
}
