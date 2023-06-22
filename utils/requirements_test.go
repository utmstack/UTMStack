package utils

import (
	"testing"
)
// Test CheckDisk
func TestCheckDisk(t *testing.T) {
	err := CheckDisk(100)
	t.Log(err)
}