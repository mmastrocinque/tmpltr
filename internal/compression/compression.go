package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// CompressData compresses data using gzip compression
func CompressData(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write(data)
	if err != nil {
		gzipWriter.Close()
		return nil, fmt.Errorf("failed to write data to gzip writer: %w", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressData decompresses gzip-compressed data
func DecompressData(compressedData []byte) ([]byte, error) {
	if len(compressedData) == 0 {
		return compressedData, nil
	}

	reader := bytes.NewReader(compressedData)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	return buf.Bytes(), nil
}

// ShouldCompress determines if a file should be compressed based on size and type
func ShouldCompress(data []byte, filePath string) bool {
	// Don't compress very small files (less than 100 bytes)
	if len(data) < 100 {
		return false
	}

	// Check if file is already compressed based on extension
	compressedExtensions := map[string]bool{
		".gz":   true,
		".zip":  true,
		".7z":   true,
		".rar":  true,
		".bz2":  true,
		".xz":   true,
		".tar":  false, // tar files can benefit from compression
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".mp3":  true,
		".mp4":  true,
		".avi":  true,
		".mkv":  true,
		".pdf":  true,
	}

	// Get file extension
	lastDot := -1
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			lastDot = i
			break
		}
		if filePath[i] == '/' || filePath[i] == '\\' {
			break
		}
	}

	if lastDot >= 0 {
		ext := filePath[lastDot:]
		if isCompressed, exists := compressedExtensions[ext]; exists {
			return !isCompressed
		}
	}

	// For unknown file types, compress if file is large enough
	return len(data) >= 100
}

// GetCompressionRatio calculates the compression ratio
func GetCompressionRatio(originalSize, compressedSize int) float64 {
	if originalSize == 0 {
		return 0
	}
	return float64(compressedSize) / float64(originalSize)
}