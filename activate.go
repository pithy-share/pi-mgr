package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// activatePath returns the path to pi's models.json
func activatePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback: use %USERPROFILE%
		home = os.Getenv("USERPROFILE")
	}
	return filepath.Join(home, ".pi", "agent", "models.json")
}

// ActivateScheme serializes the scheme and writes it to pi's models.json.
// Creates the directory if it doesn't exist (AC-31).
func ActivateScheme(scheme *Scheme) error {
	data, err := SerializeToModelsJSON(scheme)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	targetPath := activatePath()
	dir := filepath.Dir(targetPath)

	// Create directory if it doesn't exist (AC-31)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("无法创建目录 %s: %w", dir, err)
	}

	// Write models.json (AC-32: return error if write fails)
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("写入 models.json 失败: %w", err)
	}

	return nil
}