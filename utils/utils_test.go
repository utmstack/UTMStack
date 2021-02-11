package utils

import (
	"testing"
)

func TestRunEnvCmd(t *testing.T) {
	if err := runEnvCmd("cli", []string{"MYVAR=test"}, "echo", "test"); err != nil {
		t.Fatal(err)
	}
}

func TestRunCmd(t *testing.T) {
	if err := runCmd("cli", "echo", "arg1", "arg2"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckOutput(t *testing.T) {
	if out, err := checkOutput("echo", "arg1", "arg2"); out != "arg1 arg2" {
		t.Fatalf("Unexpected output: %v", out)
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
