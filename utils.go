package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func run_env_cmd(env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Executing command:", command, "with args:", arg)
	return cmd.Run()
}

func run_cmd(command string, arg ...string) error {
	return run_env_cmd([]string{}, command, arg...)
}

func check_cmd(command string, arg ...string) {
	if err := run_cmd(command, arg...); err != nil {
		log.Fatal(err)
	}
}

func check_output(command string, arg ...string) string {
	log.Println("Executing command:", command, "with args:", arg)
	out, err := exec.Command(command, arg...).Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
