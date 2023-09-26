package utils

import "os"

// GenerateConfig generates a configuration file from a text template.
//
// The `variables` argument is a map of variables to be used in the template.
// The `text` argument is the text of the template.
// The `output` argument is the path to the output file.
//
// The function returns an error if there is any problem generating the configuration file.
func GenerateConfig(variables interface{}, text, output string) error {
	tpl, err := ParseTemplate(text, variables)
	if err != nil {
		return err
	}

	keyFile, err := os.Create(output)
	if err != nil {
		return err
	}

	defer keyFile.Close()

	_, err = keyFile.Write(tpl.Bytes())
	if err != nil {
		return err
	}

	return nil
}
