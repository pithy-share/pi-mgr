package main

// BuiltInProviders is the hardcoded list of pi built-in providers.
// Source: https://pi.dev/docs/latest/providers
var BuiltInProviders = []BuiltInProvider{
	{Key: "openai", Name: "OpenAI", APIType: "openai-completions"},
	{Key: "anthropic", Name: "Anthropic", APIType: "anthropic-messages"},
	{Key: "deepseek", Name: "DeepSeek", APIType: "openai-completions"},
	{Key: "google", Name: "Google Gemini", APIType: "google-generative-ai"},
	{Key: "mistral", Name: "Mistral", APIType: "mistral-conversations"},
	{Key: "groq", Name: "Groq", APIType: "openai-completions"},
	{Key: "xai", Name: "xAI", APIType: "openai-completions"},
	{Key: "openrouter", Name: "OpenRouter", APIType: "openai-completions"},
	{Key: "together", Name: "Together AI", APIType: "openai-completions"},
	{Key: "fireworks", Name: "Fireworks", APIType: "openai-completions"},
	{Key: "cerebras", Name: "Cerebras", APIType: "openai-completions"},
	{Key: "bedrock", Name: "Amazon Bedrock", APIType: "bedrock-converse-stream"},
	{Key: "nvidia", Name: "NVIDIA NIM", APIType: "openai-completions"},
	{Key: "huggingface", Name: "Hugging Face", APIType: "openai-completions"},
	{Key: "ant-ling", Name: "Ant Ling", APIType: "openai-completions"},
	{Key: "vercel-ai-gateway", Name: "Vercel AI Gateway", APIType: "openai-completions"},
	{Key: "zai-coding-plan-global", Name: "ZAI Coding Plan (Global)", APIType: "openai-completions"},
	{Key: "zai-coding-plan-china", Name: "ZAI Coding Plan (China)", APIType: "openai-completions"},
	{Key: "opencode-zen", Name: "OpenCode Zen", APIType: "openai-completions"},
	{Key: "opencode-go", Name: "OpenCode Go", APIType: "openai-completions"},
	{Key: "kimi-for-coding", Name: "Kimi For Coding", APIType: "openai-completions"},
	{Key: "minimax", Name: "MiniMax", APIType: "openai-completions"},
	{Key: "minimax-china", Name: "MiniMax (China)", APIType: "openai-completions"},
	{Key: "xiaomi-mimo", Name: "Xiaomi MiMo", APIType: "openai-completions"},
	{Key: "xiaomi-mimo-china", Name: "Xiaomi MiMo Token Plan (China)", APIType: "openai-completions"},
	{Key: "xiaomi-mimo-amsterdam", Name: "Xiaomi MiMo Token Plan (Amsterdam)", APIType: "openai-completions"},
	{Key: "xiaomi-mimo-singapore", Name: "Xiaomi MiMo Token Plan (Singapore)", APIType: "openai-completions"},
}

// IsBuiltInProvider checks if a key is in the built-in provider catalog
func IsBuiltInProvider(key string) bool {
	for _, bp := range BuiltInProviders {
		if bp.Key == key {
			return true
		}
	}
	return false
}

// ValidAPITypes is the list of API types available for custom providers.
// Source: https://pi.dev/docs/latest/custom-provider
var ValidAPITypes = []string{
	"openai-completions",
	"anthropic-messages",
	"openai-responses",
	"azure-openai-responses",
	"openai-codex-responses",
	"mistral-conversations",
	"google-generative-ai",
	"google-vertex",
	"bedrock-converse-stream",
}

// defaultBaseURL returns the official API base URL for a known built-in provider.
// Source: pi's built-in provider definitions in @earendil-works/pi-ai/dist/providers/*.js
// Returns empty string if the provider has no known default (user must configure manually).
func defaultBaseURL(key string) string {
	urls := map[string]string{
		"openai":               "https://api.openai.com",
		"anthropic":            "https://api.anthropic.com",
		"deepseek":             "https://api.deepseek.com",
		"google":               "https://generativelanguage.googleapis.com",
		"mistral":              "https://api.mistral.ai",
		"groq":                 "https://api.groq.com",
		"xai":                  "https://api.x.ai",
		"openrouter":           "https://openrouter.ai",
		"together":             "https://api.together.ai",
		"fireworks":            "https://api.fireworks.ai",
		"cerebras":             "https://api.cerebras.ai",
		"nvidia":               "https://integrate.api.nvidia.com",
		"huggingface":          "https://router.huggingface.co",
		"vercel-ai-gateway":    "https://ai-gateway.vercel.sh",
		"zai-coding-plan-global": "https://api.z.ai",
		"zai-coding-plan-china":  "https://open.bigmodel.cn",
		"kimi-for-coding":      "https://api.kimi.com",
		"minimax":              "https://api.minimax.io",
		"minimax-china":        "https://api.minimaxi.com",
		"xiaomi-mimo":          "https://api.xiaomimimo.com",
		"xiaomi-mimo-china":   "https://token-plan-cn.xiaomimimo.com",
		"xiaomi-mimo-amsterdam": "https://token-plan-ams.xiaomimimo.com",
		"xiaomi-mimo-singapore": "https://token-plan-sgp.xiaomimimo.com",
		"ant-ling":             "https://api.ant-ling.com",
	}
	return urls[key]
}

// defaultModelListURL returns the complete model listing URL for a known built-in provider.
// Returns empty string if the provider has no known default (user must configure manually).
func defaultModelListURL(key string) string {
	// For OpenAI-compatible providers, model list is at GET {base}/models
	// where {base} includes the version prefix (e.g., https://api.openai.com/v1).
	// For providers where the model list path differs, it's defined here.
	urls := map[string]string{
		"openai":               "https://api.openai.com/v1/models",
		"anthropic":            "https://api.anthropic.com/v1/models",
		"deepseek":             "https://api.deepseek.com/models",
		"google":               "https://generativelanguage.googleapis.com/v1beta/models",
		"mistral":              "https://api.mistral.ai/v1/models",
		"groq":                 "https://api.groq.com/openai/v1/models",
		"xai":                  "https://api.x.ai/v1/models",
		"openrouter":           "https://openrouter.ai/api/v1/models",
		"together":             "https://api.together.ai/v1/models",
		"fireworks":            "https://api.fireworks.ai/inference/v1/models",
		"cerebras":             "https://api.cerebras.ai/v1/models",
		"nvidia":               "https://integrate.api.nvidia.com/v1/models",
		"huggingface":          "https://router.huggingface.co/v1/models",
		"vercel-ai-gateway":    "https://ai-gateway.vercel.sh/v1/models",
		"zai-coding-plan-global": "https://api.z.ai/api/coding/paas/v4/models",
		"zai-coding-plan-china":  "https://open.bigmodel.cn/api/coding/paas/v4/models",
		"kimi-for-coding":      "https://api.kimi.com/coding/models",
		"minimax":              "https://api.minimax.io/anthropic/v1/models",
		"minimax-china":        "https://api.minimaxi.com/anthropic/v1/models",
		"xiaomi-mimo":          "https://api.xiaomimimo.com/v1/models",
		"xiaomi-mimo-china":   "https://token-plan-cn.xiaomimimo.com/v1/models",
		"xiaomi-mimo-amsterdam": "https://token-plan-ams.xiaomimimo.com/v1/models",
		"xiaomi-mimo-singapore": "https://token-plan-sgp.xiaomimimo.com/v1/models",
		"ant-ling":             "https://api.ant-ling.com/v1/models",
	}
	return urls[key]
}