# DeepSeek Provider

DeepSeek 是一家国内领先的 AI 服务商，以其极高的性价比而著称。DeepSeek 的模型在保持高性能的同时，价格比 GPT-4 便宜约 **95%**。

## 特性

- ✅ 完全兼容 OpenAI API
- ✅ 极高的性价比（业界最低价之一）
- ✅ 支持 Chat Completions
- ✅ 支持 Streaming
- ✅ 支持 Function Calling
- ✅ 支持 JSON Mode
- ❌ 暂不支持视觉模型

## 支持的模型

| 模型 ID | 上下文长度 | 特点 | 价格（每百万 tokens） |
|---------|-----------|------|---------------------|
| `deepseek-chat` | 32K | 通用对话模型 | $0.14 / $0.28 |
| `deepseek-coder` | 16K | 代码专用模型 | $0.14 / $0.28 |

**价格优势**：
- 比 GPT-4 Turbo 便宜约 **95%**
- 比 GPT-3.5 Turbo 便宜约 **90%**
- 比 Claude 3 Sonnet 便宜约 **90%**

## 使用示例

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/llmx-ai/llmx"
)

func main() {
    // 创建 DeepSeek 客户端
    client, err := llmx.NewClientBuilder().
        DeepSeek(os.Getenv("DEEPSEEK_API_KEY")).
        Model("deepseek-chat").
        Build()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 发送聊天请求
    resp, err := client.SimpleChat(context.Background(), "你好，介绍一下 DeepSeek")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp)
}
```

### 代码生成（使用 deepseek-coder）

```go
client, err := llmx.NewClientBuilder().
    DeepSeek(os.Getenv("DEEPSEEK_API_KEY")).
    Model("deepseek-coder").  // 使用代码专用模型
    Build()

resp, err := client.SimpleChat(context.Background(), 
    "用 Go 实现一个快速排序算法")
```

### 流式响应

```go
stream, err := client.SimpleStreamChat(context.Background(), 
    "详细解释一下 Go 的 channel 机制")
if err != nil {
    log.Fatal(err)
}

for event := range stream.Events() {
    if event.Type == core.EventTypeTextDelta {
        fmt.Print(event.Delta)
    }
}
```

### Function Calling

```go
req := llmx.NewRequestBuilder().
    Model("deepseek-chat").
    AddUserMessage("北京今天天气怎么样？").
    AddTool(llmx.Tool{
        Name:        "get_weather",
        Description: "获取指定城市的天气信息",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "city": map[string]interface{}{
                    "type":        "string",
                    "description": "城市名称",
                },
            },
            "required": []string{"city"},
        },
    }).
    Build()

resp, err := client.Chat(context.Background(), req)
```

## 配置选项

### 使用 WithDeepSeek 选项

```go
client, err := llmx.NewClient(
    llmx.WithDeepSeek(os.Getenv("DEEPSEEK_API_KEY")),
    llmx.WithDefaultModel("deepseek-chat"),
    llmx.WithTemperature(0.7),
)
```

### 使用 ClientBuilder

```go
client, err := llmx.NewClientBuilder().
    DeepSeek(os.Getenv("DEEPSEEK_API_KEY")).
    Model("deepseek-chat").
    Temperature(0.7).
    MaxTokens(2048).
    Build()
```

### 自定义端点

```go
client, err := llmx.NewClient(
    llmx.WithProvider("deepseek", map[string]interface{}{
        "api_key":  os.Getenv("DEEPSEEK_API_KEY"),
        "base_url": "https://custom-deepseek-endpoint.com/v1",
    }),
)
```

## 获取 API Key

1. 访问 [DeepSeek 开放平台](https://platform.deepseek.com/)
2. 注册并登录
3. 进入 API Keys 页面
4. 创建新的 API Key

## 性能特点

### 优势
- ✅ **极高性价比**: 价格比 GPT-4 便宜 95%
- ✅ **中文优化**: 对中文支持优秀
- ✅ **代码能力**: deepseek-coder 专注代码生成
- ✅ **长上下文**: 支持 32K context

### 适用场景
- 📊 数据分析和处理
- 💬 客服聊天机器人
- 📝 内容生成和总结
- 💻 代码生成和补全
- 🌐 翻译和本地化
- 📚 知识问答

## 限制

- ❌ 不支持视觉模型（图像输入）
- ❌ 不支持嵌入（Embeddings）
- ⚠️ 免费额度有限制
- ⚠️ 响应速度可能不如 Groq

## 价格对比

| 模型 | 输入价格（每百万 tokens） | 输出价格（每百万 tokens） |
|------|------------------------|------------------------|
| **DeepSeek Chat** | **$0.14** | **$0.28** |
| GPT-4 Turbo | $10.00 | $30.00 |
| GPT-3.5 Turbo | $0.50 | $1.50 |
| Claude 3 Sonnet | $3.00 | $15.00 |
| Gemini 1.5 Pro | $3.50 | $10.50 |

**示例成本计算**（处理 1M tokens 输入 + 1M tokens 输出）：
- DeepSeek: $0.42
- GPT-4 Turbo: $40.00 (**节省 99%**)
- GPT-3.5 Turbo: $2.00 (**节省 79%**)

## 最佳实践

### 1. 选择合适的模型

```go
// 通用对话 - 使用 deepseek-chat
client.Model("deepseek-chat")

// 代码生成 - 使用 deepseek-coder
client.Model("deepseek-coder")
```

### 2. 优化 Token 使用

```go
// 设置合理的 max_tokens 避免浪费
client, err := llmx.NewClientBuilder().
    DeepSeek(apiKey).
    Model("deepseek-chat").
    MaxTokens(1024).  // 根据需求设置
    Build()
```

### 3. 错误处理

```go
resp, err := client.Chat(ctx, req)
if err != nil {
    if llmxErr, ok := err.(llmx.Error); ok {
        switch llmxErr.Code() {
        case llmx.ErrorCodeRateLimitExceeded:
            // 处理速率限制
        case llmx.ErrorCodeInsufficientQuota:
            // 处理额度不足
        }
    }
}
```

## 参考资源

- [DeepSeek 官方网站](https://www.deepseek.com/)
- [DeepSeek 开放平台](https://platform.deepseek.com/)
- [API 文档](https://platform.deepseek.com/api-docs/)
- [定价](https://platform.deepseek.com/pricing)
- [模型介绍](https://github.com/deepseek-ai)
