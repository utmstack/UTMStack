package utils_test

import (
	"testing"
	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func TestSecret(t *testing.T) {
	s := utils.GenerateSecret(10)
	t.Log(s)
}