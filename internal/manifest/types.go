package manifest

import "time"

// FileEntry represents a single file in the template manifest
type FileEntry struct {
	OriginalPath    string `json:"original_path"`    // Relative path of the file in the source directory
	Hash            string `json:"hash"`             // SHA256 hash used for lookup in the files/ store
	IncludeContents bool   `json:"include_contents"` // Boolean flag indicating whether file contents were saved
}

// Manifest represents the complete template manifest structure
type Manifest struct {
	Name      string      `json:"name"`       // Template name
	CreatedAt time.Time   `json:"created_at"` // Template creation timestamp
	Files     []FileEntry `json:"files"`      // List of files in the template
}

// NewManifest creates a new manifest with the given name
func NewManifest(name string) *Manifest {
	return &Manifest{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Files:     make([]FileEntry, 0),
	}
}

// AddFile adds a new file entry to the manifest
func (m *Manifest) AddFile(originalPath, hash string, includeContents bool) {
	entry := FileEntry{
		OriginalPath:    originalPath,
		Hash:            hash,
		IncludeContents: includeContents,
	}
	m.Files = append(m.Files, entry)
}

// GetFileByPath returns the file entry for the given original path
func (m *Manifest) GetFileByPath(originalPath string) *FileEntry {
	for i := range m.Files {
		if m.Files[i].OriginalPath == originalPath {
			return &m.Files[i]
		}
	}
	return nil
}

// GetFileCount returns the total number of files in the manifest
func (m *Manifest) GetFileCount() int {
	return len(m.Files)
}

// GetFilesWithContents returns all files that include contents
func (m *Manifest) GetFilesWithContents() []FileEntry {
	var filesWithContents []FileEntry
	for _, file := range m.Files {
		if file.IncludeContents {
			filesWithContents = append(filesWithContents, file)
		}
	}
	return filesWithContents
}