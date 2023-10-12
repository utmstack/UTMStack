package utils

import (
	"fmt"
	"os"
)

func GetLock(v int, locksDir string) bool {
	_, err := os.ReadFile(fmt.Sprintf("%s/%d.lock", locksDir, v))

	return err != nil
}

func SetLock(v int, locksDir string) error {
	if err := os.WriteFile(fmt.Sprintf("%s/%d.lock", locksDir, v), []byte(fmt.Sprintf("%d", v)), 0644); err != nil {
		return err
	}

	return nil
}
