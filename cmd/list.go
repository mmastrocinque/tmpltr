package cmd

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"tmpltr/internal/storage"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved templates",
	Long: `Display all saved templates with their metadata including name, creation date,
and file count.

Example:
  tmpltr list`,
	RunE: runList,
}

// TemplateInfo holds template information for display
type TemplateInfo struct {
	Name            string
	CreatedAt       time.Time
	FileCount       int
	ContentFiles    int
	StructureFiles  int
	CompressedFiles int
	OriginalSize    int64
	StoredSize      int64
}

// runList executes the list command logic
func runList(cmd *cobra.Command, args []string) error {
	// Initialize storage
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Get list of templates
	templateNames, err := storage.ListTemplates()
	if err != nil {
		return fmt.Errorf("failed to list templates: %w", err)
	}

	if len(templateNames) == 0 {
		fmt.Println("No templates found.")
		fmt.Println("Create a template using: tmpltr make <directory> --name=\"template-name\"")
		return nil
	}

	// Collect template information
	var templates []TemplateInfo
	for _, name := range templateNames {
		info, err := getTemplateInfo(name, storage)
		if err != nil {
			fmt.Printf("Warning: failed to read template '%s': %v\n", name, err)
			continue
		}
		templates = append(templates, info)
	}

	if len(templates) == 0 {
		fmt.Println("No valid templates found.")
		return nil
	}

	// Sort templates by creation date (newest first)
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].CreatedAt.After(templates[j].CreatedAt)
	})

	// Display templates
	displayTemplates(templates)

	return nil
}

// getTemplateInfo retrieves information about a specific template
func getTemplateInfo(templateName string, storage *storage.Storage) (TemplateInfo, error) {
	manifest, err := storage.LoadManifest(templateName)
	if err != nil {
		return TemplateInfo{}, err
	}

	contentFiles := len(manifest.GetFilesWithContents())
	totalFiles := manifest.GetFileCount()
	structureFiles := totalFiles - contentFiles
	compressedFiles, originalSize, storedSize := manifest.GetCompressionStats()

	return TemplateInfo{
		Name:            manifest.Name,
		CreatedAt:       manifest.CreatedAt,
		FileCount:       totalFiles,
		ContentFiles:    contentFiles,
		StructureFiles:  structureFiles,
		CompressedFiles: compressedFiles,
		OriginalSize:    originalSize,
		StoredSize:      storedSize,
	}, nil
}

// displayTemplates formats and prints the template list
func displayTemplates(templates []TemplateInfo) {
	fmt.Printf("Found %d template(s):\n\n", len(templates))

	for i, template := range templates {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("📁 %s\n", template.Name)
		fmt.Printf("   Created: %s\n", template.CreatedAt.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("   Files:   %d total", template.FileCount)

		if template.ContentFiles > 0 && template.StructureFiles > 0 {
			fmt.Printf(" (%d with content, %d structure only)", template.ContentFiles, template.StructureFiles)
		} else if template.ContentFiles > 0 {
			fmt.Printf(" (all with content)")
		} else if template.StructureFiles > 0 {
			fmt.Printf(" (structure only)")
		}
		fmt.Println()

		// Show compression info if applicable
		if template.CompressedFiles > 0 {
			compressionRatio := (1.0 - float64(template.StoredSize)/float64(template.OriginalSize)) * 100
			fmt.Printf("   Storage: %.1f KB (%.1f%% compression, %d files compressed)\n",
				float64(template.StoredSize)/1024, compressionRatio, template.CompressedFiles)
		} else if template.ContentFiles > 0 {
			fmt.Printf("   Storage: %.1f KB (uncompressed)\n", float64(template.StoredSize)/1024)
		}

		// Show relative time
		now := time.Now()
		duration := now.Sub(template.CreatedAt)
		
		if duration < time.Hour {
			fmt.Printf("   Age:     %d minutes ago\n", int(duration.Minutes()))
		} else if duration < 24*time.Hour {
			fmt.Printf("   Age:     %d hours ago\n", int(duration.Hours()))
		} else {
			days := int(duration.Hours() / 24)
			if days == 1 {
				fmt.Printf("   Age:     1 day ago\n")
			} else {
				fmt.Printf("   Age:     %d days ago\n", days)
			}
		}
	}

	fmt.Printf("\nUse 'tmpltr restore --name=\"<template-name>\" --output=\"<directory>\"' to restore a template.\n")
}