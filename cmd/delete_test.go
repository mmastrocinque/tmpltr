package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"tmpltr/internal/manifest"
	"tmpltr/internal/storage"
)

func TestValidateDeleteTemplateName(t *testing.T) {
	tests := []struct {
		name        string
		templateName string
		shouldError bool
	}{
		{"valid name", "my-template", false},
		{"empty name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeleteTemplateName(tt.templateName)
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for template name '%s', but got none", tt.templateName)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for template name '%s', but got: %v", tt.templateName, err)
			}
		})
	}
}

func TestConfirmDeletion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"yes", "yes\n", true},
		{"YES", "YES\n", true},
		{"Yes", "Yes\n", true},
		{"no", "no\n", false},
		{"empty", "\n", false},
		{"other", "maybe\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader with the test input
			oldStdin := os.Stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdin = r

			// Write the input
			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			// Capture stdout to avoid cluttering test output
			oldStdout := os.Stdout
			r2, w2, _ := os.Pipe()
			os.Stdout = w2

			result, err := confirmDeletion("test-template")

			// Restore stdin and stdout
			os.Stdin = oldStdin
			os.Stdout = oldStdout
			w2.Close()
			
			// Read the captured output (and discard it)
			_, _ = io.ReadAll(r2)
			r2.Close()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDeleteCommandFlags(t *testing.T) {
	// Test that required flags are properly configured
	nameFlag := deleteCmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("Expected 'name' flag to be defined")
	}

	forceFlag := deleteCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("Expected 'force' flag to be defined")
	}

	// Check flag shortcuts
	if nameFlag.Shorthand != "n" {
		t.Errorf("Expected 'name' flag shorthand to be 'n', got '%s'", nameFlag.Shorthand)
	}

	if forceFlag.Shorthand != "f" {
		t.Errorf("Expected 'force' flag shorthand to be 'f', got '%s'", forceFlag.Shorthand)
	}
}

func TestDeleteCommandCompletion(t *testing.T) {
	// Check that the delete command has completion registered for the name flag
	// This is a bit tricky to test directly, but we can verify the setup
	if deleteCmd.Flag("name") == nil {
		t.Error("Expected name flag to be configured")
	}

	// The completion function should be registered - this is tested more thoroughly
	// in the completion tests, but we can at least verify the command structure
	if deleteCmd.Use != "delete" {
		t.Errorf("Expected command use to be 'delete', got '%s'", deleteCmd.Use)
	}

	if deleteCmd.Short != "Delete a saved template" {
		t.Errorf("Expected short description to be 'Delete a saved template', got '%s'", deleteCmd.Short)
	}
}

func TestDeleteCommandIntegration(t *testing.T) {
	// Create temporary storage directory
	storageDir, err := os.MkdirTemp("", "tmpltr-delete-integration-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	// Create storage and template
	storage, err := storage.NewStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	testTemplateName := "test-delete-template"

	// Create template
	err = storage.EnsureTemplateDir(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Add a test file
	testHash := "testhash123"
	testContent := []byte("test content for deletion")
	err = storage.SaveFile(testTemplateName, testHash, testContent)
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	// Create and save manifest
	m := manifest.NewManifest(testTemplateName)
	m.AddFile("test.txt", testHash, true, false, int64(len(testContent)), int64(len(testContent)))
	err = storage.SaveManifest(testTemplateName, m)
	if err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Verify template exists
	if !storage.TemplateExists(testTemplateName) {
		t.Fatal("Expected template to exist before deletion test")
	}

	// Test the deletion logic components
	// Note: Testing runDelete directly is complex due to its dependencies on global variables
	// and user input. In a production system, you'd want to refactor this to be more testable.

	// Test template existence check
	exists := storage.TemplateExists(testTemplateName)
	if !exists {
		t.Error("Expected template to exist")
	}

	// Test loading manifest (this is what runDelete does)
	loadedManifest, err := storage.LoadManifest(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if loadedManifest.Name != testTemplateName {
		t.Errorf("Expected manifest name %s, got %s", testTemplateName, loadedManifest.Name)
	}

	if loadedManifest.GetFileCount() != 1 {
		t.Errorf("Expected 1 file in manifest, got %d", loadedManifest.GetFileCount())
	}

	// Test actual deletion
	err = storage.DeleteTemplate(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to delete template: %v", err)
	}

	// Verify template no longer exists
	if storage.TemplateExists(testTemplateName) {
		t.Error("Expected template not to exist after deletion")
	}
}

func TestRunDeleteWithMissingTemplate(t *testing.T) {
	// Create temporary storage directory
	storageDir, err := os.MkdirTemp("", "tmpltr-delete-missing-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	// Set up globals for the test
	originalDeleteTemplateName := deleteTemplateName
	originalForceDelete := forceDelete
	defer func() {
		deleteTemplateName = originalDeleteTemplateName
		forceDelete = originalForceDelete
	}()

	deleteTemplateName = "non-existent-template"
	forceDelete = true

	// Capture stderr to check error output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run the delete command (this should fail)
	err = runDelete(deleteCmd, []string{})

	// Restore stderr
	os.Stderr = oldStderr
	w.Close()
	
	// Read the error output
	output, _ := io.ReadAll(r)
	r.Close()

	// Should get an error for non-existent template
	if err == nil {
		t.Error("Expected error when deleting non-existent template")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error message to contain 'does not exist', got: %v", err)
	}

	// Clean up
	_ = output // Prevent unused variable error
}

func TestRunDeleteValidation(t *testing.T) {
	// Test empty template name validation
	originalDeleteTemplateName := deleteTemplateName
	defer func() {
		deleteTemplateName = originalDeleteTemplateName
	}()

	deleteTemplateName = ""

	err := runDelete(deleteCmd, []string{})
	if err == nil {
		t.Error("Expected error for empty template name")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected error message about empty name, got: %v", err)
	}
}

func TestDeleteCommandOutput(t *testing.T) {
	// Create temporary storage directory
	storageDir, err := os.MkdirTemp("", "tmpltr-delete-output-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	// Create storage and template
	storage, err := storage.NewStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	testTemplateName := "test-output-template"

	// Create template with manifest
	err = storage.EnsureTemplateDir(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	m := manifest.NewManifest(testTemplateName)
	m.AddFile("test.txt", "hash", true, false, 100, 100)
	err = storage.SaveManifest(testTemplateName, m)
	if err != nil {
		t.Fatalf("Failed to save manifest: %v", err)
	}

	// Set up globals
	originalDeleteTemplateName := deleteTemplateName
	originalForceDelete := forceDelete
	defer func() {
		deleteTemplateName = originalDeleteTemplateName
		forceDelete = originalForceDelete
	}()

	deleteTemplateName = testTemplateName
	forceDelete = true // Skip confirmation

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runDelete(deleteCmd, []string{})

	// Restore stdout
	os.Stdout = oldStdout
	w.Close()

	// Read output
	output, _ := io.ReadAll(r)
	r.Close()

	if err != nil {
		t.Fatalf("Unexpected error during deletion: %v", err)
	}

	outputStr := string(output)

	// Check that output contains expected information
	if !strings.Contains(outputStr, testTemplateName) {
		t.Errorf("Expected output to contain template name '%s'", testTemplateName)
	}

	if !strings.Contains(outputStr, "Successfully deleted") {
		t.Error("Expected output to contain success message")
	}

	// Verify template was actually deleted
	if storage.TemplateExists(testTemplateName) {
		t.Error("Expected template to be deleted")
	}
}