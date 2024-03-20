package netflow

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/utmstack/UTMStack/agent/agent/configuration"
)

func Dump(metrics []Metric, remote string) []string {
	var allKVPairs []string
	for _, metric := range metrics {
		t := reflect.TypeOf(metric)
		v := reflect.ValueOf(metric)
		var kvPairs []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			if value.String() == "" {
				continue
			}
			header := field.Tag.Get("header")
			kvPairs = append(kvPairs, fmt.Sprintf("%s=\"%v\"", header, value))
		}
		allKVPairs = append(allKVPairs, configuration.GetMessageFormated(remote, strings.Join(kvPairs, " ")))
	}
	return allKVPairs
}
