package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/soc-ai/configurations"
)

func ConvertFromAlertToAlertDB(alert Alert) AlertGPTDetails {
	return AlertGPTDetails{
		AlertID:                        alert.ID,
		Severity:                       alert.Severity,
		TagRulesApplied:                strings.Join(alert.TagRulesApplied, ","),
		SeverityLabel:                  alert.SeverityLabel,
		Notes:                          alert.Notes,
		DataType:                       alert.DataType,
		Description:                    alert.Description,
		StatusLabel:                    alert.StatusLabel,
		Tactic:                         alert.Tactic,
		Tags:                           strings.Join(alert.Tags, ","),
		Reference:                      strings.Join(alert.Reference, ","),
		Protocol:                       alert.Protocol,
		Timestamp:                      alert.Timestamp,
		Solution:                       alert.Solution,
		StatusObservation:              alert.StatusObservation,
		Name:                           alert.Name,
		IsIncident:                     alert.IsIncident,
		Category:                       alert.Category,
		DataSource:                     alert.DataSource,
		Logs:                           strings.Join(alert.Logs, configurations.LOGS_SEPARATOR),
		Status:                         alert.Status,
		DestinationCountry:             alert.Destination.Country,
		DestinationAccuracyRadius:      alert.Destination.AccuracyRadius,
		DestinationCity:                alert.Destination.City,
		DestinationIP:                  alert.Destination.IP,
		DestinationPort:                alert.Destination.Port,
		DestinationCountryCode:         alert.Destination.CountryCode,
		DestinationIsAnonymousProxy:    alert.Destination.IsAnonymousProxy,
		DestinationHost:                alert.Destination.Host,
		DestinationCoordinates:         strings.Join(convertFloatSliceToStringSlice(alert.Destination.Coordinates), ","),
		DestinationIsSatelliteProvider: alert.Destination.IsSatelliteProvider,
		DestinationAso:                 alert.Destination.Aso,
		DestinationAsn:                 alert.Destination.Asn,
		DestinationUser:                alert.Destination.User,
		SourceCountry:                  alert.Source.Country,
		SourceAccuracyRadius:           alert.Source.AccuracyRadius,
		SourceCity:                     alert.Source.City,
		SourceIP:                       alert.Source.IP,
		SourceCoordinates:              strings.Join(convertFloatSliceToStringSlice(alert.Source.Coordinates), ","),
		SourcePort:                     alert.Source.Port,
		SourceCountryCode:              alert.Source.CountryCode,
		SourceIsAnonymousProxy:         alert.Source.IsAnonymousProxy,
		SourceIsSatelliteProvider:      alert.Source.IsSatelliteProvider,
		SourceHost:                     alert.Source.Host,
		SourceAso:                      alert.Source.Aso,
		SourceAsn:                      alert.Source.Asn,
		SourceUser:                     alert.Source.User,
		IncidentCreatedBy:              alert.IncidentDetails.CreatedBy,
		IncidentObservation:            alert.IncidentDetails.Observation,
		IncidentSource:                 alert.IncidentDetails.Source,
		IncidentCreationDate:           alert.IncidentDetails.CreationDate,
		IncidentName:                   alert.IncidentDetails.IncidentName,
		IncidentID:                     alert.IncidentDetails.IncidentID,
		GPTTimestamp:                   time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00"),
		GPTClassification:              "",
		GPTReasoning:                   "",
		GPTNextSteps:                   "",
	}
}

func convertFloatSliceToStringSlice(floatSlice []float64) []string {
	strSlice := make([]string, len(floatSlice))
	for i, num := range floatSlice {
		strSlice[i] = strconv.FormatFloat(num, 'f', -1, 64)
	}
	return strSlice
}

func ConvertFromAlertDBToGPTResponse(alertDetails AlertGPTDetails) GPTAlertResponse {
	resp := GPTAlertResponse{
		Timestamp:      time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00"),
		Severity:       alertDetails.Severity,
		Category:       alertDetails.Category,
		AlertName:      alertDetails.Name,
		ActivityID:     alertDetails.AlertID,
		Classification: alertDetails.GPTClassification,
		Reasoning:      strings.Split(alertDetails.GPTReasoning, configurations.LOGS_SEPARATOR),
		NextSteps:      []NextStep{},
	}

	nextSteps := strings.Split(alertDetails.GPTNextSteps, "\n")
	for i, step := range nextSteps {
		actionAndDetails := strings.Split(step, "::")
		if len(actionAndDetails) < 2 {
			continue
		}
		resp.NextSteps = append(resp.NextSteps, NextStep{
			Step:    i + 1,
			Action:  actionAndDetails[0],
			Details: actionAndDetails[1],
		})
	}

	return resp
}

func ConvertGPTResponseToUpdateQuery(gptResp GPTAlertResponse) (UpdateDocRequest, error) {
	source, err := BuildScriptString(gptResp)
	if err != nil {
		return UpdateDocRequest{}, err
	}

	return UpdateDocRequest{
		Query: Query{
			Bool: Bool{
				Must: []Must{
					{Match{
						ActivityID: gptResp.ActivityID,
					}},
				},
			},
		},
		Script: Script{
			Source: source,
			Lang:   "painless",
			Params: gptResp,
		},
	}, nil
}

func BuildScriptString(alert GPTAlertResponse) (string, error) {
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
