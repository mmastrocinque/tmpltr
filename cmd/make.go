package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tmpltr/internal/hash"
	"tmpltr/internal/ignore"
	"tmpltr/internal/manifest"
	"tmpltr/internal/storage"
)

var (
	templateName    string
	ignoreContents  bool
	ignoreFiles     []string
	noCompression   bool
)

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:   "make <target_directory>",
	Short: "Create a template from a target directory",
	Long: `Create a template from a target directory by scanning all files,
hashing their contents, and saving the structure with optional file contents.
Supports compression and selective file ignoring.

Examples:
  tmpltr make ./my-project --name="my-template"
  tmpltr make ./my-project --name="structure-only" --ignore-contents
  tmpltr make ./my-project --name="selective" --ignore-files="*.log,node_modules/,temp.txt"
  tmpltr make ./my-project --name="uncompressed" --no-compression`,
	Args: cobra.ExactArgs(1),
	RunE: runMake,
}

func init() {
	makeCmd.Flags().StringVarP(&templateName, "name", "n", "", "Name for the template (required)")
	makeCmd.Flags().BoolVar(&ignoreContents, "ignore-contents", false, "Only save file structure, ignore contents")
	makeCmd.Flags().StringSliceVar(&ignoreFiles, "ignore-files", []string{}, "Comma-separated list of files/patterns to ignore")
	makeCmd.Flags().BoolVar(&noCompression, "no-compression", false, "Disable compression of file contents")
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

	// Setup ignore rules
	ignoreRules := ignore.NewIgnoreRules(targetDir)
	
	// Add default ignore patterns
	ignoreRules.AddDefaultPatterns()
	
	// Load .tmpltrignore file if it exists
	if err := ignoreRules.LoadIgnoreFile(); err != nil {
		return fmt.Errorf("failed to load ignore file: %w", err)
	}
	
	// Add command-line ignore patterns
	ignoreRules.AddPatterns(ignoreFiles)

	// Scan and process files
	err = scanDirectory(targetDir, targetDir, m, storage, ignoreRules)
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

	// Display results
	compressedFiles, originalSize, storedSize := m.GetCompressionStats()
	
	fmt.Printf("Successfully created template '%s' with %d files\n", templateName, m.GetFileCount())
	if ignoreContents {
		fmt.Printf("Template saved structure only (contents ignored)\n")
	} else {
		contentFiles := len(m.GetFilesWithContents())
		fmt.Printf("Template saved with %d files containing content\n", contentFiles)
		
		if !noCompression && compressedFiles > 0 {
			compressionRatio := (1.0 - m.GetCompressionRatio()) * 100
			fmt.Printf("Compression: %d files compressed, %.1f%% size reduction (%.1f KB → %.1f KB)\n",
				compressedFiles, compressionRatio,
				float64(originalSize)/1024, float64(storedSize)/1024)
		}
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
func scanDirectory(rootDir, currentDir string, m *manifest.Manifest, storage *storage.Storage, ignoreRules *ignore.IgnoreRules) error {
	return filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory: %w", err)
		}

		// Skip directories
		if d.IsDir() {
			// Check if directory should be ignored
			if ignoreRules.ShouldIgnore(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be ignored
		if ignoreRules.ShouldIgnore(path) {
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
	var compressed bool
	var originalSize, storedSize int64

	// Get file info for size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", relativePath, err)
	}
	originalSize = fileInfo.Size()

	if ignoreContents {
		// Generate hash based on file path for ignore-contents mode
		fileHash = hash.GenerateFileNameHash(relativePath)
		storedSize = 0
		
		// Create empty file in storage if it doesn't exist
		if !storage.FileExists(templateName, fileHash) {
			if err := storage.SaveFile(templateName, fileHash, []byte("")); err != nil {
				return fmt.Errorf("failed to save empty file placeholder: %w", err)
			}
		}
	} else {
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", relativePath, err)
		}

		// Calculate actual file content hash
		fileHash = hash.HashBytes(content)

		// Save file to storage if it doesn't already exist (deduplication)
		if !storage.FileExists(templateName, fileHash) {
			if noCompression {
				// Save without compression
				if err := storage.SaveFile(templateName, fileHash, content); err != nil {
					return fmt.Errorf("failed to save file %s to storage: %w", relativePath, err)
				}
				compressed = false
				storedSize = originalSize
			} else {
				// Save with optional compression
				isCompressed, storedBytes, err := storage.SaveFileWithCompression(templateName, fileHash, relativePath, content)
				if err != nil {
					return fmt.Errorf("failed to save file %s to storage: %w", relativePath, err)
				}
				compressed = isCompressed
				storedSize = storedBytes
			}
		} else {
			// File already exists, we need to determine if it was compressed
			// For deduplication, we'll assume the same compression settings
			storedSize = originalSize
			compressed = false
			if !noCompression {
				// This is a simplified approach - in a real implementation,
				// we might want to store compression info separately
				compressed = true
			}
		}
	}

	// Add file to manifest
	m.AddFile(relativePath, fileHash, !ignoreContents, compressed, originalSize, storedSize)

	return nil
}