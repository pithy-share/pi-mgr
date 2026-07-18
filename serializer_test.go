package main

import (
	"encoding/json"
	"testing"
)

// Built-in with only baseUrl, no apiKey
func TestSerializeBuiltIn_OnlyBaseURL(t *testing.T) {
	scheme := &Scheme{
		ID:   "test",
		Name: "test",
		Providers: []Provider{
			{
				Key:     "openai",
				Name:    "OpenAI",
				BuiltIn: true,
				APIKey:  "",
				BaseURL: "https://proxy.example.com/v1",
				Models:  []Model{},
			},
		},
	}

	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		t.Fatalf("SerializeToModelsJSON failed: %v", err)
	}

	var output modelsJSONOutput
	json.Unmarshal(data, &output)

	prov, ok := output.Providers["openai"]
	if !ok {
		t.Fatal("expected openai provider in output")
	}
	if prov.BaseURL != "https://proxy.example.com/v1" {
		t.Errorf("expected baseUrl, got %q", prov.BaseURL)
	}
	if prov.APIKey != "" {
		t.Error("apiKey should be empty")
	}
	if prov.API != "" {
		t.Error("built-in should NOT have api field")
	}
}

// AC-12: Built-in provider with only apiKey has no models/api in output
func TestSerializeBuiltIn_OnlyAPIKey(t *testing.T) {
	scheme := &Scheme{
		ID:   "test",
		Name: "test",
		Providers: []Provider{
			{
				Key:     "openai",
				Name:    "OpenAI",
				BuiltIn: true,
				APIKey:  "sk-test123",
				BaseURL: "",
				Models:  []Model{},
			},
		},
	}

	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		t.Fatalf("SerializeToModelsJSON failed: %v", err)
	}

	var output modelsJSONOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	prov, ok := output.Providers["openai"]
	if !ok {
		t.Fatal("expected openai provider in output")
	}
	if prov.APIKey == "" {
		t.Error("expected apiKey in output")
	}
	if prov.API != "" {
		t.Error("built-in should NOT have api field (AC-12)")
	}
	if len(prov.Models) != 0 {
		t.Error("built-in with no custom models should have empty models array")
	}
}

// AC-20: Skip built-in provider with no apiKey, baseUrl, or models
func TestSerializeBuiltIn_SkipEmpty(t *testing.T) {
	scheme := &Scheme{
		ID:   "test",
		Name: "test",
		Providers: []Provider{
			{
				Key:     "anthropic",
				Name:    "Anthropic",
				BuiltIn: true,
				APIKey:  "",
				BaseURL: "",
				Models:  []Model{},
			},
		},
	}

	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		t.Fatalf("SerializeToModelsJSON failed: %v", err)
	}

	var output modelsJSONOutput
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}

	if _, ok := output.Providers["anthropic"]; ok {
		t.Error("empty built-in provider should be skipped (AC-20)")
	}
}

// AC-12: Built-in with apiKey + baseUrl → both present, no api field
func TestSerializeBuiltIn_WithBaseURL(t *testing.T) {
	scheme := &Scheme{
		ID:   "test",
		Name: "test",
		Providers: []Provider{
			{
				Key:     "openai",
				Name:    "OpenAI",
				BuiltIn: true,
				APIKey:  "sk-test",
				BaseURL: "https://proxy.example.com/v1",
				Models:  []Model{},
			},
		},
	}

	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		t.Fatalf("SerializeToModelsJSON failed: %v", err)
	}

	var output modelsJSONOutput
	json.Unmarshal(data, &output)

	prov := output.Providers["openai"]
	if prov.BaseURL != "https://proxy.example.com/v1" {
		t.Errorf("expected baseUrl, got %q", prov.BaseURL)
	}
	if prov.API != "" {
		t.Error("built-in should NOT have api field")
	}
}

