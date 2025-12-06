package utils

import (
	"bytes"
	"fmt"
	"github.com/andybalholm/brotli"
)

const (
	BestSpeed          = brotli.BestSpeed
	BestCompression    = brotli.BestCompression
	DefaultCompression = brotli.DefaultCompression
)

func Compress(content []byte, compressionLevel int) ([]byte, error) {
	var compressed bytes.Buffer
	writer := brotli.NewWriterLevel(&compressed, compressionLevel)

	if _, err := writer.Write(content); err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close compressor: %w", err)
	}

	return compressed.Bytes(), nil
}
