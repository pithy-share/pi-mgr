package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// storePath returns the path to config.json in %APPDATA%/pi-mgr/
func storePath() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback for unusual environments
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "pi-mgr", "config.json")
	}
	return filepath.Join(cfgDir, "pi-mgr", "config.json")
}

// ensureConfigDir creates the store directory if it doesn't exist
func ensureConfigDir() {
	dir := filepath.Dir(storePath())
	os.MkdirAll(dir, 0755)
}

// LoadConfig reads the single global config from disk. Returns empty config if file doesn't exist.
func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(storePath())
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return &Config{}, nil
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}
	return &cfg, nil
}

// SaveConfig atomically writes config to disk (write temp + rename)
func SaveConfig(cfg *Config) error {
	ensureConfigDir()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	tmpPath := storePath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	if err := os.Rename(tmpPath, storePath()); err != nil {
		// On Windows, rename may fail across volumes; fallback to direct write
		if err2 := os.WriteFile(storePath(), data, 0644); err2 != nil {
			return fmt.Errorf("failed to save config: %w", err2)
		}
		os.Remove(tmpPath)
	}
	return nil
}

// findProviderIndex finds a provider by key in the config
func findProviderIndex(cfg *Config, key string) int {
	for i := range cfg.Providers {
		if cfg.Providers[i].Key == key {
			return i
		}
	}
	return -1
}
