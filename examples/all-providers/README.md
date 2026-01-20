# All Providers Example

这个示例展示了如何使用 llmx 的所有支持的 Providers。

## 🎯 支持的 Providers

### 国际主流
- OpenAI (GPT-4)
- Anthropic (Claude)
- Google (Gemini)
- Groq (Llama 3.3 70B)
- Mistral AI
- Azure OpenAI

### 国内厂商
- DeepSeek (深度求索)
- 智谱 AI (GLM)
- 通义千问 (Qwen)
- 文心一言 (ERNIE) *
- 豆包 (Doubao) *

### 开源生态
- Ollama (本地运行)
- Hugging Face
- LocalAI
- LM Studio
- vLLM

\* = 需要完整 SDK 集成

## 🚀 运行示例

### 1. 设置环境变量

```bash
# OpenAI
export OPENAI_API_KEY="sk-..."

# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# Google
export GOOGLE_API_KEY="AIza..."

# Groq
export GROQ_API_KEY="gsk_..."

# DeepSeek
export DEEPSEEK_API_KEY="sk-..."

# Mistral
export MISTRAL_API_KEY="..."

# 智谱 AI
export ZHIPU_API_KEY="..."

# 通义千问
export DASHSCOPE_API_KEY="sk-..."

# Hugging Face
export HF_TOKEN="hf_..."
```

### 2. 启动本地服务（可选）

```bash
# Ollama
ollama serve

# 下载模型
ollama pull llama3.3
```

### 3. 运行示例

```bash
go run main.go
```

## 📝 输出示例

```
🌐 LLMX - All Providers Demo
================================

📌 OpenAI (GPT-4)
   ✅ 我是 GPT-4，OpenAI 开发的大型语言模型...

📌 Anthropic (Claude)
   ✅ 我是 Claude，由 Anthropic 开发的 AI 助手...

📌 Groq (Llama 3.3 70B)
   ✅ 我是 Llama 3.3，Meta 开发的开源大语言模型...

📌 DeepSeek (高性价比)
   ✅ 我是 DeepSeek，专注于高性价比的 AI 服务...

📌 智谱 AI (GLM-4 Plus)
   ✅ 我是 GLM-4，清华大学研发的双语对话模型...

📌 Ollama (本地)
   ✅ 我是一个在你本地运行的 AI 模型...

✅ 测试完成！
```

## 🔧 切换 Provider

只需更改 Builder 的配置即可：

```go
// 从 OpenAI 切换到 Groq
client := llmx.NewClientBuilder().
    // OpenAI(apiKey).
    Groq(apiKey).  // 只需更改这一行
    Model("llama-3.3-70b-versatile").
    Build()

// 从云端切换到本地
client := llmx.NewClientBuilder().
    // Anthropic(apiKey).
    Ollama("http://localhost:11434").  // 切换到本地
    Model("llama3.3").
    Build()
```

## 💡 Provider 选择建议

### 性能优先
- **Groq**: 极快推理速度 (500+ tokens/s)
- **vLLM**: 本地高性能推理

### 成本优先
- **DeepSeek**: 比 GPT-4 便宜 95%
- **Ollama**: 完全免费（本地运行）

### 质量优先
- **OpenAI GPT-4**: 综合能力最强
- **Claude 3.5**: 长文本理解优秀
- **Gemini 1.5 Pro**: 超长上下文 (2M)

### 中文优化
- **智谱 GLM-4**: 中文理解优秀
- **通义千问**: 阿里云生态
- **DeepSeek**: 中文性能突出

### 本地/私有部署
- **Ollama**: 最简单易用
- **LocalAI**: OpenAI 兼容
- **LM Studio**: 图形界面友好

## 🌟 特性对比

| Provider | 速度 | 成本 | 质量 | 中文 | 本地 |
|----------|------|------|------|------|------|
| OpenAI | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ❌ |
| Claude | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ❌ |
| Groq | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ❌ |
| DeepSeek | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ |
| 智谱 GLM | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ |
| Ollama | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ✅ |

## 📚 更多信息

- [Provider 详细文档](../../PROVIDER_EXPANSION_PLAN.md)
- [API 参考](https://pkg.go.dev/github.com/llmx-ai/llmx)
- [Quick Start](../../QUICKSTART.md)
