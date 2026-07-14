package runner

import (
	"io"
	"os"
	"path/filepath"
)

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(targetPath, info.Mode())
	})
}

func DeleteFilesInDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		return os.Remove(path)
	})
}

func ClearDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}

func TruncateFile(path string) error {
	err := os.Truncate(path, 0)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func EnsureDirs(workDir string) error {
	subtrees := [][]string{
		{workDir, "input"},
		{workDir, "output"},
		{workDir, "pipeline", "filters", "system"},
		{workDir, "pipeline", "filters", "user"},
		{workDir, "rules", "system"},
		{workDir, "rules", "user"},
		{workDir, "plugins"},
		{workDir, "sockets"},
		{workDir, "geolocation"},
	}
	for _, parts := range subtrees {
		path := filepath.Join(parts...)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}
	return nil
}
