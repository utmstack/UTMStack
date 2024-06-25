package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func GetMyPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func WriteStringToFile(fileName string, body string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(body)
	return err
}

func WriteBytesToFile(fileName string, data []byte) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data) // Cambiado de WriteString a Write para aceptar []byte directamente
	return err
}

func WriteJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	err = WriteStringToFile(path, string(jsonData[:]))
	if err != nil {
		return err
	}

	return nil
}

func CreatePathIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return fmt.Errorf("error creating path: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking path: %v", err)
	}
	return nil
}
