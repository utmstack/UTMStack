package utils

import (
	"fmt"
	"os"
)

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
