package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

func runningExecutableSHA256() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	before, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !before.Mode().IsRegular() {
		return "", errors.New("running executable path is not a regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	opened, err := f.Stat()
	if err != nil || !os.SameFile(before, opened) {
		f.Close()
		return "", errors.New("running executable changed before hashing")
	}
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err := errors.Join(err, f.Close()); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
