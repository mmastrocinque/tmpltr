package ignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const IgnoreFileName = ".tmpltrignore"

// IgnoreRules holds the rules for ignoring files and directories
type IgnoreRules struct {
	patterns []string
	rootDir  string
}

// NewIgnoreRules creates a new IgnoreRules instance
func NewIgnoreRules(rootDir string) *IgnoreRules {
	return &IgnoreRules{
		patterns: make([]string, 0),
		rootDir:  rootDir,
	}
}

// LoadIgnoreFile loads ignore patterns from .tmpltrignore file
func (ir *IgnoreRules) LoadIgnoreFile() error {
	ignoreFilePath := filepath.Join(ir.rootDir, IgnoreFileName)
	
	file, err := os.Open(ignoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open ignore file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		ir.patterns = append(ir.patterns, line)
	}

	return scanner.Err()
}

// AddPattern adds a custom ignore pattern
func (ir *IgnoreRules) AddPattern(pattern string) {
	if pattern != "" {
		ir.patterns = append(ir.patterns, pattern)
	}
}

// AddPatterns adds multiple patterns from a slice
func (ir *IgnoreRules) AddPatterns(patterns []string) {
	for _, pattern := range patterns {
		ir.AddPattern(pattern)
	}
}

// AddDefaultPatterns adds common default ignore patterns
func (ir *IgnoreRules) AddDefaultPatterns() {
	defaultPatterns := []string{
		".git/",
		".svn/",
		".hg/",
		"node_modules/",
		".DS_Store",
		"Thumbs.db",
		"*.tmp",
		"*.temp",
		"*.log",
		"*.swp",
		"*.swo",
		"*~",
		".tmpltrignore",
	}
	
	ir.AddPatterns(defaultPatterns)
}

// ShouldIgnore checks if a file/directory should be ignored based on the rules
func (ir *IgnoreRules) ShouldIgnore(filePath string) bool {
	relPath, err := filepath.Rel(ir.rootDir, filePath)
	if err != nil {
		return false
	}

	relPath = filepath.ToSlash(relPath)
	
	for _, pattern := range ir.patterns {
		if ir.matchesPattern(relPath, pattern) {
			return true
		}
	}
	
	return false
}

// matchesPattern checks if a file path matches an ignore pattern
func (ir *IgnoreRules) matchesPattern(filePath, pattern string) bool {
	pattern = filepath.ToSlash(pattern)
	
	// Handle directory patterns (ending with /)
	if strings.HasSuffix(pattern, "/") {
		dirPattern := strings.TrimSuffix(pattern, "/")
		
		if filePath == dirPattern {
			return true
		}
		
		if strings.HasPrefix(filePath, dirPattern+"/") {
			return true
		}
		
		return false
	}
	
	// Handle exact matches
	if filePath == pattern {
		return true
	}
	
	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}
		
		if strings.Contains(pattern, "/") {
			matched, err := filepath.Match(pattern, filePath)
			if err == nil && matched {
				return true
			}
		}
	}
	
	// Handle path prefix patterns
	if strings.Contains(pattern, "/") {
		if strings.HasPrefix(filePath, pattern) {
			return true
		}
	} else {
		// Pattern is just a filename, check if any part of path matches
		pathParts := strings.Split(filePath, "/")
		for _, part := range pathParts {
			if part == pattern {
				return true
			}
			if strings.Contains(pattern, "*") {
				matched, err := filepath.Match(pattern, part)
				if err == nil && matched {
					return true
				}
			}
		}
	}
	
	return false
}

// GetPatterns returns all loaded patterns
func (ir *IgnoreRules) GetPatterns() []string {
	return ir.patterns
}