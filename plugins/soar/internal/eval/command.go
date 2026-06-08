package eval

import (
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

var placeholderRE = regexp.MustCompile(`\$\((.*?)\)`)

func BuildCommand(template, alertJSON string) string {
	return placeholderRE.ReplaceAllStringFunc(template, func(match string) string {
		field := strings.TrimSuffix(strings.TrimPrefix(match, "$("), ")")
		val := gjson.Get(alertJSON, field)
		if !val.Exists() {
			return match
		}
		return val.String()
	})
}
