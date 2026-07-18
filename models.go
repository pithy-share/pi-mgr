package main

// Scheme is a named collection of providers
type Scheme struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Providers []Provider `json:"providers"`
}

// Provider represents either a built-in or custom provider config
type Provider struct {
	Key     string  `json:"key"`
	Name    string  `json:"name"`
	BuiltIn bool    `json:"builtIn"`
	APIKey  string  `json:"apiKey"`
	BaseURL string  `json:"baseUrl"`
	APIType string  `json:"apiType"`
	Models  []Model `json:"models"`
}

// Model represents a model configuration
type Model struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Reasoning      bool    `json:"reasoning"`
	InputText      bool    `json:"inputText"`
	InputImage     bool    `json:"inputImage"`
	ContextWindow  int     `json:"contextWindow"`
	MaxTokens      int     `json:"maxTokens"`
	CostInput      float64 `json:"costInput"`
	CostOutput     float64 `json:"costOutput"`
	CostCacheRead  float64 `json:"costCacheRead"`
	CostCacheWrite float64 `json:"costCacheWrite"`
}

// BuiltInProvider is a built-in provider catalog entry
type BuiltInProvider struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	APIType string `json:"apiType"`
}