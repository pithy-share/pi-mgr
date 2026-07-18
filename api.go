package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// =============================================================================
// Scheme CRUD
// =============================================================================

// ListSchemes returns all saved schemes
func (a *App) ListSchemes() []Scheme {
	schemes, err := LoadSchemes()
	if err != nil {
		return []Scheme{}
	}
	return schemes
}

// CreateScheme creates a new scheme with the given name
func (a *App) CreateScheme(name string) (*Scheme, error) {
	name = strings.TrimSpace(name)
	scheme := &Scheme{
		ID:        newUUID(),
		Name:      name,
		Providers: []Provider{},
	}

	if errs := ValidateScheme(scheme); len(errs) > 0 {
		return nil, fmt.Errorf("%s", errs[0])
	}

	schemes, err := LoadSchemes()
	if err != nil {
		return nil, err
	}
	schemes = append(schemes, *scheme)
	if err := SaveSchemes(schemes); err != nil {
		return nil, err
	}
	return scheme, nil
}

// UpdateScheme updates a scheme's name
func (a *App) UpdateScheme(scheme Scheme) error {
	if errs := ValidateScheme(&scheme); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	for i := range schemes {
		if schemes[i].ID == scheme.ID {
			schemes[i].Name = scheme.Name
			return SaveSchemes(schemes)
		}
	}
	return fmt.Errorf("scheme not found: %s", scheme.ID)
}

// DeleteScheme removes a scheme by ID
func (a *App) DeleteScheme(id string) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	for i := range schemes {
		if schemes[i].ID == id {
			schemes = append(schemes[:i], schemes[i+1:]...)
			return SaveSchemes(schemes)
		}
	}
	return fmt.Errorf("scheme not found: %s", id)
}

// DuplicateScheme creates a copy of a scheme with " - 副本" suffix
func (a *App) DuplicateScheme(id string) (*Scheme, error) {
	schemes, err := LoadSchemes()
	if err != nil {
		return nil, err
	}
	for i := range schemes {
		if schemes[i].ID == id {
			newScheme := schemes[i]
			newScheme.ID = newUUID()
			newScheme.Name = newScheme.Name + " - 副本"
			// Deep copy providers
			newScheme.Providers = make([]Provider, len(schemes[i].Providers))
			copy(newScheme.Providers, schemes[i].Providers)
			for j := range newScheme.Providers {
				newScheme.Providers[j].Models = make([]Model, len(schemes[i].Providers[j].Models))
				copy(newScheme.Providers[j].Models, schemes[i].Providers[j].Models)
			}

			schemes = append(schemes, newScheme)
			if err := SaveSchemes(schemes); err != nil {
				return nil, err
			}
			return &newScheme, nil
		}
	}
	return nil, fmt.Errorf("scheme not found: %s", id)
}

// ActivateScheme writes the scheme to pi's models.json
func (a *App) ActivateScheme(id string) error {
	scheme, err := GetScheme(id)
	if err != nil {
		return err
	}
	return ActivateScheme(scheme)
}

// =============================================================================
// Provider management
// =============================================================================

// AddBuiltInProvider adds a built-in provider to a scheme
func (a *App) AddBuiltInProvider(schemeID, providerKey, apiKey, baseURL string) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]

	// Check if provider key already exists in this scheme (AC-33)
	for _, p := range scheme.Providers {
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
		APIKey:  strings.TrimSpace(apiKey),
		BaseURL: strings.TrimSpace(baseURL),
		Models:  []Model{},
	}

	// Validate
	if errs := ValidateProvider(&prov, scheme.Providers); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	scheme.Providers = append(scheme.Providers, prov)
	return SaveSchemes(schemes)
}

// AddCustomProvider adds a custom provider to a scheme
func (a *App) AddCustomProvider(schemeID, key, baseURL, apiType, apiKey string) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]

	prov := Provider{
		Key:     strings.TrimSpace(key),
		Name:    strings.TrimSpace(key),
		BuiltIn: false,
		APIKey:  strings.TrimSpace(apiKey),
		BaseURL: strings.TrimSpace(baseURL),
		APIType: apiType,
		Models:  []Model{},
	}

	if errs := ValidateProvider(&prov, scheme.Providers); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	scheme.Providers = append(scheme.Providers, prov)
	return SaveSchemes(schemes)
}

