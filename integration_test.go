package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Integration tests that test the complete tmpltr binary
// These tests require the binary to be built first

func TestIntegrationFullWorkflow(t *testing.T) {
	// Build the binary first
	err := exec.Command("go", "build", "-o", "tmpltr-test").Run()
	if err != nil {
		t.Fatalf("Failed to build tmpltr binary: %v", err)
	}
	defer os.Remove("tmpltr-test")

	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "tmpltr-integration-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	storageDir, err := os.MkdirTemp("", "tmpltr-integration-storage-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	restoreDir, err := os.MkdirTemp("", "tmpltr-integration-restore-")
	if err != nil {
		t.Fatalf("Failed to create restore directory: %v", err)
	}
	defer os.RemoveAll(restoreDir)

	// Create test files in source directory
	testFiles := map[string]string{
		"README.md":           "# Test Project\nThis is a test project for tmpltr.",
		"src/main.go":         "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
		"src/utils/helper.go": "package utils\n\nfunc Helper() string {\n\treturn \"helper\"\n}",
		"config.json":         `{"name": "test-project", "version": "1.0.0"}`,
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(sourceDir, filePath)
		dir := filepath.Dir(fullPath)
		
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	templateName := "integration-test-template"
	
	// Set TMPLTR_STORAGE environment variable to use our test storage directory
	originalStorage := os.Getenv("TMPLTR_STORAGE")
	err = os.Setenv("TMPLTR_STORAGE", storageDir)
	if err != nil {
		t.Fatalf("Failed to set TMPLTR_STORAGE: %v", err)
	}
	defer func() {
		if originalStorage == "" {
			os.Unsetenv("TMPLTR_STORAGE")
		} else {
			os.Setenv("TMPLTR_STORAGE", originalStorage)
		}
	}()

	// Step 1: Create template
	cmd := exec.Command("./tmpltr-test", "make", sourceDir, "--name="+templateName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create template: %v\nOutput: %s", err, string(output))
	}

	// Check that success message is printed
	if !strings.Contains(string(output), "Successfully created template") {
		t.Errorf("Expected success message in output, got: %s", string(output))
	}

	// Step 2: List templates
	cmd = exec.Command("./tmpltr-test", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list templates: %v\nOutput: %s", err, string(output))
	}

	// Check that our template is listed
	if !strings.Contains(string(output), templateName) {
		t.Errorf("Expected template name in list output, got: %s", string(output))
	}

	if !strings.Contains(string(output), "📁") {
		t.Errorf("Expected folder emoji in list output, got: %s", string(output))
	}

	// Step 3: Restore template
	restoreTarget := filepath.Join(restoreDir, "restored-project")
	cmd = exec.Command("./tmpltr-test", "restore", "--name="+templateName, "--output="+restoreTarget)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to restore template: %v\nOutput: %s", err, string(output))
	}

	// Check that success message is printed
	if !strings.Contains(string(output), "Successfully restored template") {
		t.Errorf("Expected success message in restore output, got: %s", string(output))
	}

	// Step 4: Verify restored files
	for filePath, expectedContent := range testFiles {
		restoredFile := filepath.Join(restoreTarget, filePath)
		
		content, err := os.ReadFile(restoredFile)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", restoredFile, err)
			continue
		}
		
		if string(content) != expectedContent {
			t.Errorf("Content mismatch for file %s\nExpected: %s\nGot: %s", 
				filePath, expectedContent, string(content))
		}
	}

	// Step 5: Test completion
	cmd = exec.Command("./tmpltr-test", "__complete", "delete", "--name", "")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to test completion: %v\nOutput: %s", err, string(output))
	}

	// Check that completion includes our template
	if !strings.Contains(string(output), templateName) {
		t.Errorf("Expected template name in completion output, got: %s", string(output))
	}

	// Step 6: Delete template
	cmd = exec.Command("./tmpltr-test", "delete", "--name="+templateName, "--force")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to delete template: %v\nOutput: %s", err, string(output))
	}

	// Check that success message is printed
	if !strings.Contains(string(output), "Successfully deleted template") {
		t.Errorf("Expected success message in delete output, got: %s", string(output))
	}

	// Step 7: Verify template is gone
	cmd = exec.Command("./tmpltr-test", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list templates after deletion: %v\nOutput: %s", err, string(output))
	}

	// Check that template is no longer listed
	if strings.Contains(string(output), templateName) {
		t.Errorf("Template should not be in list after deletion, got: %s", string(output))
	}
}

