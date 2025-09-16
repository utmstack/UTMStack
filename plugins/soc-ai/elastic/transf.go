package elastic

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
)

func ConvertFromAlertDBToGPTResponse(alertDetails *schema.AlertFields) schema.GPTAlertResponse {
	if alertDetails == nil {
		return schema.GPTAlertResponse{}
	}

	resp := schema.GPTAlertResponse{
		Timestamp:      time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00"),
		Severity:       alertDetails.Severity,
		Category:       alertDetails.Category,
		AlertName:      alertDetails.Name,
		ActivityID:     alertDetails.ID,
		Classification: alertDetails.GPTClassification,
		Reasoning:      strings.Split(alertDetails.GPTReasoning, config.LOGS_SEPARATOR),
		NextSteps:      []schema.NextStep{},
	}

	nextSteps := strings.Split(alertDetails.GPTNextSteps, "\n")
	for i, step := range nextSteps {
		actionAndDetails := strings.Split(step, "::")
		if len(actionAndDetails) < 2 {
			continue
		}
		resp.NextSteps = append(resp.NextSteps, schema.NextStep{
			Step:    i + 1,
			Action:  actionAndDetails[0],
			Details: actionAndDetails[1],
		})
	}

	return resp
}

func ConvertGPTResponseToUpdateQuery(gptResp schema.GPTAlertResponse) (schema.UpdateDocRequest, error) {
	source, err := buildScriptString(gptResp)
	if err != nil {
		return schema.UpdateDocRequest{}, err
	}

	return schema.UpdateDocRequest{
		Query: schema.Query{
			Bool: schema.Bool{
				Must: []schema.Must{
					{Match: schema.Match{
						ActivityID: gptResp.ActivityID,
					}},
				},
			},
		},
		Script: schema.Script{
			Source: source,
			Lang:   "painless",
			Params: gptResp,
		},
	}, nil
}

func buildScriptString(alert schema.GPTAlertResponse) (string, error) {
	v := reflect.ValueOf(alert)
	typeOfAlert := v.Type()

	source := ""
	for i := 0; i < v.NumField(); i++ {
		jsonTag := typeOfAlert.Field(i).Tag.Get("json")
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		fieldValue := v.Field(i).Interface()

		switch reflect.TypeOf(fieldValue).Kind() {
		case reflect.String, reflect.Int, reflect.Struct:
			if fieldValue != reflect.Zero(reflect.TypeOf(fieldValue)).Interface() {
				source += fmt.Sprintf("ctx._source['%s'] = params['%s']; ", jsonFieldName, jsonFieldName)
			}
		case reflect.Slice:
			s := reflect.ValueOf(fieldValue)
			if s.Len() > 0 {
				source += fmt.Sprintf("ctx._source['%s'] = params.%s; ", jsonFieldName, jsonFieldName)
			}
		default:
			return "", fmt.Errorf("unsupported type: %v", reflect.TypeOf(fieldValue).Kind())
		}
	}

	return source, nil
}
