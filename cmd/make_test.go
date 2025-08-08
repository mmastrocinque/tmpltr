package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"tmpltr/internal/storage"
)

func TestValidateTargetDirectory(t *testing.T) {
	// Test with valid directory
	tempDir, err := os.MkdirTemp("", "tmpltr-make-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = validateTargetDirectory(tempDir)
	if err != nil {
		t.Errorf("Expected no error for valid directory, got: %v", err)
	}

	// Test with non-existent directory
	err = validateTargetDirectory("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Test with file instead of directory
	tempFile, err := os.CreateTemp("", "test-file-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	err = validateTargetDirectory(tempFile.Name())
	if err == nil {
		t.Error("Expected error when target is a file, not directory")
	}
}

func TestValidateTemplateName(t *testing.T) {
	tests := []struct {
		name        string
		templateName string
		shouldError bool
	}{
		{"valid name", "my-template", false},
		{"valid name with numbers", "template123", false},
		{"valid name with underscores", "my_template", false},
		{"empty name", "", true},
		{"name with slash", "template/name", true},
		{"name with backslash", "template\\name", true},
		{"name with colon", "template:name", true},
		{"name with asterisk", "template*", true},
		{"name with question mark", "template?", true},
		{"name with quotes", "template\"name", true},
		{"name with angle brackets", "template<name>", true},
		{"name with pipe", "template|name", true},
		{"name starting with dot", ".template", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTemplateName(tt.templateName)
			if tt.shouldError && err == nil {
				t.Errorf("Expected error for template name '%s', but got none", tt.templateName)
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for template name '%s', but got: %v", tt.templateName, err)
			}
		})
	}
}

func TestMakeCommandFlags(t *testing.T) {
	// Test that required flags are properly configured
	nameFlag := makeCmd.Flags().Lookup("name")
	if nameFlag == nil {
		t.Error("Expected 'name' flag to be defined")
	}

	// The name flag should be marked as required
	// This is validated when the command is executed, so we just check the flag exists
	if nameFlag.Usage != "Name for the template (required)" {
		t.Logf("Name flag usage: %s", nameFlag.Usage)
	}

	// Test other flags exist
	expectedFlags := []string{"ignore-contents", "ignore-files", "no-compression"}
	for _, flagName := range expectedFlags {
		flag := makeCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected '%s' flag to be defined", flagName)
		}
	}
}

func TestMakeCommandArgs(t *testing.T) {
	// Test that command expects exactly 1 argument
	// This is a bit hard to test directly, but we can check the Args field
	if makeCmd.Args == nil {
		t.Error("Expected make command to have Args validation")
	}

	// Test with correct number of args (this tests the validation indirectly)
	// We can't easily test this without running the command, but we can verify
	// the setup is correct by checking the command configuration
	if makeCmd.Use != "make <target_directory>" {
		t.Errorf("Expected make command usage to be 'make <target_directory>', got '%s'", makeCmd.Use)
	}
}

func TestProcessFileLogic(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "tmpltr-process-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "This is test content for processing"
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create storage for testing
	storageDir, err := os.MkdirTemp("", "tmpltr-storage-test-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	storage, err := storage.NewStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	templateName = "test-template"
	err = storage.EnsureTemplateDir(templateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Test the internal manifest creation and file processing
	// Note: This is testing the logic, not the actual processFile function
	// since it's not exported. In a real scenario, you might want to make
	// these functions testable by extracting them or making them public.
	
	// We can test the overall behavior by checking the make command
	// produces expected results, but that would be more of an integration test.
	
	t.Log("processFile function logic test - this would require refactoring for proper unit testing")
	// This test demonstrates where you might want to refactor the code
	// to make individual functions testable
}

func TestMakeCommandIntegration(t *testing.T) {
	// Create a temporary source directory
	sourceDir, err := os.MkdirTemp("", "tmpltr-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// Create test files in source directory
	testFiles := map[string]string{
		"file1.txt": "Content of file 1",
		"file2.txt": "Content of file 2",
		"subdir/file3.txt": "Content of file 3 in subdirectory",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(sourceDir, filePath)
		
		// Create directory if needed
		dir := filepath.Dir(fullPath)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", fullPath, err)
		}
	}

	// Create temporary storage directory
	storageDir, err := os.MkdirTemp("", "tmpltr-storage-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	// Set up storage
	storage, err := storage.NewStorage(storageDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test template creation by simulating the make command
	// Note: This is more of an integration test as it tests the whole flow
	testTemplateName := "integration-test-template"

	// Set global variables that the make command uses
	templateName = testTemplateName
	ignoreContents = false
	noCompression = false
	ignoreFiles = []string{}

	// This would ideally call runMake, but since it's hard to test directly
	// due to the way it's structured, we test the components
	
	err = storage.EnsureTemplateDir(testTemplateName)
	if err != nil {
		t.Fatalf("Failed to ensure template directory: %v", err)
	}

	// Verify template directory was created
	if !storage.TemplateExists(testTemplateName) {
		t.Error("Expected template to exist after creation")
	}

	// Clean up globals
	templateName = ""
	ignoreContents = false
	noCompression = false
	ignoreFiles = nil
}