package utils

import (
	"runtime"
	"strings"
	"testing"
)

func failingScript() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd.exe", []string{"/C", "echo out-line & echo err-line 1>&2 & exit /B 3"}
	}
	return "sh", []string{"-c", "echo out-line; echo err-line >&2; exit 3"}
}

func TestExecuteWithResultIncludesStderrOnFailure(t *testing.T) {
	c, args := failingScript()
	result, errB := ExecuteWithResult(c, "", args...)
	if !errB {
		t.Fatalf("errB = false, want true (result %q)", result)
	}
	for _, want := range []string{"out-line", "err-line", "exit status 3"} {
		if !strings.Contains(result, want) {
			t.Errorf("result %q is missing %q", result, want)
		}
	}
}

func TestExecuteWithResultFailureWithoutStderrUnchanged(t *testing.T) {
	var c string
	var args []string
	if runtime.GOOS == "windows" {
		c, args = "cmd.exe", []string{"/C", "exit /B 2"}
	} else {
		c, args = "sh", []string{"-c", "exit 2"}
	}
	result, errB := ExecuteWithResult(c, "", args...)
	if !errB {
		t.Fatalf("errB = false, want true")
	}
	if result != "exit status 2" {
		t.Errorf("result = %q, want %q", result, "exit status 2")
	}
}

func TestExecuteWithResultSuccessUnchanged(t *testing.T) {
	var c string
	var args []string
	if runtime.GOOS == "windows" {
		c, args = "cmd.exe", []string{"/C", "echo ok"}
	} else {
		c, args = "sh", []string{"-c", "echo ok"}
	}
	result, errB := ExecuteWithResult(c, "", args...)
	if errB {
		t.Fatalf("errB = true, want false (result %q)", result)
	}
	if !strings.Contains(result, "ok") {
		t.Errorf("result = %q, want it to contain %q", result, "ok")
	}
}
