package main

import (
	"fmt"
	"strings"
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
