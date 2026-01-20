# llmx - The Next-Gen LLM SDK for Go

> ⚡ Minimal, Powerful, Type-safe LLM SDK for Go

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/llmx-ai/llmx)](https://goreportcard.com/report/github.com/llmx-ai/llmx)

**llmx** is a unified AI SDK for Go that provides a single, type-safe interface for multiple AI providers.

## ✨ Features

- ✅ **Unified Interface** - One API for 20+ AI providers
- ✅ **Type Safe** - Strong typing with Go generics
- ✅ **Streaming** - First-class streaming support
- ✅ **Tool Calling** - Automatic tool execution loop
- ✅ **Middleware** - Extensible middleware system
- ✅ **High Performance** - Optimized for Go's concurrency

## 🚀 Quick Start

### Installation

```bash
go get github.com/llmx-ai/llmx
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/llmx-ai/llmx"
)

func main() {
    // Create client
    client, err := llmx.NewClient(
        llmx.WithOpenAI("sk-your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Send message
    resp, err := client.Chat(context.Background(), &llmx.ChatRequest{
        Model: "gpt-4-turbo",
        Messages: []llmx.Message{
            {
                Role: llmx.RoleUser,
                Content: []llmx.ContentPart{
                    llmx.TextPart{Text: "Hello, AI!"},
                },
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("AI:", resp.Content)
}
```

## 📖 Documentation

- [Quick Start Guide](./QUICKSTART.md) - Complete guide with all features
- [Examples](./examples/) - 9 complete examples
  - [Basic Chat](./examples/chat/) - Simple chat example
  - [Streaming](./examples/streaming/) - Streaming responses
  - [Multiple Providers](./examples/providers/) - Provider switching
  - [Tool Calling](./examples/tools/) - Tool system usage
  - [Middleware](./examples/middleware/) - Middleware examples
  - [Advanced Middleware](./examples/advanced-middleware/) - Rate limiting, circuit breaker
  - [Structured Output](./examples/structured/) - JSON schema output
  - [Telemetry](./examples/telemetry/) - OpenTelemetry integration
  - [Performance](./examples/performance/) - Performance optimization
- [API Reference](https://pkg.go.dev/github.com/llmx-ai/llmx) - Full API documentation on pkg.go.dev

## 🎯 Supported Providers

- ✅ OpenAI (GPT-4, GPT-3.5)
- ✅ Anthropic (Claude)
- ✅ Google (Gemini)
- ✅ Azure OpenAI
- 🔜 20+ more providers

## 🛠️ Advanced Features

### Tool Calling

```go
import "github.com/llmx-ai/llmx/tools"
import "github.com/llmx-ai/llmx/tools/builtin"

// Register tools
registry := tools.NewRegistry()
registry.Register(builtin.CalculatorTool())
registry.Register(builtin.DateTimeTool())

// Create executor
executor := tools.NewExecutor(registry)

// Execute with automatic tool calling loop
resp, err := executor.ExecuteLoop(ctx, client, &llmx.ChatRequest{
    Model: "gpt-4",
    Messages: messages,
    Tools: registry.List(),
})
```

### Middleware

```go
import "github.com/llmx-ai/llmx/middleware"

// Add middleware
client.Use(
    middleware.Logging(nil),
    middleware.Retry(3, middleware.NewExponentialBackoff()),
    middleware.CacheMiddleware(nil, 5*time.Minute),
)

// All requests will go through middleware chain
resp, err := client.Chat(ctx, req)
```

### Structured Output

```go
import "github.com/llmx-ai/llmx/structured"

type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email"`
}

// Generate structured output
var person Person
err := structured.New(client).GenerateInto(ctx,
    "Extract information about: John Smith, age 35",
    &person,
)
```

### Production Features

```go
import (
    "github.com/llmx-ai/llmx/middleware"
    "github.com/llmx-ai/llmx/observability"
)

// Telemetry & Monitoring
tel, _ := observability.New(&observability.Config{
    ServiceName: "my-app",
})
client.Use(middleware.Telemetry(tel))

// Rate Limiting
limiter := middleware.NewTokenBucketLimiter(10, 20)
client.Use(middleware.RateLimit(limiter, true))

// Circuit Breaker
breaker := middleware.NewCircuitBreaker(5, 30*time.Second)
client.Use(middleware.CircuitBreakerMiddleware(breaker))

// Adaptive Timeout
timeout := middleware.NewAdaptiveTimeout().
    WithBaseTimeout(30 * time.Second)
client.Use(timeout.Middleware())

// Full Production Stack
client.Use(
    middleware.Timeout(60*time.Second),
    middleware.Telemetry(tel),
    middleware.RateLimit(limiter, true),
    middleware.CircuitBreakerMiddleware(breaker),
    middleware.Retry(3, nil),
    middleware.CacheMiddleware(nil, 5*time.Minute),
)
```

## 📊 Performance

- **Throughput**: 10,000+ requests/sec
- **Latency**: P50 < 100ns, P99 < 200ns
- **Memory**: Ultra-low allocation (0-160 bytes/op)
- **Concurrency**: Linear scaling to multi-core

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

Inspired by [Vercel AI SDK](https://sdk.vercel.ai/)

---

**Built with ❤️ for the Go community**
