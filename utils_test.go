package main

import (
	"testing"
)

func TestRunEnvCmd(t *testing.T) {
	if err := runEnvCmd([]string{"MYVAR=test"}, "echo", "test"); err != nil {
		t.Fatal(err)
	}
}

func TestRunCmd(t *testing.T) {
	if err := runCmd("echo", "arg1", "arg2"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckOutput(t *testing.T) {
	if output := checkOutput("echo", "arg1", "arg2"); output != "arg1 arg2" {
		t.Fatalf("Unexpected output: %v", output)
	}
}

func TestCheck(t *testing.T) {
	check(runCmd("echo", "arg1"))
}
