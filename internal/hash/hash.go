package hash

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// HashFile calculates the SHA256 hash of a file's contents
func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file %s for hashing: %w", filePath, err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// HashBytes calculates the SHA256 hash of a byte slice
func HashBytes(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// HashString calculates the SHA256 hash of a string
func HashString(s string) string {
	return HashBytes([]byte(s))
}

// GenerateFileNameHash generates a hash-based filename for ignore-contents mode
// This creates a deterministic hash based on the file path
func GenerateFileNameHash(filePath string) string {
	return HashString(filePath)
}

// ValidateHash checks if a hash string is a valid SHA256 hash (64 hex characters)
func ValidateHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	
	for _, char := range hash {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	
	return true
}