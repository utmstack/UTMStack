package eventstore

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type Migration struct {
	Name    string
	Content string
}


func fillTemplate(template_name, template_content string, args map[string]string) (string,error) {
	t,err := template.New(template_name).Parse(template_content)
	if err!=nil{
		return "",err
	}

	var buf bytes.Buffer

	if err := t.Execute(&buf, args); err != nil {
		panic(err)
	}

	result := buf.String()

	return result,nil
}

func loadMigrations(dir string, args map[string]string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil,err
	}

	var migrations []Migration

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		content, err := os.ReadFile(path)
		if err != nil {
			return nil,  err
		}

		str_content,err := fillTemplate(entry.Name(),string(content),args)
		if err !=nil{
			return nil,err
		}

		migrations = append(migrations, Migration{
			Name:    entry.Name(),
			Content: str_content,
		})
	}

	return migrations, nil
}

