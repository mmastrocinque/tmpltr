package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tmpltr/internal/hash"
	"tmpltr/internal/manifest"
	"tmpltr/internal/storage"
)

var (
	templateName    string
	ignoreContents  bool
)

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:   "make <target_directory>",
	Short: "Create a template from a target directory",
	Long: `Create a template from a target directory by scanning all files,
hashing their contents, and saving the structure with optional file contents.

Example:
  tmpltr make ./my-project --name="my-template"
  tmpltr make ./my-project --name="structure-only" --ignore-contents`,
	Args: cobra.ExactArgs(1),
	RunE: runMake,
}

func init() {
	makeCmd.Flags().StringVarP(&templateName, "name", "n", "", "Name for the template (required)")
	makeCmd.Flags().BoolVar(&ignoreContents, "ignore-contents", false, "Only save file structure, ignore contents")
	makeCmd.MarkFlagRequired("name")
}

// runMake executes the make command logic
func runMake(cmd *cobra.Command, args []string) error {
	targetDir := args[0]
	
	// Validate target directory
	if err := validateTargetDirectory(targetDir); err != nil {
		return err
	}

	// Validate template name
	if err := validateTemplateName(templateName); err != nil {
		return err
	}

	// Initialize storage
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check if template already exists
	if storage.TemplateExists(templateName) {
		return fmt.Errorf("template '%s' already exists", templateName)
	}

	// Create template structure
	if err := storage.EnsureTemplateDir(templateName); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// Create manifest
	m := manifest.NewManifest(templateName)

	// Scan and process files
	err = scanDirectory(targetDir, targetDir, m, storage)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	// Validate manifest
	if err := manifest.ValidateManifest(m); err != nil {
		return fmt.Errorf("invalid manifest generated: %w", err)
	}

	// Save manifest
	if err := storage.SaveManifest(templateName, m); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	fmt.Printf("Successfully created template '%s' with %d files\n", templateName, m.GetFileCount())
	if ignoreContents {
		fmt.Printf("Template saved structure only (contents ignored)\n")
	} else {
		contentFiles := len(m.GetFilesWithContents())
		fmt.Printf("Template saved with %d files containing content\n", contentFiles)
	}

	return nil
}

// validateTargetDirectory checks if the target directory is valid
func validateTargetDirectory(targetDir string) error {
	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("target directory does not exist: %s", targetDir)
		}
		return fmt.Errorf("failed to access target directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("target path is not a directory: %s", targetDir)
	}

	return nil
}

// validateTemplateName checks if the template name is valid
func validateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return fmt.Errorf("template name contains invalid characters")
	}

	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("template name cannot start with a dot")
	}

	return nil
}

// scanDirectory recursively scans a directory and processes all files
func scanDirectory(rootDir, currentDir string, m *manifest.Manifest, storage *storage.Storage) error {
	return filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory: %w", err)
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Skip hidden files and directories
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		// Calculate relative path from root
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path: %w", err)
		}

		// Process the file
		return processFile(path, relPath, m, storage)
	})
}

// processFile processes a single file for the template
func processFile(filePath, relativePath string, m *manifest.Manifest, storage *storage.Storage) error {
	var fileHash string
	var err error

	if ignoreContents {
		// Generate hash based on file path for ignore-contents mode
		fileHash = hash.GenerateFileNameHash(relativePath)
		
		// Create empty file in storage if it doesn't exist
		if !storage.FileExists(templateName, fileHash) {
			if err := storage.SaveFile(templateName, fileHash, []byte("")); err != nil {
				return fmt.Errorf("failed to save empty file placeholder: %w", err)
			}
		}
	} else {
		// Calculate actual file content hash
		fileHash, err = hash.HashFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to hash file %s: %w", relativePath, err)
		}

		// Copy file to storage if it doesn't already exist (deduplication)
		if !storage.FileExists(templateName, fileHash) {
			if err := storage.CopyFileToStorage(filePath, templateName, fileHash); err != nil {
				return fmt.Errorf("failed to copy file %s to storage: %w", relativePath, err)
			}
		}
	}

	// Add file to manifest
	m.AddFile(relativePath, fileHash, !ignoreContents)

	return nil
}