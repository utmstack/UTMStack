package utils

import (
	"encoding/csv"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ReadYaml(url string, result interface{}) {
	f, err := os.Open(url)
	if err != nil {
		log.Printf("Could not open file: %v", err)
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	if err := d.Decode(result); err != nil {
		log.Printf("Could not decode YAML: %v", err)
	}
}

func ReadCSV(url string) [][]string {
	f, err := os.Open(url)
	if err != nil {
		log.Printf("Could not open file: %v", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	result, err := r.ReadAll()
	if err != nil {
		log.Printf("Could not read CSV: %v", err)
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
			log.Printf("Environment variable %s not set, skipping...", envTag)
			continue
		}

		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			log.Printf("Cannot set field %s, skipping...", field.Name)
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intValue, err := strconv.ParseInt(envValue, 10, fieldValue.Type().Bits()); err == nil {
				fieldValue.SetInt(intValue)
			} else {
				log.Printf("Failed to convert %s to int for field %s: %v", envValue, field.Name, err)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if uintValue, err := strconv.ParseUint(envValue, 10, fieldValue.Type().Bits()); err == nil {
				fieldValue.SetUint(uintValue)
			} else {
				log.Printf("Failed to convert %s to uint for field %s: %v", envValue, field.Name, err)
			}
		case reflect.Float32, reflect.Float64:
			if floatValue, err := strconv.ParseFloat(envValue, fieldValue.Type().Bits()); err == nil {
				fieldValue.SetFloat(floatValue)
			} else {
				log.Printf("Failed to convert %s to float for field %s: %v", envValue, field.Name, err)
			}
		case reflect.Bool:
			if boolValue, err := strconv.ParseBool(envValue); err == nil {
				fieldValue.SetBool(boolValue)
			} else {
				log.Printf("Failed to convert %s to bool for field %s: %v", envValue, field.Name, err)
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
			log.Printf("Unsupported field type %s for field %s", fieldValue.Kind(), field.Name)
		}
	}
}
