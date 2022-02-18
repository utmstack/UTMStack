package utils

import (
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"
)

func writeToFile(fileName string, body string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(body)
	return err
}

func getMyPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
}

func readYAML(url string, result interface{}) error {
	f, err := os.Open(url)
	if err != nil {
		return err
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	if err := d.Decode(result); err != nil {
		return err
	}
	return nil
}

func writeYAML(url string, data interface{}) error {
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

func generateFromTemplate(data interface{}, tfile string, cfile string) error {
	_, fileName := filepath.Split(tfile)
	ut, err := template.New(fileName).ParseFiles(tfile)

	if err != nil {
		return err
	}

	writer, err := os.OpenFile(cfile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	err = ut.Execute(writer, data)

	if err != nil {
		return err
	}

	return nil
}