// AC-21: Built-in with custom models → models appear in output
func TestSerializeBuiltIn_WithModels(t *testing.T) {
	scheme := &Scheme{
		ID:   "test",
		Name: "test",
		Providers: []Provider{
			{
				Key:     "openai",
				Name:    "OpenAI",
				BuiltIn: true,
				APIKey:  "sk-test",
				Models: []Model{
					{ID: "gpt-4-custom", Name: "GPT-4 Custom", InputText: true},
				},
			},
		},
	}

	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		t.Fatalf("SerializeToModelsJSON failed: %v", err)
	}

	var output modelsJSONOutput
	json.Unmarshal(data, &output)

	prov := output.Providers["openai"]
	if len(prov.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(prov.Models))
	}
	if prov.Models[0].ID != "gpt-4-custom" {
		t.Errorf("expected model id gpt-4-custom, got %q", prov.Models[0].ID)
	}
}

// AC-15: Custom provider must have baseUrl + api + models
func TestSerializeCustomProvider(t *testing.T) {
	scheme := &Scheme{
		ID:   "test",
		Name: "test",
		Providers: []Provider{
			{
				Key:     "my-proxy",
				Name:    "my-proxy",
				BuiltIn: false,
				APIKey:  "sk-custom",
				BaseURL: "https://my-api.example.com/v1",
				APIType: "openai-completions",
				Models: []Model{
					{ID: "my-model", InputText: true},
				},
			},
		},
	}

	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		t.Fatalf("SerializeToModelsJSON failed: %v", err)
	}

	var output modelsJSONOutput
	json.Unmarshal(data, &output)

	prov, ok := output.Providers["my-proxy"]
	if !ok {
		t.Fatal("expected custom provider in output")
	}
	if prov.BaseURL != "https://my-api.example.com/v1" {
		t.Error("custom provider must have baseUrl")
	}
	if prov.API != "openai-completions" {
		t.Error("custom provider must have api field")
	}
	if len(prov.Models) == 0 {
		t.Error("custom provider must have models array")
	}
	if prov.APIKey == "" {
		t.Error("custom provider with apiKey should have apiKey in output")
	}
}

// AC-24: Validate scheme name not empty
func TestValidateScheme_EmptyName(t *testing.T) {
	scheme := &Scheme{Name: ""}
	errs := ValidateScheme(scheme)
	if len(errs) == 0 {
		t.Error("expected error for empty scheme name (AC-24)")
	}

	scheme.Name = "   "
	errs = ValidateScheme(scheme)
	if len(errs) == 0 {
		t.Error("expected error for whitespace-only scheme name (AC-24)")
	}

	scheme.Name = "valid"
	errs = ValidateScheme(scheme)
	if len(errs) != 0 {
		t.Error("expected no error for valid scheme name")
	}
}

// AC-25: Custom provider baseUrl required
func TestValidateProvider_CustomBaseURLRequired(t *testing.T) {
	prov := &Provider{
		Key:     "custom",
		BuiltIn: false,
		BaseURL: "",
		APIType: "openai-completions",
	}
	errs := ValidateProvider(prov, nil)
	if len(errs) == 0 {
		t.Error("expected error for custom provider with empty baseUrl (AC-25)")
	}
}

// AC-26: Custom provider apiType required
func TestValidateProvider_CustomAPITypeRequired(t *testing.T) {
	prov := &Provider{
		Key:     "custom",
		BuiltIn: false,
		BaseURL: "https://example.com",
		APIType: "",
	}
	errs := ValidateProvider(prov, nil)
	if len(errs) == 0 {
		t.Error("expected error for custom provider with empty apiType (AC-26)")
	}
}

// AC-27: Model ID required
func TestValidateModel_EmptyID(t *testing.T) {
	m := &Model{ID: ""}
	errs := ValidateModel(m, nil)
	if len(errs) == 0 {
		t.Error("expected error for model with empty id (AC-27)")
	}

	m.ID = "   "
	errs = ValidateModel(m, nil)
	if len(errs) == 0 {
		t.Error("expected error for model with whitespace id (AC-27)")
	}
}

// AC-28: Model ID duplicate in same provider
func TestValidateModel_DuplicateID(t *testing.T) {
	existing := []Model{{ID: "gpt-4"}}
	m := &Model{ID: "gpt-4"}
	errs := ValidateModel(m, existing)
	if len(errs) == 0 {
		t.Error("expected error for duplicate model id (AC-28)")
	}
}

