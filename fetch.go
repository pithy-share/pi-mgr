package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// FetchProviderModels calls GET {baseURL}/v1/models and parses OpenAI-compatible
// List Models response: {"object": "list", "data": [{"id": "...", ...}]}
// Returns []Model with only id and name populated (rest are zero/default values).
func (a *App) FetchProviderModels(providerKey string) ([]Model, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return nil, fmt.Errorf("provider not found: %s", providerKey)
	}
	prov := cfg.Providers[pidx]

	var url string
	baseURL := strings.TrimSpace(prov.BaseURL)
	if baseURL == "" && prov.BuiltIn {
		modelURL := defaultModelListURL(prov.Key)
		if modelURL == "" {
			return nil, fmt.Errorf("该供应商没有默认 API 地址，请先配置 Base URL")
		}
		url = modelURL
	} else {
		baseURL = strings.TrimRight(baseURL, "/")
		url = baseURL + "/v1/models"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("网络不可达: %w", err)
	}

	// AC-15: no Authorization header when apiKey is empty
	if prov.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+prov.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络不可达: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try to read a small portion of the response body for the error message
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		bodyMsg := strings.TrimSpace(string(bodyBytes))
		if bodyMsg != "" {
			return nil, fmt.Errorf("API 返回状态码 %d: %s", resp.StatusCode, bodyMsg)
		}
		return nil, fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	var listResp struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("响应解析失败: %w", err)
	}

	// AC-09: response missing data field or data is not an array
	if listResp.Data == nil {
		return nil, fmt.Errorf("响应解析失败: 响应中缺少 data 字段")
	}

	// Map to []Model: only id and name filled; rest are default values
	// Defaults match defaultModel() in types.ts:
	//   Reasoning=false, InputText=true, InputImage=false,
	//   ContextWindow=256000, MaxTokens=64000, Cost*=0
	models := make([]Model, 0, len(listResp.Data))
	for _, d := range listResp.Data {
		m := Model{
			ID:             d.ID,
			Name:           d.Name,
			Reasoning:      false,
			InputText:      true,
			InputImage:     false,
			ContextWindow:  256000,
			MaxTokens:      64000,
			CostInput:      0,
			CostOutput:     0,
			CostCacheRead:  0,
			CostCacheWrite: 0,
		}
		// Name: prefer response name, fall back to id
		if m.Name == "" {
			m.Name = m.ID
		}
		models = append(models, m)
	}

	return models, nil
}
