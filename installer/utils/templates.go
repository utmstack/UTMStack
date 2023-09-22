package utils

import (
	"bytes"
	"text/template"
)

// ParseTemplate parses a template and executes it with the given data.
//
// text: The text of the template.
// data: The data to use to render the template.
//
// Returns: A bytes.Buffer containing the rendered template, or an error if
// parsing or execution failed.
func ParseTemplate(text string, data interface{}) (*bytes.Buffer, error) {
	var tpl = new(bytes.Buffer)

	t := template.Must(template.New("file").Parse(text))

	err := t.Execute(tpl, data)
	if err != nil {
		return tpl, err
	}

	return tpl, nil
}
