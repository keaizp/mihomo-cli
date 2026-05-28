package kernel

import (
	"compress/gzip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ExtractEmbedded writes the embedded mihomo kernel to destPath.
// The embedded data is gzip-compressed; it decompresses on the fly.
func (m *Manager) ExtractEmbedded(destPath string) error {
	data, err := embeddedKernel()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decompress embedded kernel: %w", err)
	}
	defer gzReader.Close()

	tmpPath := destPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(f, gzReader); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write kernel: %w", err)
	}
	f.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("chmod: %w", err)
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}
