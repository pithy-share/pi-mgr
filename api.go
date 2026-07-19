package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// Provider management
// =============================================================================

// AddBuiltInProvider adds a built-in provider to the config
func (a *App) AddBuiltInProvider(providerKey, apiKey, baseURL string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Check if provider key already exists
	for _, p := range cfg.Providers {
		if p.Key == providerKey {
			return fmt.Errorf("该内置供应商已添加")
		}
	}

	// Find built-in catalog entry
	var displayName string
	for _, b := range BuiltInProviders {
		if b.Key == providerKey {
			displayName = b.Name
			break
		}
	}
	if displayName == "" {
		return fmt.Errorf("unknown built-in provider: %s", providerKey)
	}

	prov := Provider{
		Key:     providerKey,
		Name:    displayName,
		BuiltIn: true,
		Enabled: true,
		APIKey:  strings.TrimSpace(apiKey),
		BaseURL: strings.TrimSpace(baseURL),
		Models:  []Model{},
	}

	if errs := ValidateProvider(&prov, cfg.Providers); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	cfg.Providers = append(cfg.Providers, prov)
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// AddCustomProvider adds a custom provider to the config
func (a *App) AddCustomProvider(key, baseURL, apiType, apiKey string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	prov := Provider{
		Key:     strings.TrimSpace(key),
		Name:    strings.TrimSpace(key),
		BuiltIn: false,
		Enabled: true,
		APIKey:  strings.TrimSpace(apiKey),
		BaseURL: strings.TrimSpace(baseURL),
		APIType: apiType,
		Models:  []Model{},
	}

	if errs := ValidateProvider(&prov, cfg.Providers); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	cfg.Providers = append(cfg.Providers, prov)
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// UpdateProvider updates an existing provider's configuration
func (a *App) UpdateProvider(provider Provider) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, provider.Key)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", provider.Key)
	}

	// Build list of other providers (excluding self) for uniqueness check
	others := make([]Provider, 0, len(cfg.Providers)-1)
	for i, p := range cfg.Providers {
		if i != pidx {
			others = append(others, p)
		}
	}

	// Preserve models that might not be sent from frontend
	cfg.Providers[pidx].APIKey = provider.APIKey
	cfg.Providers[pidx].BaseURL = provider.BaseURL
	if !provider.BuiltIn {
		cfg.Providers[pidx].APIType = provider.APIType
	}

	updated := cfg.Providers[pidx]
	if errs := ValidateProvider(&updated, others); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// SetProviderEnabled enables or disables a provider and syncs models.json immediately
