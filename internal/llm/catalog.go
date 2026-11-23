package llm

// Catalog lists the built-in providers: hosted APIs, then local runtimes.
// Model lists change often, so DefaultModel and Models are starting points
// (checked against each provider's docs in November 2026); the settings
// fetch the live list from the provider.
func Catalog() []Provider {
	return []Provider{
		{
			ID: "anthropic", Name: "Anthropic", Kind: Cloud, Protocol: Anthropic, BaseURL: AnthropicBaseURL, NeedsKey: true,
			DefaultModel: "claude-opus-5-5",
			Models:       []string{"claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5", "claude-fable-5-2"},
			KeyEnv:       "ANTHROPIC_API_KEY", DocsURL: "https://platform.claude.com/docs",
		},
		{
			ID: "openai", Name: "OpenAI", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.openai.com/v1", NeedsKey: true,
			TokenField:   "max_completion_tokens",
			DefaultModel: "gpt-6.2-sol",
			Models:       []string{"gpt-6.2-sol", "gpt-6.2-astra", "gpt-6.1-terra", "gpt-6-luna"},
			KeyEnv:       "OPENAI_API_KEY", DocsURL: "https://developers.openai.com/api/docs/models",
		},
		{
			ID: "gemini", Name: "Google Gemini", Kind: Cloud, Protocol: Gemini, BaseURL: GeminiBaseURL, NeedsKey: true,
			DefaultModel: "gemini-4-hydrogen",
			Models:       []string{"gemini-4-hydrogen", "gemini-4-carbon", "gemini-3.8-flash", "gemini-3.5-flash-lite"},
			KeyEnv:       "GEMINI_API_KEY", DocsURL: "https://aistudio.google.com/docs/models",
		},
		{
			ID: "mistral", Name: "Mistral", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.mistral.ai/v1", NeedsKey: true,
			DefaultModel: "mistral-large-4",
			Models:       []string{"mistral-large-4", "mistral-small-2603", "codestral-2508"},
			KeyEnv:       "MISTRAL_API_KEY", DocsURL: "https://docs.mistral.ai/models/overview",
		},
		{
			ID: "groq", Name: "Groq", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.groq.com/openai/v1", NeedsKey: true,
			TokenField:   "max_completion_tokens",
			DefaultModel: "openai/gpt-oss-120b",
			Models:       []string{"openai/gpt-oss-120b", "openai/gpt-oss-20b", "llama-3.3-70b-versatile", "llama-3.1-8b-instant"},
			KeyEnv:       "GROQ_API_KEY", DocsURL: "https://console.groq.com/docs/models",
		},
		{
			ID: "openrouter", Name: "OpenRouter", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://openrouter.ai/api/v1", NeedsKey: true,
			Headers: map[string]string{"X-OpenRouter-Title": "LiteCode"},
			KeyEnv:  "OPENROUTER_API_KEY", DocsURL: "https://openrouter.ai/docs/api-reference/overview",
			Note: "Hundreds of models behind one key; pick one from the list.",
		},
		{
			ID: "deepseek", Name: "DeepSeek", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.deepseek.com", NeedsKey: true,
			DefaultModel: "deepseek-v4-pro",
			Models:       []string{"deepseek-v4-pro", "deepseek-flash"},
			KeyEnv:       "DEEPSEEK_API_KEY", DocsURL: "https://api-docs.deepseek.com/",
		},
		{
			ID: "xai", Name: "xAI", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.x.ai/v1", NeedsKey: true,
			TokenField:   "max_completion_tokens",
			DefaultModel: "grok-4.7",
			Models:       []string{"grok-4.7", "grok-4.20-0309-reasoning", "grok-4.20-0309-non-reasoning", "grok-build-0.1"},
			KeyEnv:       "XAI_API_KEY", DocsURL: "https://docs.x.ai/docs/models",
		},
		{
			ID: "together", Name: "Together AI", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.together.ai/v1", NeedsKey: true,
			DefaultModel: "zai-org/GLM-5.3",
			Models:       []string{"zai-org/GLM-5.3", "moonshotai/Kimi-K3.1-Pro", "deepseek-ai/DeepSeek-V4-Pro-0813", "Qwen/Qwen4-Max", "openai/gpt-oss-120b"},
			KeyEnv:       "TOGETHER_API_KEY", DocsURL: "https://docs.together.ai/docs/serverless-models",
		},
		{
			ID: "fireworks", Name: "Fireworks AI", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://api.fireworks.ai/inference/v1", NeedsKey: true,
			DefaultModel: "accounts/fireworks/models/glm-5p3",
			Models:       []string{"accounts/fireworks/models/glm-5p3", "accounts/fireworks/models/minimax-m3", "accounts/fireworks/models/qwen3-coder-480b-a35b-instruct"},
			KeyEnv:       "FIREWORKS_API_KEY", DocsURL: "https://docs.fireworks.ai/tools-sdks/openai-compatibility",
		},
		{
			ID: "azure", Name: "Azure OpenAI", Kind: Cloud, Protocol: OpenAI, BaseURL: "https://YOUR-RESOURCE.openai.azure.com/openai/v1", NeedsKey: true,
			AuthHeader: "api-key", TokenField: "max_completion_tokens",
			KeyEnv: "AZURE_OPENAI_API_KEY", DocsURL: "https://learn.microsoft.com/en-us/azure/foundry/openai/api-version-lifecycle",
			Note: "Put your resource name in the URL; the model is your deployment's name.",
		},

		{
			ID: "ollama", Name: "Ollama", Kind: Local, Protocol: Ollama, BaseURL: OllamaBaseURL,
			DocsURL: "https://docs.ollama.com/api/chat", Note: "Uses the models you've pulled (ollama pull <model>).",
		},
		{
			ID: "lmstudio", Name: "LM Studio", Kind: Local, Protocol: OpenAI, BaseURL: "http://localhost:1234/v1",
			DocsURL: "https://lmstudio.ai/docs", Note: "Start the server from LM Studio's Developer tab.",
		},
		{
			ID: "llamacpp", Name: "llama.cpp server", Kind: Local, Protocol: OpenAI, BaseURL: "http://127.0.0.1:9931/v1",
			DocsURL: "https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md",
			Note:    "llama-server's default port is 9931 (older builds and Docker setups use 8080).",
		},
		{
			ID: "vllm", Name: "vLLM", Kind: Local, Protocol: OpenAI, BaseURL: "http://localhost:8000/v1",
			DocsURL: "https://docs.vllm.ai/en/latest/serving/openai_compatible_server.html",
		},
		{
			ID: "jan", Name: "Jan", Kind: Local, Protocol: OpenAI, BaseURL: "http://127.0.0.1:1337/v1",
			DocsURL: "https://docs.jan.ai/desktop/api-server", Note: "Turn on Jan's Local API Server first.",
		},
		{
			ID: "localai", Name: "LocalAI", Kind: Local, Protocol: OpenAI, BaseURL: "http://localhost:8080/v1",
			DocsURL: "https://localai.io/",
		},
	}
}
