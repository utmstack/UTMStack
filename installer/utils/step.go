package utils

import (
	"fmt"
	"os"
)

func GetStep(v int) bool {
	_, err := os.ReadFile(fmt.Sprintf("/root/%d.lock", v))

	return err != nil
}

func SetStep(v int) error {
	if err := os.WriteFile(fmt.Sprintf("/root/%d.lock", v), []byte(fmt.Sprintf("%d", v)), 0644); err != nil {
		return err
	}

	return nil
}
