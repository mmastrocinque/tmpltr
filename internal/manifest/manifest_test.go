package manifest

import (
	"testing"
	"time"
)

func TestNewManifest(t *testing.T) {
	templateName := "test-template"
	m := NewManifest(templateName)

	if m.Name != templateName {
		t.Errorf("Expected name %s, got %s", templateName, m.Name)
	}

	// Note: Version field not implemented in current manifest structure
	// This test can be added when versioning is implemented

	if len(m.Files) != 0 {
		t.Errorf("Expected empty files list, got %d files", len(m.Files))
	}

	// Check that CreatedAt is set and recent
	now := time.Now()
	if m.CreatedAt.After(now) {
		t.Error("CreatedAt should not be in the future")
	}

	if now.Sub(m.CreatedAt) > time.Minute {
		t.Error("CreatedAt should be recent (within last minute)")
	}
}

func TestManifestAddFile(t *testing.T) {
	m := NewManifest("test-template")

	// Add first file
	m.AddFile("path/to/file1.txt", "hash1", true, false, 100, 100)

	if len(m.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(m.Files))
	}

	file := m.Files[0]
	if file.OriginalPath != "path/to/file1.txt" {
		t.Errorf("Expected path 'path/to/file1.txt', got %s", file.OriginalPath)
	}

	if file.Hash != "hash1" {
		t.Errorf("Expected hash 'hash1', got %s", file.Hash)
	}

	if !file.IncludeContents {
		t.Error("Expected IncludeContents to be true")
	}

	if file.Compressed {
		t.Error("Expected Compressed to be false")
	}

	if file.OriginalSize != 100 {
		t.Errorf("Expected OriginalSize 100, got %d", file.OriginalSize)
	}

	if file.StoredSize != 100 {
		t.Errorf("Expected StoredSize 100, got %d", file.StoredSize)
	}

	// Add second file with compression
	m.AddFile("file2.txt", "hash2", true, true, 200, 150)

	if len(m.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(m.Files))
	}

	file2 := m.Files[1]
	if !file2.Compressed {
		t.Error("Expected second file to be compressed")
	}

	if file2.StoredSize != 150 {
		t.Errorf("Expected StoredSize 150, got %d", file2.StoredSize)
	}

	// Add file without contents
	m.AddFile("file3.txt", "hash3", false, false, 0, 0)

	if len(m.Files) != 3 {
		t.Fatalf("Expected 3 files, got %d", len(m.Files))
	}

	file3 := m.Files[2]
	if file3.IncludeContents {
		t.Error("Expected third file not to include contents")
	}
}

func TestGetFileCount(t *testing.T) {
	m := NewManifest("test-template")

	if m.GetFileCount() != 0 {
		t.Errorf("Expected 0 files, got %d", m.GetFileCount())
	}

	m.AddFile("file1.txt", "hash1", true, false, 100, 100)
	if m.GetFileCount() != 1 {
		t.Errorf("Expected 1 file, got %d", m.GetFileCount())
	}

	m.AddFile("file2.txt", "hash2", false, false, 0, 0)
	if m.GetFileCount() != 2 {
		t.Errorf("Expected 2 files, got %d", m.GetFileCount())
	}
}

func TestGetFilesWithContents(t *testing.T) {
	m := NewManifest("test-template")

	// Initially no files with contents
	filesWithContents := m.GetFilesWithContents()
	if len(filesWithContents) != 0 {
		t.Errorf("Expected 0 files with contents, got %d", len(filesWithContents))
	}

	// Add file with contents
	m.AddFile("file1.txt", "hash1", true, false, 100, 100)
	filesWithContents = m.GetFilesWithContents()
	if len(filesWithContents) != 1 {
		t.Errorf("Expected 1 file with contents, got %d", len(filesWithContents))
	}

	// Add file without contents
	m.AddFile("file2.txt", "hash2", false, false, 0, 0)
	filesWithContents = m.GetFilesWithContents()
	if len(filesWithContents) != 1 {
		t.Errorf("Expected still 1 file with contents, got %d", len(filesWithContents))
	}

	// Add another file with contents
	m.AddFile("file3.txt", "hash3", true, true, 200, 150)
	filesWithContents = m.GetFilesWithContents()
	if len(filesWithContents) != 2 {
		t.Errorf("Expected 2 files with contents, got %d", len(filesWithContents))
	}

	// Verify the correct files are returned
	expectedHashes := map[string]bool{"hash1": true, "hash3": true}
	for _, file := range filesWithContents {
		if !expectedHashes[file.Hash] {
			t.Errorf("Unexpected file with contents: %s", file.Hash)
		}
	}
}

