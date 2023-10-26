package utils

import (
	"fmt"
	"os"
	"strconv"
)

func GetLock(v int, locksDir string) bool {
	if GetStep() != 0 {
		for i := 1; i <= 10; i++ {
			SetLock(i, locksDir)
		}
		Remove("/root/step.txt")
	}

	_, err := os.ReadFile(fmt.Sprintf("%s/%d.lock", locksDir, v))

	return err != nil
}

func SetLock(v int, locksDir string) error {
	if err := os.WriteFile(fmt.Sprintf("%s/%d.lock", locksDir, v), []byte(fmt.Sprintf("%d", v)), 0644); err != nil {
		return err
	}

	return nil
}

func GetStep() int {
	v, err := os.ReadFile("/root/step.txt")
	if err != nil {
		return 0
	}

	vi, err := strconv.Atoi(string(v))
	if err != nil {
		return 0
	}

	return vi
}
