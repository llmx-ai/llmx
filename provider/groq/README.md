# Groq Provider

Groq 是一家提供超快 LLM 推理速度的 AI 服务商，推理速度可达 500+ tokens/s。

## 特性

- ✅ 完全兼容 OpenAI API
- ✅ 极快的推理速度（业界领先）
- ✅ 支持 Chat Completions
- ✅ 支持 Streaming
- ✅ 支持 Function Calling
- ✅ 支持 JSON Mode
- ❌ 不支持视觉模型

## 支持的模型

| 模型 ID | 上下文长度 | 特点 | 价格（每百万 tokens） |
|---------|-----------|------|---------------------|
| `llama-3.3-70b-versatile` | 131K | 最新最强 | $0.59 / $0.79 |
| `llama-3.1-70b-versatile` | 131K | 高性能 | $0.59 / $0.79 |
| `llama-3.1-8b-instant` | 131K | 快速响应 | $0.05 / $0.08 |
| `mixtral-8x7b-32768` | 32K | Mistral 开源 | $0.24 / $0.24 |
| `gemma2-9b-it` | 8K | Google Gemma | $0.20 / $0.20 |

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
    // 创建 Groq 客户端
    client, err := llmx.NewClientBuilder().
        Provider("groq", os.Getenv("GROQ_API_KEY")).
        Model("llama-3.3-70b-versatile").
        Build()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 发送聊天请求
    resp, err := client.SimpleChat(context.Background(), "你好，介绍一下 Groq")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp)
}
```

### 流式响应

```go
stream, err := client.SimpleStreamChat(context.Background(), "用 Go 写一个 HTTP 服务器")
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
    Model("llama-3.3-70b-versatile").
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

### 使用 WithGroq 选项

```go
client, err := llmx.NewClient(
    llmx.WithGroq(os.Getenv("GROQ_API_KEY")),
    llmx.WithDefaultModel("llama-3.3-70b-versatile"),
)
```

### 使用 ClientBuilder

```go
client, err := llmx.NewClientBuilder().
    Groq(os.Getenv("GROQ_API_KEY")).
    Model("llama-3.1-70b-versatile").
    Temperature(0.7).
    MaxTokens(1024).
    Build()
```

### 自定义端点

```go
client, err := llmx.NewClient(
    llmx.WithProvider("groq", map[string]interface{}{
        "api_key":  os.Getenv("GROQ_API_KEY"),
        "base_url": "https://custom-groq-endpoint.com/v1",
    }),
)
```

## 获取 API Key

1. 访问 [Groq Console](https://console.groq.com/)
2. 注册并登录
3. 进入 API Keys 页面
4. 创建新的 API Key

## 性能特点

Groq 的主要优势是**极快的推理速度**：

- **平均速度**: 500+ tokens/s
- **峰值速度**: 1000+ tokens/s
- **延迟**: 通常 < 100ms 首 token

非常适合：
- 实时聊天应用
- 代码生成
- 需要快速响应的场景

## 限制

- ❌ 不支持视觉模型（图像输入）
- ❌ 不支持嵌入（Embeddings）
- ⚠️ 免费额度有限制
- ⚠️ 部分模型可能不支持所有功能

## 价格

Groq 的定价非常有竞争力：

| 模型 | 输入价格 | 输出价格 |
|------|---------|---------|
| Llama 3.3 70B | $0.59/1M tokens | $0.79/1M tokens |
| Llama 3.1 8B | $0.05/1M tokens | $0.08/1M tokens |
| Mixtral 8x7B | $0.24/1M tokens | $0.24/1M tokens |

比 OpenAI 的 GPT-4 便宜约 **20-30 倍**！

## 参考资源

- [Groq 官方网站](https://groq.com/)
- [Groq API 文档](https://console.groq.com/docs)
- [Groq Models](https://console.groq.com/docs/models)
- [定价](https://groq.com/pricing/)
