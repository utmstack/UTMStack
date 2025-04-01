package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func MoveFolderContents(srcDir, destDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		if entry.Type().IsRegular() {
			srcPath := filepath.Join(srcDir, entry.Name())
			destPath := filepath.Join(destDir, entry.Name())

			if err := os.Rename(srcPath, destPath); err != nil {
				return fmt.Errorf("error moving file from %s to %s: %w", srcPath, destPath, err)
			}
		}
	}

	return nil
}
