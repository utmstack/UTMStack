// Package fs provides filesystem utility functions.
package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GetExecutablePath returns the directory path of the current executable.
func GetExecutablePath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(ex)
}

// Exists checks if a path exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirIfNotExist creates a directory and all parent directories if they don't exist.
func CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking path: %w", err)
	}
	return nil
}

// Copy copies a file from src to dst.
func Copy(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}
	return nil
}

// WriteString writes a string to a file, creating or truncating it.
func WriteString(path string, content string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}

// ReadLines reads a file and returns its lines as a slice.
func ReadLines(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lines []string
	var line []byte
	for _, b := range content {
		if b == '\n' {
			lines = append(lines, string(line))
			line = nil
		} else if b != '\r' {
			line = append(line, b)
		}
	}
	if len(line) > 0 {
		lines = append(lines, string(line))
	}
	return lines, nil
}

// IsEmpty checks if a directory is empty.
func IsEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