func (a *App) SetProviderEnabled(providerKey string, enabled bool) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	cfg.Providers[pidx].Enabled = enabled
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// RemoveProvider removes a provider from the config
func (a *App) RemoveProvider(providerKey string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	cfg.Providers = append(cfg.Providers[:pidx], cfg.Providers[pidx+1:]...)
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// =============================================================================
// Model management
// =============================================================================

// AddModel adds a model to a provider
func (a *App) AddModel(providerKey string, model Model) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &cfg.Providers[pidx]

	if errs := ValidateModel(&model, prov.Models); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	prov.Models = append(prov.Models, model)
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// UpdateModel updates an existing model in a provider
func (a *App) UpdateModel(providerKey string, model Model) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &cfg.Providers[pidx]
	midx := findModelIndex(prov, model.ID)
	if midx < 0 {
		return fmt.Errorf("model not found: %s", model.ID)
	}

	// Build other models (excluding self) for uniqueness check
	others := make([]Model, 0, len(prov.Models)-1)
	for i, m := range prov.Models {
		if i != midx {
			others = append(others, m)
		}
	}

	if errs := ValidateModel(&model, others); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	prov.Models[midx] = model
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// RemoveModel removes a model from a provider
func (a *App) RemoveModel(providerKey, modelID string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &cfg.Providers[pidx]
	midx := findModelIndex(prov, modelID)
	if midx < 0 {
		return fmt.Errorf("model not found: %s", modelID)
	}

	prov.Models = append(prov.Models[:midx], prov.Models[midx+1:]...)
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// ImportProviderModels bulk-imports models into a provider, skipping duplicates by ID.
// Returns the number of models actually added (0 if all were skipped).
func (a *App) ImportProviderModels(providerKey string, models []Model) (int, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return 0, err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return 0, fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &cfg.Providers[pidx]

	// Build existing ID set for O(1) duplicate check
	existing := make(map[string]bool, len(prov.Models))
	for _, m := range prov.Models {
		existing[m.ID] = true
	}

	added := 0
	for _, m := range models {
		if existing[m.ID] {
			continue
		}
		prov.Models = append(prov.Models, m)
		existing[m.ID] = true
		added++
	}

	if added == 0 {
		return 0, nil
	}

	if err := SaveConfig(cfg); err != nil {
		return 0, err
	}
	if err := syncModelsJSON(cfg); err != nil {
		return 0, err
	}
	return added, nil
}

// =============================================================================
// Batch operations
// =============================================================================

// RemoveModels removes multiple models from a provider. Best-effort: skips
// models that don't exist, returns count of successfully removed models.
func (a *App) RemoveModels(providerKey string, modelIDs []string) (int, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return 0, err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return 0, fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &cfg.Providers[pidx]

	// Build set for O(1) lookup
	toRemove := make(map[string]bool, len(modelIDs))
	for _, id := range modelIDs {
		toRemove[id] = true
	}

	// Filter out removed models
	kept := make([]Model, 0, len(prov.Models))
	for _, m := range prov.Models {
		if !toRemove[m.ID] {
			kept = append(kept, m)
		}
	}

	removed := len(prov.Models) - len(kept)
	if removed == 0 {
		return 0, nil
	}

	prov.Models = kept
	if err := SaveConfig(cfg); err != nil {
		return 0, err
	}
	if err := syncModelsJSON(cfg); err != nil {
		return 0, err
	}
	return removed, nil
}

// ReorderProviders reorders providers in the config by the given key order.
func (a *App) ReorderProviders(orderedKeys []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Validate: orderedKeys must match existing provider keys exactly
	if len(orderedKeys) != len(cfg.Providers) {
		return fmt.Errorf("供应商列表不一致")
	}

	// Build index map for reordering
	indexByKey := make(map[string]int, len(cfg.Providers))
	for i, p := range cfg.Providers {
		indexByKey[p.Key] = i
	}

	// Check every ordered key exists, and check for duplicates
	seen := make(map[string]bool, len(orderedKeys))
	reordered := make([]Provider, 0, len(cfg.Providers))
	for _, key := range orderedKeys {
		pos, ok := indexByKey[key]
		if !ok {
			return fmt.Errorf("供应商列表不一致")
		}
		if seen[key] {
			return fmt.Errorf("供应商列表不一致")
		}
		seen[key] = true
		reordered = append(reordered, cfg.Providers[pos])
	}

	cfg.Providers = reordered
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// ReorderModels reorders models in a provider.
func (a *App) ReorderModels(providerKey string, orderedIDs []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &cfg.Providers[pidx]

	// Validate: orderedIDs must match existing model IDs exactly
	if len(orderedIDs) != len(prov.Models) {
		return fmt.Errorf("模型列表不一致")
	}

	// Build index map for reordering
	indexByID := make(map[string]int, len(prov.Models))
	for i, m := range prov.Models {
		indexByID[m.ID] = i
	}

	seen := make(map[string]bool, len(orderedIDs))
	reordered := make([]Model, 0, len(prov.Models))
	for _, id := range orderedIDs {
		pos, ok := indexByID[id]
		if !ok {
			return fmt.Errorf("模型列表不一致")
		}
		if seen[id] {
			return fmt.Errorf("模型列表不一致")
		}
		seen[id] = true
		reordered = append(reordered, prov.Models[pos])
	}

	prov.Models = reordered
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	return syncModelsJSON(cfg)
}

// =============================================================================
// Connectivity test
// =============================================================================

// TestProviderConnectivity tests connectivity to a provider's API endpoint.
func (a *App) TestProviderConnectivity(providerKey string) (string, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	pidx := findProviderIndex(cfg, providerKey)
	if pidx < 0 {
		return "", fmt.Errorf("provider not found: %s", providerKey)
	}
	prov := cfg.Providers[pidx]

		// For built-in providers with empty BaseURL, use the official default endpoint
	baseURL := strings.TrimSpace(prov.BaseURL)
	var url string
	if baseURL == "" {
		if prov.BuiltIn {
			modelURL := defaultModelListURL(prov.Key)
			if modelURL == "" {
				return "", fmt.Errorf("该内置供应商暂不支持空 Base URL 测试，请手动配置")
			}
			url = modelURL
		} else {
			return "", fmt.Errorf("请先配置 Base URL")
		}
	} else {
		baseURL = strings.TrimRight(baseURL, "/")
		url = baseURL + "/v1/models"
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if prov.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+prov.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "无法连接，请检查 Base URL 和网络", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "连接成功，API 可达", nil
	}

	return fmt.Sprintf("API 返回错误（状态码 %d），请检查 API Key", resp.StatusCode), nil
}

// =============================================================================
// Export / Import
// =============================================================================

// ExportConfig exports the current config to a user-chosen file
func (a *App) ExportConfig() error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "pi-mgr-config.json",
		Title:           "导出配置",
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON 文件 (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		return fmt.Errorf("打开保存对话框失败: %w", err)
	}
	if path == "" {
		return nil // user cancelled
	}

	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("写入导出文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		if err2 := os.WriteFile(path, data, 0644); err2 != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("保存导出文件失败: %w", err2)
		}
		os.Remove(tmpPath)
	}

	return nil
}

// ImportConfig imports a config from a user-chosen JSON file, replacing the current config
func (a *App) ImportConfig() error {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入配置",
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON 文件 (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		return fmt.Errorf("打开文件对话框失败: %w", err)
	}
	if path == "" {
		return nil // user cancelled
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "[]" {
		return nil // empty → no-op
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return fmt.Errorf("JSON 格式错误: %w", err)
	}

	var imported Config

	if len(raw) > 0 && raw[0] == '{' {
		// Try parsing as Config directly
		if err := json.Unmarshal(raw, &imported); err != nil {
			return fmt.Errorf("JSON 格式错误: %w", err)
		}
		// If parsed as Config but looks like a legacy Scheme (has "id" field), extract providers
		if imported.Providers == nil {
			// Try legacy: parse as Scheme and extract providers
			var legacy struct {
				ID        string     `json:"id"`
				Name      string     `json:"name"`
				Providers []Provider `json:"providers"`
			}
			if err := json.Unmarshal(raw, &legacy); err != nil {
				return fmt.Errorf("JSON 格式错误: %w", err)
			}
			if legacy.Providers != nil {
				imported.Providers = legacy.Providers
			}
		}
	} else if len(raw) > 0 && raw[0] == '[' {
		// Legacy: array of schemes → take first scheme's providers
		var schemes []struct {
			ID        string     `json:"id"`
			Name      string     `json:"name"`
			Providers []Provider `json:"providers"`
		}
		if err := json.Unmarshal(raw, &schemes); err != nil {
			return fmt.Errorf("JSON 格式错误: %w", err)
		}
		if len(schemes) > 0 {
			imported.Providers = schemes[0].Providers
		}
	} else {
		return fmt.Errorf("JSON 格式错误: 文件根层级应为对象或数组")
	}

	if len(imported.Providers) == 0 {
		// Import empty providers → overwrite with empty config
		if err := SaveConfig(&Config{}); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
		return syncModelsJSON(&Config{})
	}

	// Downgrade unknown builtIn providers
	for pi := range imported.Providers {
		prov := &imported.Providers[pi]
		if prov.BuiltIn && !IsBuiltInProvider(prov.Key) {
			prov.BuiltIn = false
		}
	}

	// Full validation
	for pi := range imported.Providers {
		prov := &imported.Providers[pi]
		others := make([]Provider, 0, len(imported.Providers)-1)
		for qi, other := range imported.Providers {
			if qi != pi {
				others = append(others, other)
			}
		}
		if errs := ValidateProvider(prov, others); len(errs) > 0 {
			return fmt.Errorf("供应商 \"%s\" 校验失败: %s", prov.Key, errs[0])
		}
		for mi := range prov.Models {
			model := &prov.Models[mi]
			omodels := make([]Model, 0, len(prov.Models)-1)
			for ni, other := range prov.Models {
				if ni != mi {
					omodels = append(omodels, other)
				}
			}
			if errs := ValidateModel(model, omodels); len(errs) > 0 {
				return fmt.Errorf("供应商 \"%s\" 模型 \"%s\" 校验失败: %s", prov.Key, model.ID, errs[0])
			}
		}
	}

	if err := SaveConfig(&imported); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	return syncModelsJSON(&imported)
}

// =============================================================================
// Config query
// =============================================================================

// GetConfig returns the current config
func (a *App) GetConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		return &Config{}
	}
	return cfg
}

// =============================================================================
// Catalog queries
// =============================================================================

// ListBuiltInProviders returns the built-in provider catalog
func (a *App) ListBuiltInProviders() []BuiltInProvider {
	return BuiltInProviders
}

// ListAPITypes returns valid API types for custom providers
func (a *App) ListAPITypes() []string {
	return ValidAPITypes
}

// =============================================================================
// Helpers
// =============================================================================

func findModelIndex(prov *Provider, id string) int {
	for i := range prov.Models {
		if prov.Models[i].ID == id {
			return i
		}
	}
	return -1
}