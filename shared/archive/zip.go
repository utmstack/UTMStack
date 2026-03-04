// Package archive provides utilities for working with archive files.
package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip extracts a zip archive to the specified destination directory.
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if err := extractFile(f, dest); err != nil {
			return err
		}
	}
	return nil
}

func extractFile(f *zip.File, dest string) error {
	// Sanitize the file path to prevent zip slip attacks
	filePath := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
		return nil
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return fmt.Errorf("error creating parent directory: %w", err)
	}

	// Create the file
	outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer outFile.Close()

	// Open the file in the archive
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("error opening archived file: %w", err)
	}
	defer rc.Close()

	// Copy contents
	if _, err = io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("error extracting file: %w", err)
	}

	return nil
}
