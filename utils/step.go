package utils

import (
	"fmt"
	"os"
	"strconv"
)

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

func SetStep(v int) error {
	if err := os.WriteFile("/root/step.txt", []byte(fmt.Sprintf("%d", v)), 0644); err != nil {
		return err
	}

	return nil
}