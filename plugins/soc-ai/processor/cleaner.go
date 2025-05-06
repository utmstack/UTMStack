package processor

import (
	"reflect"
	"regexp"

	"github.com/utmstack/UTMStack/plugins/soc-ai/configurations"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

func cleanAlerts(alertDetails *schema.AlertGPTDetails) schema.AlertGPTDetails {
	v := reflect.ValueOf(alertDetails).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String {
			str := field.String()
			for _, pattern := range configurations.SensitivePatterns {
				re := regexp.MustCompile(pattern.Regexp)
				if re.MatchString(str) {
					str = re.ReplaceAllString(str, pattern.FakeValue)
				}
			}
			field.SetString(str)
		}
	}
	return *alertDetails
}
