package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func RunEnvCmd(env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)

	f, err := os.OpenFile("/var/log/utmstack-installer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("error opening log file: %v", err)
	}

	defer f.Close()

	cmd.Stdout = f

	var eBuf bytes.Buffer
	cmd.Stderr = &eBuf

	err = cmd.Run()
	if err != nil {
		f.Write(eBuf.Bytes())
		return fmt.Errorf("error running command: %v, %v", err, eBuf.String())
	}

	return nil
}

func RunCmd(command string, arg ...string) error {
	return RunEnvCmd([]string{}, command, arg...)
}

func MakeDir(mode os.FileMode, arg ...string) string {
	path := ""
	for _, folder := range arg {
		path = filepath.Join(path, folder)
		os.MkdirAll(path, mode)
		os.Chmod(path, mode)
	}
	return path
}

func Remove(path string) error {
	return os.Remove(path)
}

func CreatePathIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("error creating path: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking path: %v", err)
	}
	return nil
}

func writeToFile(fileName string, body string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(body)
	return err
}

func WriteYAML(url string, data interface{}) error {
	config, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = writeToFile(url, string(config[:]))
	if err != nil {
		return err
	}

	return nil
}

func ReadYAML(path string, result interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(result); err != nil {
		return err
	}
	return nil
}

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func WriteJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	err = writeToFile(path, string(jsonData[:]))
	if err != nil {
		return err
	}

	return nil
}

func ReadJson(fileName string, data interface{}) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, data)
	if err != nil {
		return err
	}
	return nil
}
