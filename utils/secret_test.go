package utils

import (
	"testing"
)

func TestSecret(t *testing.T) {
	s := GenerateSecret(10)
	t.Log(s)
}