// UpdateProvider updates an existing provider's configuration
func (a *App) UpdateProvider(schemeID string, provider Provider) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]
	pidx := findProviderIndex(scheme, provider.Key)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", provider.Key)
	}

	// Build list of other providers (excluding self) for uniqueness check
	others := make([]Provider, 0, len(scheme.Providers)-1)
	for i, p := range scheme.Providers {
		if i != pidx {
			others = append(others, p)
		}
	}

	// Preserve models that might not be sent from frontend
	scheme.Providers[pidx].APIKey = provider.APIKey
	scheme.Providers[pidx].BaseURL = provider.BaseURL
	if !provider.BuiltIn {
		scheme.Providers[pidx].APIType = provider.APIType
	}

	// Validate with others
	updated := scheme.Providers[pidx]
	if errs := ValidateProvider(&updated, others); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	return SaveSchemes(schemes)
}

// RemoveProvider removes a provider from a scheme
func (a *App) RemoveProvider(schemeID, providerKey string) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]
	pidx := findProviderIndex(scheme, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	scheme.Providers = append(scheme.Providers[:pidx], scheme.Providers[pidx+1:]...)
	return SaveSchemes(schemes)
}

// =============================================================================
// Model management
// =============================================================================

// AddModel adds a model to a provider within a scheme
func (a *App) AddModel(schemeID, providerKey string, model Model) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]
	pidx := findProviderIndex(scheme, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &scheme.Providers[pidx]

	if errs := ValidateModel(&model, prov.Models); len(errs) > 0 {
		return fmt.Errorf("%s", errs[0])
	}

	prov.Models = append(prov.Models, model)
	return SaveSchemes(schemes)
}

// UpdateModel updates an existing model in a provider
func (a *App) UpdateModel(schemeID, providerKey string, model Model) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]
	pidx := findProviderIndex(scheme, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &scheme.Providers[pidx]
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
	return SaveSchemes(schemes)
}

// RemoveModel removes a model from a provider
func (a *App) RemoveModel(schemeID, providerKey, modelID string) error {
	schemes, err := LoadSchemes()
	if err != nil {
		return err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]
	pidx := findProviderIndex(scheme, providerKey)
	if pidx < 0 {
		return fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &scheme.Providers[pidx]
	midx := findModelIndex(prov, modelID)
	if midx < 0 {
		return fmt.Errorf("model not found: %s", modelID)
	}

	prov.Models = append(prov.Models[:midx], prov.Models[midx+1:]...)
	return SaveSchemes(schemes)
}

// ImportProviderModels bulk-imports models into a provider, skipping duplicates by ID.
// Returns the number of models actually added (0 if all were skipped).
func (a *App) ImportProviderModels(schemeID, providerKey string, models []Model) (int, error) {
	schemes, err := LoadSchemes()
	if err != nil {
		return 0, err
	}
	idx := findSchemeIndex(schemes, schemeID)
	if idx < 0 {
		return 0, fmt.Errorf("scheme not found: %s", schemeID)
	}

	scheme := &schemes[idx]
	pidx := findProviderIndex(scheme, providerKey)
	if pidx < 0 {
		return 0, fmt.Errorf("provider not found: %s", providerKey)
	}

	prov := &scheme.Providers[pidx]

	// Build existing ID set for O(1) duplicate check (AC-10)
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

	// AC-11: no error when all skipped — just return 0
	if added == 0 {
		return 0, nil
	}

	if err := SaveSchemes(schemes); err != nil {
		return 0, err
	}
	return added, nil
}

// =============================================================================
// Export / Import
// =============================================================================

// ExportSchemes exports all schemes to a user-chosen file (AC-01)
func (a *App) ExportSchemes() error {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: "pi-mgr-schemes.json",
		Title:           "导出配置方案",
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON 文件 (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		return fmt.Errorf("打开保存对话框失败: %w", err)
	}
	if path == "" {
		// User cancelled
		return nil
	}

	schemes, err := LoadSchemes()
	if err != nil {
		return fmt.Errorf("读取方案失败: %w", err)
	}

	data, err := json.MarshalIndent(schemes, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化方案失败: %w", err)
	}

	// Atomic write: temp + rename, fallback to direct write (same pattern as SaveSchemes)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("写入导出文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		// Rename may fail across volumes; fallback to direct write
		if err2 := os.WriteFile(path, data, 0644); err2 != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("保存导出文件失败: %w", err2)
		}
		os.Remove(tmpPath)
	}

	return nil
}

