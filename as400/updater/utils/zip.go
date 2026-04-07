package utils

import (
	"archive/zip"
	"io"
	"os"
	"path"
	"path/filepath"
)

func Unzip(zipFile, destPath string) error {
	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		err := func() error {
			filePath := path.Join(destPath, f.Name)
			if f.FileInfo().IsDir() {
				os.MkdirAll(filePath, os.ModePerm)
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				return err
			}

			dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer dstFile.Close()

			fileInArchive, err := f.Open()
			if err != nil {
				return err
			}
			defer fileInArchive.Close()

			if _, err := io.Copy(dstFile, fileInArchive); err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
