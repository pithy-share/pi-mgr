package main

import "strings"

// ValidateScheme checks scheme-level constraints. Returns nil if valid.
func ValidateScheme(scheme *Scheme) []string {
	var errs []string
	if strings.TrimSpace(scheme.Name) == "" {
		errs = append(errs, "方案名称不能为空")
	}
	return errs
}

// ValidateProvider checks provider-level constraints.
// allProviders is the list of other providers in the scheme (excluding the one being validated).
func ValidateProvider(prov *Provider, allProviders []Provider) []string {
	var errs []string

	if strings.TrimSpace(prov.Key) == "" {
		errs = append(errs, "供应商标识不能为空")
	}

	if !prov.BuiltIn {
		if strings.TrimSpace(prov.BaseURL) == "" {
			errs = append(errs, "自定义供应商的 baseUrl 不能为空")
		}
		if prov.APIType == "" {
			errs = append(errs, "自定义供应商的 API 类型不能为空")
		}
	}

	// Check key uniqueness within scheme (skip self for updates)
	for _, other := range allProviders {
		if other.Key == prov.Key {
			errs = append(errs, "供应商标识已存在")
			break
		}
	}

	return errs
}

// ValidateModel checks model-level constraints.
// existingModels includes the model itself if it already exists in the provider.
func ValidateModel(m *Model, existingModels []Model) []string {
	var errs []string

	if strings.TrimSpace(m.ID) == "" {
		errs = append(errs, "模型 ID 不能为空")
	}

	// Check ID uniqueness within provider (skip self for updates)
	for _, other := range existingModels {
		if other.ID == m.ID {
			errs = append(errs, "模型 ID 在该供应商下已存在")
			break
		}
	}

	return errs
}