// ImportSchemes imports schemes from a user-chosen JSON file (AC-04, AC-05, AC-06, AC-07, AC-09, AC-10)
func (a *App) ImportSchemes() error {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入配置方案",
		Filters: []runtime.FileFilter{{
			DisplayName: "JSON 文件 (*.json)",
			Pattern:     "*.json",
		}},
	})
	if err != nil {
		return fmt.Errorf("打开文件对话框失败: %w", err)
	}
	if path == "" {
		// User cancelled
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	// Trim whitespace for empty/whitespace-only detection
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "[]" {
		// Empty array → no-op (AC-09)
		return nil
	}

	// Parse as raw message to determine top-level type (AC-04: object or array)
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return fmt.Errorf("JSON 格式错误: %w", err)
	}

	var imported []Scheme
	if len(raw) > 0 && raw[0] == '{' {
		// Single Scheme object → wrap in array
		var single Scheme
		if err := json.Unmarshal(raw, &single); err != nil {
			return fmt.Errorf("JSON 格式错误: %w", err)
		}
		imported = []Scheme{single}
	} else if len(raw) > 0 && raw[0] == '[' {
		// Array of schemes
		if err := json.Unmarshal(raw, &imported); err != nil {
			return fmt.Errorf("JSON 格式错误: %w", err)
		}
	} else {
		// Not an object or array (AC-10: non-object)
		return fmt.Errorf("JSON 格式错误: 文件根层级应为对象或数组")
	}

	// Check each scheme has an ID (AC-10)
	for i := range imported {
		if strings.TrimSpace(imported[i].ID) == "" {
			return fmt.Errorf("导入的方案缺少 id 字段")
		}
	}

	// Downgrade unknown builtIn providers (AC-06)
	for si := range imported {
		for pi := range imported[si].Providers {
			prov := &imported[si].Providers[pi]
			if prov.BuiltIn && !IsBuiltInProvider(prov.Key) {
				prov.BuiltIn = false
			}
		}
	}

	// Merge with existing schemes (AC-05)
	existing, err := LoadSchemes()
	if err != nil {
		return fmt.Errorf("读取现有方案失败: %w", err)
	}

	merged := make([]Scheme, 0, len(existing)+len(imported))
	covered := make(map[string]bool, len(imported))
	for _, imp := range imported {
		covered[imp.ID] = true
	}

	// Add existing schemes not being overwritten
	for _, ex := range existing {
		if !covered[ex.ID] {
			merged = append(merged, ex)
		}
	}
	// Add all imported schemes (new or overwrite)
	merged = append(merged, imported...)

	// Full validation before save (AC-07)
	for si := range merged {
		scheme := &merged[si]
		if errs := ValidateScheme(scheme); len(errs) > 0 {
			return fmt.Errorf("方案 \"%s\" 校验失败: %s", scheme.Name, errs[0])
		}
		for pi := range scheme.Providers {
			prov := &scheme.Providers[pi]
			// Build list of other providers in this scheme (excluding self) for uniqueness check
			others := make([]Provider, 0, len(scheme.Providers)-1)
			for qi, other := range scheme.Providers {
				if qi != pi {
					others = append(others, other)
				}
			}
			if errs := ValidateProvider(prov, others); len(errs) > 0 {
				return fmt.Errorf("方案 \"%s\" 供应商 \"%s\" 校验失败: %s", scheme.Name, prov.Key, errs[0])
			}
			for mi := range prov.Models {
				model := &prov.Models[mi]
				// Build list of other models in this provider (excluding self) for uniqueness check
				omodels := make([]Model, 0, len(prov.Models)-1)
				for ni, other := range prov.Models {
					if ni != mi {
						omodels = append(omodels, other)
					}
				}
				if errs := ValidateModel(model, omodels); len(errs) > 0 {
					return fmt.Errorf("方案 \"%s\" 供应商 \"%s\" 模型 \"%s\" 校验失败: %s", scheme.Name, prov.Key, model.ID, errs[0])
				}
			}
		}
	}

	// Save (AC-07: only called after all validation passes)
	if err := SaveSchemes(merged); err != nil {
		return fmt.Errorf("保存方案失败: %w", err)
	}

	return nil
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

func findSchemeIndex(schemes []Scheme, id string) int {
	for i := range schemes {
		if schemes[i].ID == id {
			return i
		}
	}
	return -1
}

func findProviderIndex(scheme *Scheme, key string) int {
	for i := range scheme.Providers {
		if scheme.Providers[i].Key == key {
			return i
		}
	}
	return -1
}

func findModelIndex(prov *Provider, id string) int {
	for i := range prov.Models {
		if prov.Models[i].ID == id {
			return i
		}
	}
	return -1
}