// AC-29: Custom provider key duplicate
func TestValidateProvider_DuplicateKey(t *testing.T) {
	existing := []Provider{{Key: "my-provider", BuiltIn: false}}
	prov := &Provider{Key: "my-provider", BuiltIn: false, BaseURL: "https://example.com", APIType: "openai-completions"}
	errs := ValidateProvider(prov, existing)
	if len(errs) == 0 {
		t.Error("expected error for duplicate provider key (AC-29)")
	}
}

// AC-33: Built-in provider duplicate
func TestValidateProvider_DuplicateBuiltIn(t *testing.T) {
	existing := []Provider{{Key: "openai", BuiltIn: true}}
	prov := &Provider{Key: "openai", BuiltIn: true}
	errs := ValidateProvider(prov, existing)
	if len(errs) == 0 {
		t.Error("expected error for duplicate built-in provider (AC-33)")
	}
}

// Model serialization defaults - AC-17 defaults should be omitted
func TestSerializeModel_DefaultsOmitted(t *testing.T) {
	m := Model{
		ID:            "test-model",
		Name:          "test-model", // same as ID, should be omitted
		InputText:     true,          // default, should result in omitted input
		InputImage:    false,
		ContextWindow: 128000,        // default, should be omitted
		MaxTokens:     16384,         // default, should be omitted
		Reasoning:     false,
	}

	models := serializeModels([]Model{m})
	if len(models) != 1 {
		t.Fatalf("expected 1 model")
	}
	mj := models[0]
	if mj.ID != "test-model" {
		t.Errorf("id mismatch")
	}
	if mj.Name != "" {
		t.Error("name should be omitted when same as id")
	}
	if mj.Reasoning != nil {
		t.Error("reasoning should be omitted when false")
	}
	if len(mj.Input) != 0 {
		t.Error("input should be omitted for default [\"text\"]")
	}
	if mj.ContextWindow != nil {
		t.Error("contextWindow should be omitted at default 128000")
	}
	if mj.MaxTokens != nil {
		t.Error("maxTokens should be omitted at default 16384")
	}
	if mj.Cost != nil {
		t.Error("cost should be omitted when all zero")
	}
}

// Model serialization with non-defaults
func TestSerializeModel_NonDefaults(t *testing.T) {
	m := Model{
		ID:            "gpt-4-vision",
		Name:          "GPT-4 Vision",
		InputText:     true,
		InputImage:    true,
		ContextWindow: 256000,
		MaxTokens:     4096,
		Reasoning:     true,
		CostInput:     0.01,
		CostOutput:    0.03,
	}

	models := serializeModels([]Model{m})
	mj := models[0]
	if mj.Name != "GPT-4 Vision" {
		t.Error("name should be present when different from id")
	}
	if mj.Reasoning == nil || !*mj.Reasoning {
		t.Error("reasoning should be true")
	}
	if len(mj.Input) != 2 || mj.Input[0] != "text" || mj.Input[1] != "image" {
		t.Errorf("expected [\"text\", \"image\"], got %v", mj.Input)
	}
	if mj.ContextWindow == nil || *mj.ContextWindow != 256000 {
		t.Error("contextWindow should be 256000")
	}
	if mj.MaxTokens == nil || *mj.MaxTokens != 4096 {
		t.Error("maxTokens should be 4096")
	}
	if mj.Cost == nil {
		t.Fatal("cost should be present")
	}
	if mj.Cost.Input == nil || *mj.Cost.Input != 0.01 {
		t.Error("cost.input should be 0.01")
	}
	if mj.Cost.Output == nil || *mj.Cost.Output != 0.03 {
		t.Error("cost.output should be 0.03")
	}
	if mj.Cost.CacheRead != nil {
		t.Error("cost.cacheRead should be omitted when zero")
	}
	if mj.Cost.CacheWrite != nil {
		t.Error("cost.cacheWrite should be omitted when zero")
	}
}

// Model with only image input
func TestSerializeModel_ImageOnly(t *testing.T) {
	m := Model{
		ID:         "vision-only",
		InputText:  false,
		InputImage: true,
	}
	models := serializeModels([]Model{m})
	if len(models[0].Input) != 1 || models[0].Input[0] != "image" {
		t.Errorf("expected [\"image\"], got %v", models[0].Input)
	}
}