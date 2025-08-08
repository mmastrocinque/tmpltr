package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"tmpltr/internal/manifest"
	"tmpltr/internal/storage"
)

var (
	restoreTemplateName string
	outputDirectory     string
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a template to a destination directory",
	Long: `Restore a saved template by recreating its directory structure and files
at the specified output location.

Example:
  tmpltr restore --name="my-template" --output="./restored-project"`,
	RunE: runRestore,
}

func init() {
	restoreCmd.Flags().StringVarP(&restoreTemplateName, "name", "n", "", "Name of the template to restore (required)")
	restoreCmd.Flags().StringVarP(&outputDirectory, "output", "o", "", "Output directory path (required)")
	restoreCmd.MarkFlagRequired("name")
	restoreCmd.MarkFlagRequired("output")
}

// runRestore executes the restore command logic
func runRestore(cmd *cobra.Command, args []string) error {
	// Validate template name
	if err := validateRestoreTemplateName(restoreTemplateName); err != nil {
		return err
	}

	// Validate output directory
	if err := validateOutputDirectory(outputDirectory); err != nil {
		return err
	}

	// Initialize storage
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check if template exists
	if !storage.TemplateExists(restoreTemplateName) {
		return fmt.Errorf("template '%s' does not exist", restoreTemplateName)
	}

	// Load manifest
	m, err := storage.LoadManifest(restoreTemplateName)
	if err != nil {
		return fmt.Errorf("failed to load template manifest: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Restore files
	restoredCount := 0
	for _, fileEntry := range m.Files {
		if err := restoreFile(fileEntry, outputDirectory, storage); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", fileEntry.OriginalPath, err)
		}
		restoredCount++
	}

	fmt.Printf("Successfully restored template '%s' with %d files to: %s\n", 
		restoreTemplateName, restoredCount, outputDirectory)
	
	contentFiles := len(m.GetFilesWithContents())
	emptyFiles := restoredCount - contentFiles
	if contentFiles > 0 {
		fmt.Printf("Restored %d files with content\n", contentFiles)
	}
	if emptyFiles > 0 {
		fmt.Printf("Created %d empty placeholder files\n", emptyFiles)
	}

	return nil
}

// validateRestoreTemplateName checks if the template name is valid for restoration
func validateRestoreTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	return nil
}

// validateOutputDirectory checks if the output directory path is valid
func validateOutputDirectory(outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Check if path exists and is not a file
	if info, err := os.Stat(outputDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("output path exists but is not a directory: %s", outputDir)
		}
		
		// Check if directory is empty
		entries, err := os.ReadDir(outputDir)
		if err != nil {
			return fmt.Errorf("failed to read output directory: %w", err)
		}
		
		if len(entries) > 0 {
			return fmt.Errorf("output directory is not empty: %s", outputDir)
		}
	}

	// Validate path components
	if strings.ContainsAny(outputDir, "*?\"<>|") {
		return fmt.Errorf("output directory path contains invalid characters")
	}

	return nil
}

// restoreFile restores a single file from the template storage
func restoreFile(fileEntry manifest.FileEntry, outputDir string, storage *storage.Storage) error {
	// Calculate target file path
	targetPath := filepath.Join(outputDir, fileEntry.OriginalPath)
	
	// Create parent directories if they don't exist
	parentDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", fileEntry.OriginalPath, err)
	}

	if fileEntry.IncludeContents {
		// Load file content from storage with decompression if needed
		content, err := storage.LoadFileWithDecompression(restoreTemplateName, fileEntry.Hash, fileEntry.Compressed)
		if err != nil {
			return fmt.Errorf("failed to load file content for %s: %w", fileEntry.OriginalPath, err)
		}

		// Write content to target file
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fileEntry.OriginalPath, err)
		}
	} else {
		// Create empty file for ignore-contents mode
		file, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create empty file %s: %w", fileEntry.OriginalPath, err)
		}
		file.Close()
	}

	return nil
}