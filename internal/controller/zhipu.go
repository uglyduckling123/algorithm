package controller

import "C"
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mango/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ZhiPuHandler struct {
}

func NewZhiPuHandler(userService service.UserService, textService service.TextRiskLogService) *ZhiPuHandler {
	return &ZhiPuHandler{}
}

// Register 注册路由
func (v *ZhiPuHandler) Register(router *gin.RouterGroup) {
	userRouter := router.Group("/zhipu")
	{
		userRouter.GET("/steam", v.stream)
	}
}

// 请求和响应数据结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Prompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type CompletionRequest struct {
	Messages        []Messages                                    `json:"messages"`
	RequestId       *string                                       `json:"request_id"`
	Model           string                                        `json:"model"`
	MaxTokens       int                                           `json:"max_tokens,omitempty"`
	Temperature     float64                                       `json:"temperature,omitempty"`
	TopP            float64                                       `json:"top_p,omitempty"`
	StopWords       []string                                      `json:"stop,omitempty"`
	Stream          bool                                          `json:"stream,omitempty"`
	SmoothingConfig *SmoothingConfig                              `json:"smoothing_config,omitempty"`
	StreamingFunc   func(ctx context.Context, chunk []byte) error `json:"-"`
}

type SmoothingConfig struct {
	Length       int `json:"length,omitempty"`
	BaseInterval int `json:"base_interval,omitempty"`
	MinInterval  int `json:"min_interval,omitempty"`
	MaxInterval  int `json:"max_interval,omitempty"`
}

type Messages struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (v *ZhiPuHandler) stream(c *gin.Context) {
	//apiKey := "7sgllz66HxpqnPGP" // 替换为你的API Key
	model := "glm-4-air" // 替换为你要调用的模型名称

	// 创建符合要求的 CompletionRequest 实例
	requestId := "req_123456" // 示例请求ID

	request := CompletionRequest{
		Messages: []Messages{
			{
				Role:    "human",
				Content: "你好",
			},
		},
		Model:       model, // 示例模型名称
		Temperature: 0.7,
		Stream:      true,
		RequestId:   &requestId, // 使用指针传递请求ID
		// 可选设置其他字段
		MaxTokens: 1024,
		TopP:      0.9,
		StopWords: []string{"\n", "。", "！"},
	}

	// 设置流式回调函数
	request.StreamingFunc = func(ctx context.Context, chunk []byte) error {
		fmt.Printf("收到流式响应: %s\n", chunk)
		return nil
	}

	// 如果需要，可以设置 StreamingFunc
	request.StreamingFunc = func(ctx context.Context, chunk []byte) error {
		fmt.Printf("收到流式响应: %s\n", chunk)
		return nil
	}

	// 将请求数据转换为JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request data: %v\n", err)
		return
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(
		context.Background(),
		"GET",
		"https://open.bigmodel.cn/api/paas/v3/model-api/sse-invoke/completions",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	fmt.Printf("Error decoding response chunk: %v\n", err)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("API request failed with status %d: %s\n", resp.StatusCode, string(body))
		return
	}

	// 流式读取响应
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk ChatCompletionChunk
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error decoding response chunk: %v\n", err)
			return
		}

		// 打印每个chunk的内容
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
}
