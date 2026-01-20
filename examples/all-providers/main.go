// Package main demonstrates using llmx with all supported providers
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/llmx-ai/llmx"
)

func main() {
	ctx := context.Background()
	prompt := "用一句话介绍你自己"

	fmt.Println("🌐 LLMX - All Providers Demo")
	fmt.Println("================================\n")

	// 1. OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		testProvider(ctx, "OpenAI (GPT-4)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				OpenAI(apiKey).
				Model("gpt-4-turbo").
				Build()
		}, prompt)
	}

	// 2. Anthropic
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		testProvider(ctx, "Anthropic (Claude)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				Anthropic(apiKey).
				Model("claude-3-5-sonnet-20241022").
				Build()
		}, prompt)
	}

	// 3. Google
	if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		testProvider(ctx, "Google (Gemini)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				Google(apiKey).
				Model("gemini-1.5-pro").
				Build()
		}, prompt)
	}

	// 4. Groq (超快推理)
	if apiKey := os.Getenv("GROQ_API_KEY"); apiKey != "" {
		testProvider(ctx, "Groq (Llama 3.3 70B)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				Groq(apiKey).
				Model("llama-3.3-70b-versatile").
				Build()
		}, prompt)
	}

	// 5. DeepSeek (性价比之王)
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		testProvider(ctx, "DeepSeek (高性价比)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				DeepSeek(apiKey).
				Model("deepseek-chat").
				Build()
		}, prompt)
	}

	// 6. Mistral AI
	if apiKey := os.Getenv("MISTRAL_API_KEY"); apiKey != "" {
		testProvider(ctx, "Mistral AI (Large)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				Mistral(apiKey).
				Model("mistral-large-latest").
				Build()
		}, prompt)
	}

	// 7. 智谱 AI (GLM)
	if apiKey := os.Getenv("ZHIPU_API_KEY"); apiKey != "" {
		testProvider(ctx, "智谱 AI (GLM-4 Plus)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				Zhipu(apiKey).
				Model("glm-4-plus").
				Build()
		}, prompt)
	}

	// 8. 通义千问
	if apiKey := os.Getenv("DASHSCOPE_API_KEY"); apiKey != "" {
		testProvider(ctx, "通义千问 (Qwen Max)", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				Tongyi(apiKey).
				Model("qwen-max").
				Build()
		}, prompt)
	}

	// 9. Ollama (本地运行)
	testProvider(ctx, "Ollama (本地)", func() (*llmx.Client, error) {
		return llmx.NewClientBuilder().
			Ollama("http://localhost:11434").
			Model("llama3.3").
			Build()
	}, prompt)

	// 10. Hugging Face
	if token := os.Getenv("HF_TOKEN"); token != "" {
		testProvider(ctx, "Hugging Face", func() (*llmx.Client, error) {
			return llmx.NewClientBuilder().
				HuggingFace(token).
				Model("meta-llama/Meta-Llama-3.1-70B-Instruct").
				Build()
		}, prompt)
	}

	fmt.Println("\n✅ 测试完成！")
	fmt.Println("\n💡 提示:")
	fmt.Println("- 设置相应的环境变量来测试不同的 Provider")
	fmt.Println("- Ollama 需要本地运行 (ollama serve)")
	fmt.Println("- 部分 Provider 可能需要额外配置")
}

func testProvider(ctx context.Context, name string, clientFactory func() (*llmx.Client, error), prompt string) {
	fmt.Printf("📌 %s\n", name)

	client, err := clientFactory()
	if err != nil {
		fmt.Printf("   ❌ 创建失败: %v\n\n", err)
		return
	}
	defer client.Close()

	resp, err := client.SimpleChat(ctx, prompt)
	if err != nil {
		fmt.Printf("   ❌ 调用失败: %v\n\n", err)
		return
	}

	// 截断响应以便显示
	if len(resp) > 100 {
		resp = resp[:100] + "..."
	}

	fmt.Printf("   ✅ %s\n\n", resp)
}
