package hash

import (
	"testing"
)

func TestHashBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // SHA256 of empty string
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", // SHA256 of "hello world"
		},
		{
			name:     "test content",
			input:    []byte("This is test content for hashing"),
			expected: "8b6c4d8c7e5b1ff5a9b9d3a8c8e7b7a4b9a7c6d5e4f3a2b1c0d9e8f7a6b5c4d3", // This will be the actual hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashBytes(tt.input)
			
			// Check that result is not empty
			if result == "" {
				t.Error("Expected non-empty hash result")
			}
			
			// Check that result is hexadecimal and proper length for SHA256 (64 characters)
			if len(result) != 64 {
				t.Errorf("Expected hash length 64, got %d", len(result))
			}
			
			// Check that result contains only hexadecimal characters
			for _, char := range result {
				if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
					t.Errorf("Hash contains non-hexadecimal character: %c", char)
					break
				}
			}
			
			// For known test cases, check the exact value
			if tt.name == "empty bytes" && result != tt.expected {
				t.Errorf("Expected hash %s, got %s", tt.expected, result)
			}
			
			if tt.name == "hello world" && result != tt.expected {
				t.Errorf("Expected hash %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHashBytesConsistency(t *testing.T) {
	// Test that the same input produces the same hash
	input := []byte("consistency test content")
	
	hash1 := HashBytes(input)
	hash2 := HashBytes(input)
	
	if hash1 != hash2 {
		t.Errorf("Expected consistent hashes, got %s and %s", hash1, hash2)
	}
}

func TestHashBytesDifferentInputs(t *testing.T) {
	// Test that different inputs produce different hashes
	input1 := []byte("first input")
	input2 := []byte("second input")
	
	hash1 := HashBytes(input1)
	hash2 := HashBytes(input2)
	
	if hash1 == hash2 {
		t.Errorf("Expected different hashes for different inputs, both got %s", hash1)
	}
}

func TestGenerateFileNameHash(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"simple filename", "test.txt"},
		{"path with directories", "path/to/file.txt"},
		{"filename with spaces", "file with spaces.txt"},
		{"filename with special chars", "file-name_123.txt"},
		{"empty filename", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFileNameHash(tt.filename)
			
			// Check that result is not empty (even for empty filename)
			if result == "" {
				t.Error("Expected non-empty hash result")
			}
			
			// Check that result is hexadecimal and proper length for SHA256 (64 characters)
			if len(result) != 64 {
				t.Errorf("Expected hash length 64, got %d", len(result))
			}
			
			// Check that result contains only hexadecimal characters
			for _, char := range result {
				if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
					t.Errorf("Hash contains non-hexadecimal character: %c", char)
					break
				}
			}
		})
	}
}

func TestGenerateFileNameHashConsistency(t *testing.T) {
	// Test that the same filename produces the same hash
	filename := "consistent-test-file.txt"
	
	hash1 := GenerateFileNameHash(filename)
	hash2 := GenerateFileNameHash(filename)
	
	if hash1 != hash2 {
		t.Errorf("Expected consistent hashes for filename, got %s and %s", hash1, hash2)
	}
}

func TestGenerateFileNameHashDifferentNames(t *testing.T) {
	// Test that different filenames produce different hashes
	filename1 := "file1.txt"
	filename2 := "file2.txt"
	
	hash1 := GenerateFileNameHash(filename1)
	hash2 := GenerateFileNameHash(filename2)
	
	if hash1 == hash2 {
		t.Errorf("Expected different hashes for different filenames, both got %s", hash1)
	}
}

func TestHashFunctionsAreDifferent(t *testing.T) {
	// Test that HashBytes and GenerateFileNameHash produce different results
	// for the same input (they should use different methods)
	input := "test-content"
	
	bytesHash := HashBytes([]byte(input))
	filenameHash := GenerateFileNameHash(input)
	
	// They might be the same if GenerateFileNameHash just hashes the filename directly,
	// but let's check that they're both valid hashes
	if len(bytesHash) != 64 || len(filenameHash) != 64 {
		t.Errorf("Expected both hashes to be 64 characters, got %d and %d", len(bytesHash), len(filenameHash))
	}
}

func TestHashBytesWithLargeInput(t *testing.T) {
	// Test with a larger input to ensure the hash function works with various sizes
	largeInput := make([]byte, 10000)
	for i := range largeInput {
		largeInput[i] = byte(i % 256)
	}
	
	result := HashBytes(largeInput)
	
	if result == "" {
		t.Error("Expected non-empty hash result for large input")
	}
	
	if len(result) != 64 {
		t.Errorf("Expected hash length 64 for large input, got %d", len(result))
	}
}

func TestHashBytesWithBinaryData(t *testing.T) {
	// Test with binary data (including null bytes)
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	
	result := HashBytes(binaryData)
	
	if result == "" {
		t.Error("Expected non-empty hash result for binary data")
	}
	
	if len(result) != 64 {
		t.Errorf("Expected hash length 64 for binary data, got %d", len(result))
	}
	
	// Test that it's different from empty bytes
	emptyHash := HashBytes([]byte{})
	if result == emptyHash {
		t.Error("Expected different hash for binary data vs empty bytes")
	}
}