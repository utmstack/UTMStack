package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

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

func writeToFile(fileName string, body string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(body)
	return err
}

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
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

func ReadFilePart(filePath string, partSizeMB int, partIndex int) ([]byte, int, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error getting file info: %v", err)
	}

	partSizeBytes := partSizeMB * 1024 * 1024
	totalParts := int((fileInfo.Size() + int64(partSizeBytes) - 1) / int64(partSizeBytes))
	offset := int64((partIndex - 1) * partSizeBytes)

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error seeking file: %v", err)
	}

	buf := make([]byte, partSizeBytes)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, 0, 0, fmt.Errorf("error reading file: %v", err)
	}

	return buf[:n], totalParts, fileInfo.Size(), nil
}
