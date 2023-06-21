package utils

import (
	"fmt"
	"os"
	"strconv"
)

func GetVersion() int {
	v, err := os.ReadFile("/root/version.txt")
	if err != nil {
		return 0
	}

	vi, err := strconv.Atoi(string(v))
	if err != nil {
		return 0
	}

	return vi
}

func SetVersion(v int) error {
	if err := os.WriteFile("/root/version.txt", []byte(fmt.Sprintf("%d", v)), 0644); err != nil {
		return err
	}

	return nil
}