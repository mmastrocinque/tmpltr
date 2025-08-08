package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"tmpltr/internal/manifest"
	"tmpltr/internal/storage"
)

func TestGenerateNushellCompletion(t *testing.T) {
	// Create a temporary file to capture output
	tmpFile, err := os.CreateTemp("", "nushell-completion-test-")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	err = generateNushellCompletion(tmpFile)
	if err != nil {
		t.Fatalf("Failed to generate Nushell completion: %v", err)
	}

	// Read the output
	tmpFile.Seek(0, 0)
	outputBytes, err := io.ReadAll(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read completion output: %v", err)
	}

	output := string(outputBytes)

	// Check that the output contains expected Nushell completion elements
	expectedStrings := []string{
		"# tmpltr completions for Nushell",
		"export extern \"tmpltr\"",
		"export extern \"tmpltr make\"",
		"export extern \"tmpltr restore\"",
		"export extern \"tmpltr delete\"",
		"export extern \"tmpltr list\"",
		"export extern \"tmpltr completion\"",
		"nu-complete tmpltr template-names",
		"nu-complete tmpltr completion shells",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected Nushell completion to contain '%s', but it didn't", expected)
		}
	}

	// Check that completion shells are properly defined
	shells := []string{"bash", "zsh", "fish", "powershell", "nushell"}
	for _, shell := range shells {
		if !strings.Contains(output, shell) {
			t.Errorf("Expected Nushell completion to include shell '%s'", shell)
		}
	}
}

func TestTemplateNameCompletion(t *testing.T) {
	// Create a temporary directory for testing storage
	tempDir, err := os.MkdirTemp("", "tmpltr-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize storage in temp directory
	storage, err := storage.NewStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create test templates
	testTemplates := []struct {
		name      string
		fileCount int
	}{
		{"test-template-1", 5},
		{"test-template-2", 3},
		{"another-template", 1},
	}

	for _, template := range testTemplates {
		// Create template directory
		err := storage.EnsureTemplateDir(template.name)
		if err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		// Create manifest
		m := manifest.NewManifest(template.name)
		
		// Add some files to the manifest
		for i := 0; i < template.fileCount; i++ {
			filename := filepath.Join("test", "file"+string(rune('0'+i))+".txt")
			m.AddFile(filename, "hash"+string(rune('0'+i)), true, false, 100, 100)
		}

		// Save manifest
		err = storage.SaveManifest(template.name, m)
		if err != nil {
			t.Fatalf("Failed to save manifest: %v", err)
		}
	}

	// Test completion function
	completions, directive := templateNameCompletion(nil, []string{}, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected directive %v, got %v", cobra.ShellCompDirectiveNoFileComp, directive)
	}

	if len(completions) != len(testTemplates) {
		t.Errorf("Expected %d completions, got %d", len(testTemplates), len(completions))
	}

	// Check that completions contain expected template names and descriptions
	for _, completion := range completions {
		// Completion format should be: "template-name\tX files, created YYYY-MM-DD"
		parts := strings.Split(completion, "\t")
		if len(parts) < 2 {
			t.Errorf("Expected completion to have description, got: %s", completion)
			continue
		}

		templateName := parts[0]
		description := parts[1]

		// Find the expected template
		var expectedTemplate *struct {
			name      string
			fileCount int
		}
		for j := range testTemplates {
			if testTemplates[j].name == templateName {
				expectedTemplate = &testTemplates[j]
				break
			}
		}

		if expectedTemplate == nil {
			t.Errorf("Unexpected template name in completion: %s", templateName)
			continue
		}

		// Check that description contains file count
		expectedFileCount := string(rune('0' + expectedTemplate.fileCount))
		if !strings.Contains(description, expectedFileCount+" files") {
			t.Errorf("Expected description to contain '%s files', got: %s", expectedFileCount, description)
		}

		// Check that description contains creation date
		if !strings.Contains(description, "created") {
			t.Errorf("Expected description to contain 'created', got: %s", description)
		}
	}
}

func TestTemplateNameCompletionEmpty(t *testing.T) {
	// Test with empty storage directory
	tempDir, err := os.MkdirTemp("", "tmpltr-test-empty-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with empty storage directory
	// Note: This test shows the expected behavior with empty storage

	completions, directive := templateNameCompletion(nil, []string{}, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected directive %v, got %v", cobra.ShellCompDirectiveNoFileComp, directive)
	}

	// With no templates, should return empty slice
	if len(completions) > 0 {
		t.Errorf("Expected no completions for empty storage, got %d", len(completions))
	}
}

func TestCompletionCommand(t *testing.T) {
	// Test that completion command is properly configured
	if completionCmd.Use != "completion [bash|zsh|fish|powershell|nushell]" {
		t.Errorf("Unexpected completion command usage: %s", completionCmd.Use)
	}

	if completionCmd.Short != "Generate shell completion scripts" {
		t.Errorf("Unexpected completion command short description: %s", completionCmd.Short)
	}

	// Test valid args
	expectedShells := []string{"bash", "zsh", "fish", "powershell", "nushell"}
	if len(completionCmd.ValidArgs) != len(expectedShells) {
		t.Errorf("Expected %d valid args, got %d", len(expectedShells), len(completionCmd.ValidArgs))
	}

	for i, expected := range expectedShells {
		if i >= len(completionCmd.ValidArgs) || completionCmd.ValidArgs[i] != expected {
			if i < len(completionCmd.ValidArgs) {
				t.Errorf("Expected valid arg %d to be %s, got %s", i, expected, completionCmd.ValidArgs[i])
			} else {
				t.Errorf("Missing valid arg %d, expected %s", i, expected)
			}
		}
	}
}

func TestRunCompletion(t *testing.T) {
	tests := []struct {
		name        string
		shell       string
		shouldError bool
	}{
		{"bash", "bash", false},
		{"zsh", "zsh", false},
		{"fish", "fish", false},
		{"powershell", "powershell", false},
		{"nushell", "nushell", false},
		{"invalid", "invalid-shell", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for output
			tmpFile, err := os.CreateTemp("", "completion-test-")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())
			defer tmpFile.Close()

			// Backup original stdout and redirect
			originalStdout := os.Stdout
			os.Stdout = tmpFile

			err = runCompletion(completionCmd, []string{tt.shell})

			// Restore stdout
			os.Stdout = originalStdout

			if tt.shouldError && err == nil {
				t.Errorf("Expected error for shell %s, but got none", tt.shell)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error for shell %s, but got: %v", tt.shell, err)
			}

			if !tt.shouldError {
				// Check that something was written
				tmpFile.Seek(0, 0)
				content := make([]byte, 100)
				n, _ := tmpFile.Read(content)
				if n == 0 {
					t.Errorf("Expected completion output for shell %s, but got nothing", tt.shell)
				}
			}
		})
	}
}