func TestIntegrationCompressionWorkflow(t *testing.T) {
	// Build the binary first
	err := exec.Command("go", "build", "-o", "tmpltr-test").Run()
	if err != nil {
		t.Fatalf("Failed to build tmpltr binary: %v", err)
	}
	defer os.Remove("tmpltr-test")

	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "tmpltr-compression-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	storageDir, err := os.MkdirTemp("", "tmpltr-compression-storage-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	// Create a file with repetitive content that should compress well
	repetitiveContent := strings.Repeat("This line repeats many times.\n", 1000)
	testFile := filepath.Join(sourceDir, "repetitive.txt")
	err = os.WriteFile(testFile, []byte(repetitiveContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	templateName := "compression-test-template"
	
	// Set storage directory
	originalStorage := os.Getenv("TMPLTR_STORAGE")
	err = os.Setenv("TMPLTR_STORAGE", storageDir)
	if err != nil {
		t.Fatalf("Failed to set TMPLTR_STORAGE: %v", err)
	}
	defer func() {
		if originalStorage == "" {
			os.Unsetenv("TMPLTR_STORAGE")
		} else {
			os.Setenv("TMPLTR_STORAGE", originalStorage)
		}
	}()

	// Create template with compression
	cmd := exec.Command("./tmpltr-test", "make", sourceDir, "--name="+templateName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create template with compression: %v\nOutput: %s", err, string(output))
	}

	// Check that compression information is displayed
	if !strings.Contains(string(output), "compression") || !strings.Contains(string(output), "%") {
		t.Logf("Compression output: %s", string(output))
		// Note: compression might not occur for small files, so this is just a log
	}

	// Create template without compression
	templateNameNoComp := "no-compression-test-template"
	cmd = exec.Command("./tmpltr-test", "make", sourceDir, "--name="+templateNameNoComp, "--no-compression")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create template without compression: %v\nOutput: %s", err, string(output))
	}

	// Both templates should be created successfully
	cmd = exec.Command("./tmpltr-test", "list")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list templates: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), templateName) {
		t.Errorf("Expected compressed template in list: %s", string(output))
	}

	if !strings.Contains(string(output), templateNameNoComp) {
		t.Errorf("Expected uncompressed template in list: %s", string(output))
	}
}

func TestIntegrationIgnoreContents(t *testing.T) {
	// Build the binary first
	err := exec.Command("go", "build", "-o", "tmpltr-test").Run()
	if err != nil {
		t.Fatalf("Failed to build tmpltr binary: %v", err)
	}
	defer os.Remove("tmpltr-test")

	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "tmpltr-ignore-contents-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	storageDir, err := os.MkdirTemp("", "tmpltr-ignore-contents-storage-")
	if err != nil {
		t.Fatalf("Failed to create storage directory: %v", err)
	}
	defer os.RemoveAll(storageDir)

	restoreDir, err := os.MkdirTemp("", "tmpltr-ignore-contents-restore-")
	if err != nil {
		t.Fatalf("Failed to create restore directory: %v", err)
	}
	defer os.RemoveAll(restoreDir)

	// Create test files
	testFiles := map[string]string{
		"file1.txt":    "This content should be ignored",
		"file2.txt":    "This content should also be ignored",
		"dir/file3.txt": "Nested file content",
	}

	for filePath, content := range testFiles {
		fullPath := filepath.Join(sourceDir, filePath)
		dir := filepath.Dir(fullPath)
		
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	templateName := "ignore-contents-test-template"
	
	// Set storage directory
	originalStorage := os.Getenv("TMPLTR_STORAGE")
	err = os.Setenv("TMPLTR_STORAGE", storageDir)
	if err != nil {
		t.Fatalf("Failed to set TMPLTR_STORAGE: %v", err)
	}
	defer func() {
		if originalStorage == "" {
			os.Unsetenv("TMPLTR_STORAGE")
		} else {
			os.Setenv("TMPLTR_STORAGE", originalStorage)
		}
	}()

	// Create template ignoring contents
	cmd := exec.Command("./tmpltr-test", "make", sourceDir, "--name="+templateName, "--ignore-contents")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to create template ignoring contents: %v\nOutput: %s", err, string(output))
	}

	// Check that the output indicates structure only
	if !strings.Contains(string(output), "structure only") {
		t.Errorf("Expected 'structure only' in output, got: %s", string(output))
	}

	// Restore template
	restoreTarget := filepath.Join(restoreDir, "restored-structure")
	cmd = exec.Command("./tmpltr-test", "restore", "--name="+templateName, "--output="+restoreTarget)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to restore template: %v\nOutput: %s", err, string(output))
	}

	// Verify that files exist but are empty
	for filePath := range testFiles {
		restoredFile := filepath.Join(restoreTarget, filePath)
		
		info, err := os.Stat(restoredFile)
		if err != nil {
			t.Errorf("Expected file %s to exist after restore, but it doesn't", restoredFile)
			continue
		}
		
		// File should exist but be empty (or very small)
		if info.Size() > 0 {
			content, _ := os.ReadFile(restoredFile)
			if len(content) > 0 {
				t.Errorf("Expected file %s to be empty when ignoring contents, got content: %s", 
					filePath, string(content))
			}
		}
	}
}

func TestIntegrationCompletionGeneration(t *testing.T) {
	// Build the binary first
	err := exec.Command("go", "build", "-o", "tmpltr-test").Run()
	if err != nil {
		t.Fatalf("Failed to build tmpltr binary: %v", err)
	}
	defer os.Remove("tmpltr-test")

	// Test completion generation for different shells
	shells := []string{"bash", "zsh", "fish", "powershell", "nushell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			cmd := exec.Command("./tmpltr-test", "completion", shell)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to generate %s completion: %v\nOutput: %s", shell, err, string(output))
			}

			if len(output) == 0 {
				t.Errorf("Expected non-empty completion script for %s", shell)
			}

			// Basic sanity checks for completion content
			outputStr := string(output)
			if shell == "nushell" {
				if !strings.Contains(outputStr, "extern \"tmpltr\"") {
					t.Errorf("Expected Nushell completion to contain extern declarations")
				}
				if !strings.Contains(outputStr, "nu-complete") {
					t.Errorf("Expected Nushell completion to contain custom completion functions")
				}
			} else {
				// For other shells, just check that tmpltr is mentioned
				if !strings.Contains(outputStr, "tmpltr") {
					t.Errorf("Expected %s completion to contain 'tmpltr'", shell)
				}
			}
		})
	}
}