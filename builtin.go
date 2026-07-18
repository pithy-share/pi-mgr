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