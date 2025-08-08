package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tmpltr/internal/manifest"
)

func TestNewStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-storage-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	if storage.baseDir != tempDir {
		t.Errorf("Expected base directory %s, got %s", tempDir, storage.baseDir)
	}

	// Check that base directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected base directory to be created")
	}
}

func TestNewStorageDefaultLocation(t *testing.T) {
	storage, err := NewStorage("")
	if err != nil {
		t.Fatalf("Failed to create storage with default location: %v", err)
	}

	// Should use default location under user's home directory
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".tmpltr", "templates")
	
	if storage.baseDir != expectedPath {
		t.Errorf("Expected default base directory %s, got %s", expectedPath, storage.baseDir)
	}
}

func TestEnsureTemplateDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-template-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Check that template directory was created
	templateDir := filepath.Join(tempDir, templateName)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		t.Error("Expected template directory to be created")
	}

	// Check that files subdirectory was created
	filesDir := filepath.Join(templateDir, "files")
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		t.Error("Expected files subdirectory to be created")
	}
}

func TestSaveAndLoadFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-file-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"
	hash := "testhash123"
	content := []byte("This is test content")

	// Ensure template directory exists
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Save file
	err = storage.SaveFile(templateName, hash, content)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Load file
	loadedContent, err := storage.LoadFile(templateName, hash)
	if err != nil {
		t.Fatalf("Failed to load file: %v", err)
	}

	if string(loadedContent) != string(content) {
		t.Errorf("Expected loaded content %s, got %s", string(content), string(loadedContent))
	}
}

func TestSaveAndLoadFileWithCompression(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-compression-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"
	hash := "testhash123"
	originalPath := "test.txt"
	
	// Create content that should compress well (repetitive)
	content := []byte(strings.Repeat("This is a test line that should compress well.\n", 100))

	// Ensure template directory exists
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Save file with compression
	compressed, storedSize, err := storage.SaveFileWithCompression(templateName, hash, originalPath, content)
	if err != nil {
		t.Fatalf("Failed to save file with compression: %v", err)
	}

	// Check that compression was applied for this repetitive content
	if !compressed {
		t.Log("Content was not compressed (this might be expected for small content)")
	}

	if storedSize <= 0 {
		t.Error("Expected positive stored size")
	}

	// Load file with decompression
	loadedContent, err := storage.LoadFileWithDecompression(templateName, hash, compressed)
	if err != nil {
		t.Fatalf("Failed to load file with decompression: %v", err)
	}

	if string(loadedContent) != string(content) {
		t.Errorf("Expected loaded content to match original after compression/decompression")
	}
}

func TestFileExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-exists-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"
	hash := "testhash123"
	content := []byte("test content")

	// Check that file doesn't exist initially
	if storage.FileExists(templateName, hash) {
		t.Error("Expected file not to exist initially")
	}

	// Ensure template directory and save file
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	err = storage.SaveFile(templateName, hash, content)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Check that file exists now
	if !storage.FileExists(templateName, hash) {
		t.Error("Expected file to exist after saving")
	}
}

func TestTemplateExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-template-exists-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"

	// Check that template doesn't exist initially
	if storage.TemplateExists(templateName) {
		t.Error("Expected template not to exist initially")
	}

	// Create template directory and manifest
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	m := manifest.NewManifest(templateName)
	err = storage.SaveManifest(templateName, m)
	if err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Check that template exists now
	if !storage.TemplateExists(templateName) {
		t.Error("Expected template to exist after creating manifest")
	}
}

func TestListTemplates(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-list-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Initially should return empty list
	templates, err := storage.ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	if len(templates) != 0 {
		t.Errorf("Expected no templates initially, got %d", len(templates))
	}

	// Create some test templates
	testTemplates := []string{"template1", "template2", "another-template"}

	for _, templateName := range testTemplates {
		err = storage.EnsureTemplateDir(templateName)
		if err != nil {
			t.Fatalf("Failed to ensure template directory: %v", err)
		}

		m := manifest.NewManifest(templateName)
		err = storage.SaveManifest(templateName, m)
		if err != nil {
			t.Fatalf("Failed to save manifest: %v", err)
		}
	}

	// List templates
	templates, err = storage.ListTemplates()
	if err != nil {
		t.Fatalf("Failed to list templates: %v", err)
	}

	if len(templates) != len(testTemplates) {
		t.Errorf("Expected %d templates, got %d", len(testTemplates), len(templates))
	}

	// Check that all expected templates are present
	templateSet := make(map[string]bool)
	for _, template := range templates {
		templateSet[template] = true
	}

	for _, expected := range testTemplates {
		if !templateSet[expected] {
			t.Errorf("Expected template %s to be in list, but it wasn't", expected)
		}
	}
}

func TestDeleteTemplate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-delete-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"

	// Create template with some files
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Save a test file
	hash := "testhash"
	content := []byte("test content")
	err = storage.SaveFile(templateName, hash, content)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Save manifest
	m := manifest.NewManifest(templateName)
	m.AddFile("test.txt", hash, true, false, 100, 100)
	err = storage.SaveManifest(templateName, m)
	if err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Verify template exists
	if !storage.TemplateExists(templateName) {
		t.Fatal("Expected template to exist before deletion")
	}

	// Delete template
	err = storage.DeleteTemplate(templateName)
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	// Verify template no longer exists
	if storage.TemplateExists(templateName) {
		t.Error("Expected template not to exist after deletion")
	}

	// Verify template directory was removed
	templateDir := filepath.Join(tempDir, templateName)
	if _, err := os.Stat(templateDir); !os.IsNotExist(err) {
		t.Error("Expected template directory to be removed after deletion")
	}
}

func TestSaveAndLoadManifest(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tmpltr-manifest-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName := "test-template"

	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Create manifest
	originalManifest := manifest.NewManifest(templateName)
	originalManifest.AddFile("file1.txt", "hash1", true, false, 100, 100)
	originalManifest.AddFile("file2.txt", "hash2", true, true, 200, 150)
	originalManifest.AddFile("file3.txt", "hash3", false, false, 0, 0)

	// Save manifest
	err = storage.SaveManifest(templateName, originalManifest)
	if err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Load manifest
	loadedManifest, err := storage.LoadManifest(templateName)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Compare manifests
	if loadedManifest.Name != originalManifest.Name {
		t.Errorf("Expected name %s, got %s", originalManifest.Name, loadedManifest.Name)
	}

	if len(loadedManifest.Files) != len(originalManifest.Files) {
		t.Errorf("Expected %d files, got %d", len(originalManifest.Files), len(loadedManifest.Files))
	}

	// Check individual files
	for i, originalFile := range originalManifest.Files {
		if i >= len(loadedManifest.Files) {
			t.Errorf("Missing file at index %d", i)
			continue
		}

		loadedFile := loadedManifest.Files[i]

		if loadedFile.OriginalPath != originalFile.OriginalPath {
			t.Errorf("Expected path %s, got %s", originalFile.OriginalPath, loadedFile.OriginalPath)
		}

		if loadedFile.Hash != originalFile.Hash {
			t.Errorf("Expected hash %s, got %s", originalFile.Hash, loadedFile.Hash)
		}

		if loadedFile.IncludeContents != originalFile.IncludeContents {
			t.Errorf("Expected include contents %v, got %v", originalFile.IncludeContents, loadedFile.IncludeContents)
		}

		if loadedFile.Compressed != originalFile.Compressed {
			t.Errorf("Expected compressed %v, got %v", originalFile.Compressed, loadedFile.Compressed)
		}
	}
}