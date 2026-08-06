package runner

import (
	"os"
	"path/filepath"
)

func WriteFileAtomic(stageDir, path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		return err
	}

	// The leading dot and the absence of a .yaml/.yml suffix keep this file out
	// of the extension-filtered listings the EventProcessor builds.
	tmp, err := os.CreateTemp(stageDir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
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

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		return WriteFileAtomic(filepath.Dir(targetPath), targetPath, data, info.Mode())
	})
}

func DeleteFilesInDir(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})
}

func ClearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0755)
		}
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
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
