package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveManifest saves a manifest to the specified file path
func SaveManifest(manifest *Manifest, filePath string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// LoadManifest loads a manifest from the specified file path
func LoadManifest(filePath string) (*Manifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	var manifest Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// ManifestExists checks if a manifest file exists at the given path
func ManifestExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// ValidateManifest performs basic validation on a manifest
func ValidateManifest(manifest *Manifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("manifest name cannot be empty")
	}

	if len(manifest.Files) == 0 {
		return fmt.Errorf("manifest must contain at least one file")
	}

	pathSet := make(map[string]bool)
	for _, file := range manifest.Files {
		if file.OriginalPath == "" {
			return fmt.Errorf("file original_path cannot be empty")
		}

		if file.Hash == "" {
			return fmt.Errorf("file hash cannot be empty for path: %s", file.OriginalPath)
		}

		if pathSet[file.OriginalPath] {
			return fmt.Errorf("duplicate file path in manifest: %s", file.OriginalPath)
		}
		pathSet[file.OriginalPath] = true

		if filepath.IsAbs(file.OriginalPath) {
			return fmt.Errorf("file path must be relative: %s", file.OriginalPath)
		}
	}

	return nil
}