package main

import (
	"encoding/json"
)

// modelsJSONOutput mirrors the structure written to pi's models.json
type modelsJSONOutput struct {
	Providers map[string]providerJSON `json:"providers"`
}

type providerJSON struct {
	BaseURL string       `json:"baseUrl,omitempty"`
	API     string       `json:"api,omitempty"`
	APIKey  string       `json:"apiKey,omitempty"`
	Models  []modelJSON  `json:"models,omitempty"`
}

type modelJSON struct {
	ID            string   `json:"id"`
	Name          string   `json:"name,omitempty"`
	Reasoning     *bool    `json:"reasoning,omitempty"`
	Input         []string `json:"input,omitempty"`
	ContextWindow *int     `json:"contextWindow,omitempty"`
	MaxTokens     *int     `json:"maxTokens,omitempty"`
	Cost          *costJSON `json:"cost,omitempty"`
}

type costJSON struct {
	Input      *float64 `json:"input,omitempty"`
	Output     *float64 `json:"output,omitempty"`
	CacheRead  *float64 `json:"cacheRead,omitempty"`
	CacheWrite *float64 `json:"cacheWrite,omitempty"`
}

// SerializeToModelsJSON converts providers to pi's models.json format. Only enabled providers are included.
func SerializeToModelsJSON(providers []Provider) ([]byte, error) {
	output := modelsJSONOutput{
		Providers: make(map[string]providerJSON),
	}

	for _, prov := range providers {
		// Skip disabled providers (AC-03, AC-04)
		if !prov.Enabled {
			continue
		}
		if prov.BuiltIn {
			// Skip built-in providers with no overrides (AC-20)
			if prov.APIKey == "" && prov.BaseURL == "" && len(prov.Models) == 0 {
				continue
			}
			// Built-in: only apiKey, baseUrl, models. NEVER api (AC-12)
			pj := providerJSON{}
			if prov.APIKey != "" {
				pj.APIKey = prov.APIKey
			}
			if prov.BaseURL != "" {
				pj.BaseURL = prov.BaseURL
			}
			if len(prov.Models) > 0 {
				pj.Models = serializeModels(prov.Models)
			}
			output.Providers[prov.Key] = pj
		} else {
			// Custom: must have baseUrl, api, models; optional apiKey (AC-15)
			pj := providerJSON{
				BaseURL: prov.BaseURL,
				API:     prov.APIType,
				Models:  serializeModels(prov.Models),
			}
			if prov.APIKey != "" {
				pj.APIKey = prov.APIKey
			}
			output.Providers[prov.Key] = pj
		}
	}

	return json.MarshalIndent(output, "", "  ")
}

func serializeModels(models []Model) []modelJSON {
	result := make([]modelJSON, 0, len(models))
	for _, m := range models {
		mj := modelJSON{
			ID: m.ID,
		}

		// name: only if different from id
		if m.Name != "" && m.Name != m.ID {
			mj.Name = m.Name
		}

		// reasoning: only if true
		if m.Reasoning {
			t := true
			mj.Reasoning = &t
		}

		// input: omit if default ["text"]
		input := buildInput(m.InputText, m.InputImage)
		if len(input) > 0 {
			mj.Input = input
		}

		// contextWindow: omit only if zero (default values still write)
		if m.ContextWindow != 0 {
			cw := m.ContextWindow
			mj.ContextWindow = &cw
		}

		// maxTokens: omit only if zero (default values still write)
		if m.MaxTokens != 0 {
			mt := m.MaxTokens
			mj.MaxTokens = &mt
		}

		// cost: omit if all zero
		cost := buildCost(m)
		if cost != nil {
			mj.Cost = cost
		}

		result = append(result, mj)
	}
	return result
}

func buildInput(inputText, inputImage bool) []string {
	var input []string
	// text is always present in the UI but the default is ["text"]
	// If both text and image, emit ["text", "image"]
	// If only text (default), omit
	if inputImage {
		input = append(input, "text")
		input = append(input, "image")
	}
	// If only text and no image, input stays nil (omitted)
	// But if text is false and image is true, still emit ["image"]
	if !inputText && inputImage {
		return []string{"image"}
	}
	return input
}

func buildCost(m Model) *costJSON {
	if m.CostInput == 0 && m.CostOutput == 0 && m.CostCacheRead == 0 && m.CostCacheWrite == 0 {
		return nil
	}
	cj := &costJSON{}
	if m.CostInput != 0 {
		cj.Input = &m.CostInput
	}
	if m.CostOutput != 0 {
		cj.Output = &m.CostOutput
	}
	if m.CostCacheRead != 0 {
		cj.CacheRead = &m.CostCacheRead
	}
	if m.CostCacheWrite != 0 {
		cj.CacheWrite = &m.CostCacheWrite
	}
	return cj
}