func TestGetCompressionStats(t *testing.T) {
	m := NewManifest("test-template")

	// Initially no compression stats
	compressedFiles, originalSize, storedSize := m.GetCompressionStats()
	if compressedFiles != 0 {
		t.Errorf("Expected 0 compressed files, got %d", compressedFiles)
	}
	if originalSize != 0 {
		t.Errorf("Expected 0 original size, got %d", originalSize)
	}
	if storedSize != 0 {
		t.Errorf("Expected 0 stored size, got %d", storedSize)
	}

	// Add uncompressed file
	m.AddFile("file1.txt", "hash1", true, false, 100, 100)
	compressedFiles, originalSize, storedSize = m.GetCompressionStats()
	if compressedFiles != 0 {
		t.Errorf("Expected 0 compressed files, got %d", compressedFiles)
	}
	if originalSize != 100 {
		t.Errorf("Expected 100 original size, got %d", originalSize)
	}
	if storedSize != 100 {
		t.Errorf("Expected 100 stored size, got %d", storedSize)
	}

	// Add compressed file
	m.AddFile("file2.txt", "hash2", true, true, 200, 150)
	compressedFiles, originalSize, storedSize = m.GetCompressionStats()
	if compressedFiles != 1 {
		t.Errorf("Expected 1 compressed file, got %d", compressedFiles)
	}
	if originalSize != 300 {
		t.Errorf("Expected 300 original size, got %d", originalSize)
	}
	if storedSize != 250 {
		t.Errorf("Expected 250 stored size, got %d", storedSize)
	}

	// Add file without contents (should not affect compression stats)
	m.AddFile("file3.txt", "hash3", false, false, 0, 0)
	compressedFiles, originalSize, storedSize = m.GetCompressionStats()
	if compressedFiles != 1 {
		t.Errorf("Expected 1 compressed file after adding structure-only file, got %d", compressedFiles)
	}
	if originalSize != 300 {
		t.Errorf("Expected 300 original size after adding structure-only file, got %d", originalSize)
	}
	if storedSize != 250 {
		t.Errorf("Expected 250 stored size after adding structure-only file, got %d", storedSize)
	}

	// Add another compressed file
	m.AddFile("file4.txt", "hash4", true, true, 400, 300)
	compressedFiles, originalSize, storedSize = m.GetCompressionStats()
	if compressedFiles != 2 {
		t.Errorf("Expected 2 compressed files, got %d", compressedFiles)
	}
	if originalSize != 700 {
		t.Errorf("Expected 700 original size, got %d", originalSize)
	}
	if storedSize != 550 {
		t.Errorf("Expected 550 stored size, got %d", storedSize)
	}
}

func TestGetCompressionRatio(t *testing.T) {
	m := NewManifest("test-template")

	// With no files, ratio should be 0.0 (based on implementation: 0/0 case)
	ratio := m.GetCompressionRatio()
	if ratio != 0.0 {
		t.Errorf("Expected ratio 0.0 for no files, got %f", ratio)
	}

	// Add uncompressed file
	m.AddFile("file1.txt", "hash1", true, false, 100, 100)
	ratio = m.GetCompressionRatio()
	if ratio != 1.0 {
		t.Errorf("Expected ratio 1.0 for uncompressed file, got %f", ratio)
	}

	// Add compressed file (200 -> 150, so 150/200 = 0.75)
	m.AddFile("file2.txt", "hash2", true, true, 200, 150)
	ratio = m.GetCompressionRatio()
	expected := float64(250) / float64(300) // (100 + 150) / (100 + 200)
	if ratio != expected {
		t.Errorf("Expected ratio %f, got %f", expected, ratio)
	}

	// Test with highly compressed file
	m.AddFile("file3.txt", "hash3", true, true, 1000, 100)
	ratio = m.GetCompressionRatio()
	expected = float64(350) / float64(1300) // (100 + 150 + 100) / (100 + 200 + 1000)
	if ratio != expected {
		t.Errorf("Expected ratio %f, got %f", expected, ratio)
	}
}

func TestValidateManifest(t *testing.T) {
	// Test valid manifest
	m := NewManifest("valid-template")
	m.AddFile("file1.txt", "hash1", true, false, 100, 100)

	err := ValidateManifest(m)
	if err != nil {
		t.Errorf("Expected no error for valid manifest, got: %v", err)
	}

	// Test manifest with empty name
	m2 := NewManifest("")
	err = ValidateManifest(m2)
	if err == nil {
		t.Error("Expected error for manifest with empty name")
	}

	// Test manifest with invalid version - skipped as version field not implemented
	// This test can be added when versioning is implemented

	// Test manifest with file with empty path
	m4 := NewManifest("test-template")
	m4.Files = append(m4.Files, FileEntry{
		OriginalPath:    "",
		Hash:            "hash1",
		IncludeContents: true,
		Compressed:      false,
		OriginalSize:    100,
		StoredSize:      100,
	})
	err = ValidateManifest(m4)
	if err == nil {
		t.Error("Expected error for manifest with file with empty path")
	}

	// Test manifest with file with empty hash
	m5 := NewManifest("test-template")
	m5.Files = append(m5.Files, FileEntry{
		OriginalPath:    "file.txt",
		Hash:            "",
		IncludeContents: true,
		Compressed:      false,
		OriginalSize:    100,
		StoredSize:      100,
	})
	err = ValidateManifest(m5)
	if err == nil {
		t.Error("Expected error for manifest with file with empty hash")
	}
}

func TestFileEntry(t *testing.T) {
	// Test FileEntry creation and properties
	entry := FileEntry{
		OriginalPath:    "test/file.txt",
		Hash:            "abcd1234",
		IncludeContents: true,
		Compressed:      true,
		OriginalSize:    1000,
		StoredSize:      750,
	}

	if entry.OriginalPath != "test/file.txt" {
		t.Errorf("Expected path 'test/file.txt', got %s", entry.OriginalPath)
	}

	if entry.Hash != "abcd1234" {
		t.Errorf("Expected hash 'abcd1234', got %s", entry.Hash)
	}

	if !entry.IncludeContents {
		t.Error("Expected IncludeContents to be true")
	}

	if !entry.Compressed {
		t.Error("Expected Compressed to be true")
	}

	if entry.OriginalSize != 1000 {
		t.Errorf("Expected OriginalSize 1000, got %d", entry.OriginalSize)
	}

	if entry.StoredSize != 750 {
		t.Errorf("Expected StoredSize 750, got %d", entry.StoredSize)
	}
}