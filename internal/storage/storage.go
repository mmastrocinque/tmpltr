package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"tmpltr/internal/manifest"
)

const (
	DefaultTemplateDir = ".tmpltr/templates"
	ManifestFileName   = "manifest.json"
	FilesSubDir        = "files"
)

// Storage manages template storage operations
type Storage struct {
	baseDir string
}

// NewStorage creates a new Storage instance with the given base directory
// If baseDir is empty, uses the default directory in the user's home
func NewStorage(baseDir string) (*Storage, error) {
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, DefaultTemplateDir)
	}

	return &Storage{baseDir: baseDir}, nil
}

// GetTemplatePath returns the full path to a template directory
func (s *Storage) GetTemplatePath(templateName string) string {
	return filepath.Join(s.baseDir, templateName)
}

// GetManifestPath returns the full path to a template's manifest file
func (s *Storage) GetManifestPath(templateName string) string {
	return filepath.Join(s.GetTemplatePath(templateName), ManifestFileName)
}

// GetFilesPath returns the full path to a template's files directory
func (s *Storage) GetFilesPath(templateName string) string {
	return filepath.Join(s.GetTemplatePath(templateName), FilesSubDir)
}

// GetFileContentPath returns the full path to a stored file by hash
func (s *Storage) GetFileContentPath(templateName, hash string) string {
	return filepath.Join(s.GetFilesPath(templateName), hash)
}

// TemplateExists checks if a template exists in storage
func (s *Storage) TemplateExists(templateName string) bool {
	manifestPath := s.GetManifestPath(templateName)
	_, err := os.Stat(manifestPath)
	return err == nil
}

// EnsureTemplateDir creates the template directory structure if it doesn't exist
func (s *Storage) EnsureTemplateDir(templateName string) error {
	templatePath := s.GetTemplatePath(templateName)
	filesPath := s.GetFilesPath(templateName)

	if err := os.MkdirAll(templatePath, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	if err := os.MkdirAll(filesPath, 0755); err != nil {
		return fmt.Errorf("failed to create files directory: %w", err)
	}

	return nil
}

// SaveFile saves file content to the storage with the given hash as filename
func (s *Storage) SaveFile(templateName, hash string, content []byte) error {
	filePath := s.GetFileContentPath(templateName, hash)
	
	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to save file with hash %s: %w", hash, err)
	}

	return nil
}

// LoadFile loads file content from storage by hash
func (s *Storage) LoadFile(templateName, hash string) ([]byte, error) {
	filePath := s.GetFileContentPath(templateName, hash)
	
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load file with hash %s: %w", hash, err)
	}

	return content, nil
}

// CopyFileToStorage copies a source file to storage with the given hash
func (s *Storage) CopyFileToStorage(sourcePath, templateName, hash string) error {
	destPath := s.GetFileContentPath(templateName, hash)
	
	return copyFile(sourcePath, destPath)
}

// SaveManifest saves a manifest to storage
func (s *Storage) SaveManifest(templateName string, m *manifest.Manifest) error {
	manifestPath := s.GetManifestPath(templateName)
	return manifest.SaveManifest(m, manifestPath)
}

// LoadManifest loads a manifest from storage
func (s *Storage) LoadManifest(templateName string) (*manifest.Manifest, error) {
	manifestPath := s.GetManifestPath(templateName)
	return manifest.LoadManifest(manifestPath)
}

// DeleteTemplate removes a template and all its files from storage
func (s *Storage) DeleteTemplate(templateName string) error {
	templatePath := s.GetTemplatePath(templateName)
	
	err := os.RemoveAll(templatePath)
	if err != nil {
		return fmt.Errorf("failed to delete template %s: %w", templateName, err)
	}

	return nil
}

// ListTemplates returns a list of all template names in storage
func (s *Storage) ListTemplates() ([]string, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			manifestPath := s.GetManifestPath(entry.Name())
			if _, err := os.Stat(manifestPath); err == nil {
				templates = append(templates, entry.Name())
			}
		}
	}

	return templates, nil
}

// FileExists checks if a file with the given hash exists in storage
func (s *Storage) FileExists(templateName, hash string) bool {
	filePath := s.GetFileContentPath(templateName, hash)
	_, err := os.Stat(filePath)
	return err == nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}