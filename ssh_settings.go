package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// appSettings holds app-level persistent settings (separate from schemes.json)
type appSettings struct {
	SSHAddress string `json:"sshAddress"`
}

// appSettingsPath returns the path to settings.json in %APPDATA%/pi-mgr/
func appSettingsPath() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback for unusual environments
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "pi-mgr", "settings.json")
	}
	return filepath.Join(cfgDir, "pi-mgr", "settings.json")
}

// ensureAppSettingsDir creates the settings directory if it doesn't exist
func ensureAppSettingsDir() {
	dir := filepath.Dir(appSettingsPath())
	os.MkdirAll(dir, 0755)
}

// loadAppSettings reads app settings from disk. Returns empty settings if file doesn't exist.
func loadAppSettings() (*appSettings, error) {
	data, err := os.ReadFile(appSettingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return &appSettings{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return &appSettings{}, nil
	}
	var settings appSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		// Graceful degradation: return empty settings on parse error
		return &appSettings{}, nil
	}
	return &settings, nil
}

// saveAppSettings atomically writes app settings to disk (temp + rename)
func saveAppSettings(settings *appSettings) error {
	ensureAppSettingsDir()
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	path := appSettingsPath()
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		// Rename may fail across volumes; fallback to direct write
		if err2 := os.WriteFile(path, data, 0644); err2 != nil {
			os.Remove(tmpPath)
			return err2
		}
		os.Remove(tmpPath)
	}
	return nil
}

// SaveSSHAddress persists the SSH address to settings.json
func (a *App) SaveSSHAddress(address string) error {
	settings, err := loadAppSettings()
	if err != nil {
		return err
	}
	settings.SSHAddress = address
	return saveAppSettings(settings)
}

// LoadSSHAddress reads the saved SSH address from settings.json
func (a *App) LoadSSHAddress() (string, error) {
	settings, err := loadAppSettings()
	if err != nil {
		return "", err
	}
	return settings.SSHAddress, nil
}
