package utils

import (
	"encoding/csv"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
)

func ReadYaml(url string, result interface{}) {
	f, err := os.Open(url)
	if err != nil {
		catcher.Error("Could not open file", err, nil)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	if err := d.Decode(result); err != nil {
		catcher.Error("Could not decode YAML", err, nil)
	}
}

func ReadCSV(url string) [][]string {
	f, err := os.Open(url)
	if err != nil {
		catcher.Error("Could not open file", err, nil)
	}
	defer f.Close()
	r := csv.NewReader(f)
	result, err := r.ReadAll()
	if err != nil {
		catcher.Error("Could not read CSV", err, nil)
	}
	return result
}

func ReadEnvVars(cfg interface{}) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// Check if the environment variable exists
		envValue, exists := os.LookupEnv(envTag)
		if !exists {
			catcher.Error("Environment variable not set", nil, map[string]any{"env_var": envTag})
			continue
		}

		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			catcher.Error("Cannot set field", nil, map[string]any{"field": field.Name})
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intValue, err := strconv.ParseInt(envValue, 10, fieldValue.Type().Bits()); err == nil {
				fieldValue.SetInt(intValue)
			} else {
				catcher.Error("Failed to convert to int", err, map[string]any{"value": envValue, "field": field.Name})
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if uintValue, err := strconv.ParseUint(envValue, 10, fieldValue.Type().Bits()); err == nil {
				fieldValue.SetUint(uintValue)
			} else {
				catcher.Error("Failed to convert to uint", err, map[string]any{"value": envValue, "field": field.Name})
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, err := strconv.ParseFloat(envValue, fieldValue.Type().Bits()); err == nil {
				fieldValue.SetFloat(floatValue)
			} else {
				catcher.Error("Failed to convert to float", err, map[string]any{"value": envValue, "field": field.Name})
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(envValue); err == nil {
				fieldValue.SetBool(boolValue)
			} else {
				catcher.Error("Failed to convert to bool", err, map[string]any{"value": envValue, "field": field.Name})
			}
		case reflect.Slice:
			elements := reflect.MakeSlice(fieldValue.Type(), 0, 0)
			for _, elem := range strings.Split(envValue, ",") {
				elements = reflect.Append(elements, reflect.ValueOf(elem))
			}
			fieldValue.Set(elements)
		case reflect.Ptr:
			ptr := reflect.New(fieldValue.Type().Elem())
			fieldValue.Set(ptr)
		default:
			catcher.Error("Unsupported field type", nil, map[string]any{"field": field.Name, "type": fieldValue.Kind()})
		}
	}
}
