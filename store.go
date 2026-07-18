package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// storePath returns the path to schemes.json in %APPDATA%/pi-mgr/
func storePath() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback for unusual environments
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "pi-mgr", "schemes.json")
	}
	return filepath.Join(cfgDir, "pi-mgr", "schemes.json")
}

// ensureStoreDir creates the store directory if it doesn't exist
func ensureStoreDir() {
	dir := filepath.Dir(storePath())
	os.MkdirAll(dir, 0755)
}

// LoadSchemes reads all schemes from disk. Returns empty slice if file doesn't exist.
func LoadSchemes() ([]Scheme, error) {
	data, err := os.ReadFile(storePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []Scheme{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []Scheme{}, nil
	}
	var schemes []Scheme
	if err := json.Unmarshal(data, &schemes); err != nil {
		return nil, fmt.Errorf("failed to parse schemes.json: %w", err)
	}
	return schemes, nil
}

// SaveSchemes atomically writes schemes to disk (write temp + rename)
func SaveSchemes(schemes []Scheme) error {
	ensureStoreDir()
	data, err := json.MarshalIndent(schemes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schemes: %w", err)
	}
	tmpPath := storePath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write schemes: %w", err)
	}
	if err := os.Rename(tmpPath, storePath()); err != nil {
		// On Windows, rename may fail across volumes; fallback to direct write
		if err2 := os.WriteFile(storePath(), data, 0644); err2 != nil {
			return fmt.Errorf("failed to save schemes: %w", err2)
		}
		os.Remove(tmpPath)
	}
	return nil
}

// GetScheme finds a scheme by ID
func GetScheme(id string) (*Scheme, error) {
	schemes, err := LoadSchemes()
	if err != nil {
		return nil, err
	}
	for i := range schemes {
		if schemes[i].ID == id {
			return &schemes[i], nil
		}
	}
	return nil, fmt.Errorf("scheme not found: %s", id)
}

// ---------------------------------------------------------------------------
// Active scheme tracking
// ---------------------------------------------------------------------------

// activeStorePath returns the path to active.json in %APPDATA%/pi-mgr/
func activeStorePath() string {
	return filepath.Join(filepath.Dir(storePath()), "active.json")
}

// GetActiveSchemeID reads the currently active scheme ID from active.json.
// Returns empty string if no scheme has been activated yet.
func GetActiveSchemeID() (string, error) {
	data, err := os.ReadFile(activeStorePath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var active struct {
		ActiveSchemeID string `json:"activeSchemeId"`
	}
	if err := json.Unmarshal(data, &active); err != nil {
		return "", nil
	}
	return active.ActiveSchemeID, nil
}

// SaveActiveSchemeID persists the active scheme ID to active.json.
func SaveActiveSchemeID(id string) error {
	ensureStoreDir()
	active := struct {
		ActiveSchemeID string `json:"activeSchemeId"`
	}{ActiveSchemeID: id}
	data, err := json.MarshalIndent(active, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal active state: %w", err)
	}
	tmpPath := activeStorePath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write active state: %w", err)
	}
	if err := os.Rename(tmpPath, activeStorePath()); err != nil {
		if err2 := os.WriteFile(activeStorePath(), data, 0644); err2 != nil {
			return fmt.Errorf("failed to save active state: %w", err2)
		}
		os.Remove(tmpPath)
	}
	return nil
}

// ClearActiveSchemeID removes the active.json file (e.g. when the active scheme is deleted).
func ClearActiveSchemeID() {
	os.Remove(activeStorePath())
}

// newUUID generates a simple UUID v4-like identifier
func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
