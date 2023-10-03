package utils_test

import (
	"testing"
	"github.com/utmstack/UTMStack/installer/utils"
)

func TestSecret(t *testing.T) {
	s := utils.GenerateSecret(10)
	t.Log(s)
}