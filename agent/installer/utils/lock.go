package utils

import (
	"fmt"
	"os"
)

func SetLock(lockdir string) error {
	if !CheckIfPathExist(lockdir) {
		file, err := os.OpenFile(lockdir, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func RemoveLock(lockdir string) error {
	if CheckIfPathExist(lockdir) {
		err := os.Remove(lockdir)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("lock file %s not exists", lockdir)
	}
	return nil
}
