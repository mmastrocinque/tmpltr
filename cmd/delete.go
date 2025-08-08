package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"tmpltr/internal/storage"
)

var (
	deleteTemplateName string
	forceDelete        bool
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a saved template",
	Long: `Delete a saved template and all its associated files from storage.
This action cannot be undone.

Example:
  tmpltr delete --name="my-template"
  tmpltr delete --name="my-template" --force`,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteTemplateName, "name", "n", "", "Name of the template to delete (required)")
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.MarkFlagRequired("name")
	
	// Add completion for template names
	deleteCmd.RegisterFlagCompletionFunc("name", templateNameCompletion)
}

// runDelete executes the delete command logic
func runDelete(cmd *cobra.Command, args []string) error {
	// Validate template name
	if err := validateDeleteTemplateName(deleteTemplateName); err != nil {
		return err
	}

	// Initialize storage
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check if template exists
	if !storage.TemplateExists(deleteTemplateName) {
		return fmt.Errorf("template '%s' does not exist", deleteTemplateName)
	}

	// Load manifest to get template information
	manifest, err := storage.LoadManifest(deleteTemplateName)
	if err != nil {
		return fmt.Errorf("failed to load template information: %w", err)
	}

	// Show template information
	fmt.Printf("Template to delete: %s\n", manifest.Name)
	fmt.Printf("Created: %s\n", manifest.CreatedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Files: %d\n", manifest.GetFileCount())

	// Confirm deletion unless force flag is used
	if !forceDelete {
		confirmed, err := confirmDeletion(deleteTemplateName)
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirmed {
			fmt.Println("Delete operation cancelled.")
			return nil
		}
	}

	// Perform deletion
	if err := storage.DeleteTemplate(deleteTemplateName); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("Successfully deleted template '%s'\n", deleteTemplateName)
	return nil
}

// validateDeleteTemplateName checks if the template name is valid for deletion
func validateDeleteTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("template name cannot be empty")
	}
	return nil
}

// confirmDeletion prompts the user to confirm template deletion
func confirmDeletion(templateName string) (bool, error) {
	fmt.Printf("\nAre you sure you want to delete template '%s'? This action cannot be undone.\n", templateName)
	fmt.Print("Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes", nil
}