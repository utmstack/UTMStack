package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/plugins/soc-ai/configurations"
)

func ConvertFromAlertToAlertDB(alert Alert) AlertGPTDetails {
	return AlertGPTDetails{
		Timestamp:         alert.Timestamp,
		ID:                alert.ID,
		ParentID:          alert.ParentID,
		Status:            alert.Status,
		StatusLabel:       alert.StatusLabel,
		StatusObservation: alert.StatusObservation,
		IsIncident:        alert.IsIncident,
		Name:              alert.Name,
		Category:          alert.Category,
		Severity:          alert.Severity,
		SeverityLabel:     alert.SeverityLabel,
		Description:       alert.Description,
		Solution:          alert.Solution,
		Technique:         alert.Technique,
		Reference:         alert.Reference,
		DataType:          alert.DataType,
		ImpactScore:       alert.ImpactScore,
		DataSource:        alert.DataSource,
		Tags:              alert.Tags,
		Notes:             alert.Notes,
		TagRulesApplied:   alert.TagRulesApplied,
		DeduplicatedBy:    alert.DeduplicatedBy,

		IncidentCreatedBy:    alert.IncidentDetail.CreatedBy,
		IncidentObservation:  alert.IncidentDetail.Observation,
		IncidentCreationDate: alert.IncidentDetail.CreationDate,
		IncidentSource:       alert.IncidentDetail.Source,

		ImpactConfidentiality: alert.Impact.Confidentiality,
		ImpactIntegrity:       alert.Impact.Integrity,
		ImpactAvailability:    alert.Impact.Availability,

		EventsId:               alert.Events[0].Id,
		EventsTimestamp:        alert.Events[0].Timestamp,
		EventsDeviceTime:       alert.Events[0].DeviceTime,
		EventsDataType:         alert.Events[0].DataType,
		EventsDataSource:       alert.Events[0].DataSource,
		EventsTenantId:         alert.Events[0].TenantId,
		EventsTenantName:       alert.Events[0].TenantName,
		EventsRaw:              alert.Events[0].Raw,
		EventsLog:              alert.Events[0].Log,
		EventsProtocol:         alert.Events[0].Protocol,
		EventsConnectionStatus: alert.Events[0].ConnectionStatus,
		EventsStatusCode:       alert.Events[0].StatusCode,
		EventsActionResult:     alert.Events[0].ActionResult,
		EventsAction:           alert.Events[0].Action,
		EventsCommand:          alert.Events[0].Command,
		EventsSeverity:         alert.Events[0].Severity,

		LastEventId:               alert.LastEvent.Id,
		LastEventTimestamp:        alert.LastEvent.Timestamp,
		LastEventDeviceTime:       alert.LastEvent.DeviceTime,
		LastEventDataType:         alert.LastEvent.DataType,
		LastEventDataSource:       alert.LastEvent.DataSource,
		LastEventTenantId:         alert.LastEvent.TenantId,
		LastEventTenantName:       alert.LastEvent.TenantName,
		LastEventRaw:              alert.LastEvent.Raw,
		LastEventLog:              alert.LastEvent.Log,
		LastEventProtocol:         alert.LastEvent.Protocol,
		LastEventConnectionStatus: alert.LastEvent.ConnectionStatus,
		LastEventStatusCode:       alert.LastEvent.StatusCode,
		LastEventActionResult:     alert.LastEvent.ActionResult,
		LastEventAction:           alert.LastEvent.Action,
		LastEventCommand:          alert.LastEvent.Command,
		LastEventSeverity:         alert.LastEvent.Severity,

		AdversaryBytesSent:        alert.Adversary.BytesSent,
		AdversaryBytesReceived:    alert.Adversary.BytesReceived,
		AdversaryPackagesSent:     alert.Adversary.PackagesSent,
		AdversaryPackagesReceived: alert.Adversary.PackagesReceived,
		AdversaryConnections:      alert.Adversary.Connections,
		AdversaryUsedCpuPercent:   alert.Adversary.UsedCpuPercent,
		AdversaryUsedMemPercent:   alert.Adversary.UsedMemPercent,
		AdversaryTotalCpuUnits:    alert.Adversary.TotalCpuUnits,
		AdversaryTotalMem:         alert.Adversary.TotalMem,
		AdversaryIp:               alert.Adversary.Ip,
		AdversaryHost:             alert.Adversary.Host,
		AdversaryUser:             alert.Adversary.User,
		AdversaryGroup:            alert.Adversary.Group,
		AdversaryPort:             alert.Adversary.Port,
		AdversaryDomain:           alert.Adversary.Domain,
		AdversaryFqdn:             alert.Adversary.Fqdn,
		AdversaryMac:              alert.Adversary.Mac,
		AdversaryProcess:          alert.Adversary.Process,
		AdversaryFile:             alert.Adversary.File,
		AdversaryPath:             alert.Adversary.Path,
		AdversaryHash:             alert.Adversary.Hash,
		AdversaryUrl:              alert.Adversary.Url,
		AdversaryEmail:            alert.Adversary.Email,

		AdversaryGeolocationCountry:     alert.Adversary.Geolocation.GeolocationCountry,
		AdversaryGeolocationCity:        alert.Adversary.Geolocation.GeolocationCity,
		AdversaryGeolocationLatitude:    alert.Adversary.Geolocation.GeolocationLatitude,
		AdversaryGeolocationLongitude:   alert.Adversary.Geolocation.GeolocationLongitude,
		AdversaryGeolocationAsn:         alert.Adversary.Geolocation.GeolocationAsn,
		AdversaryGeolocationAso:         alert.Adversary.Geolocation.GeolocationAso,
		AdversaryGeolocationCountryCode: alert.Adversary.Geolocation.GeolocationCountryCode,
		AdversaryGeolocationAccuracy:    alert.Adversary.Geolocation.GeolocationAccuracy,

		TargetBytesSent:        alert.Target.BytesSent,
		TargetBytesReceived:    alert.Target.BytesReceived,
		TargetPackagesSent:     alert.Target.PackagesSent,
		TargetPackagesReceived: alert.Target.PackagesReceived,
		TargetConnections:      alert.Target.Connections,
		TargetUsedCpuPercent:   alert.Target.UsedCpuPercent,
		TargetUsedMemPercent:   alert.Target.UsedMemPercent,
		TargetTotalCpuUnits:    alert.Target.TotalCpuUnits,
		TargetTotalMem:         alert.Target.TotalMem,
		TargetIp:               alert.Target.Ip,
		TargetHost:             alert.Target.Host,
		TargetUser:             alert.Target.User,
		TargetGroup:            alert.Target.Group,
		TargetPort:             alert.Target.Port,
		TargetDomain:           alert.Target.Domain,
		TargetFqdn:             alert.Target.Fqdn,
		TargetMac:              alert.Target.Mac,
		TargetProcess:          alert.Target.Process,
		TargetFile:             alert.Target.File,
		TargetPath:             alert.Target.Path,
		TargetHash:             alert.Target.Hash,
		TargetUrl:              alert.Target.Url,
		TargetEmail:            alert.Target.Email,

		TargetGeolocationCountry:     alert.Target.Geolocation.GeolocationCountry,
		TargetGeolocationCity:        alert.Target.Geolocation.GeolocationCity,
		TargetGeolocationLatitude:    alert.Target.Geolocation.GeolocationLatitude,
		TargetGeolocationLongitude:   alert.Target.Geolocation.GeolocationLongitude,
		TargetGeolocationAsn:         alert.Target.Geolocation.GeolocationAsn,
		TargetGeolocationAso:         alert.Target.Geolocation.GeolocationAso,
		TargetGeolocationCountryCode: alert.Target.Geolocation.GeolocationCountryCode,
		TargetGeolocationAccuracy:    alert.Target.Geolocation.GeolocationAccuracy,

		EventOriginBytesSent:        alert.Events[0].Origin.BytesSent,
		EventOriginBytesReceived:    alert.Events[0].Origin.BytesReceived,
		EventOriginPackagesSent:     alert.Events[0].Origin.PackagesSent,
		EventOriginPackagesReceived: alert.Events[0].Origin.PackagesReceived,
		EventOriginConnections:      alert.Events[0].Origin.Connections,
		EventOriginUsedCpuPercent:   alert.Events[0].Origin.UsedCpuPercent,
		EventOriginUsedMemPercent:   alert.Events[0].Origin.UsedMemPercent,
		EventOriginTotalCpuUnits:    alert.Events[0].Origin.TotalCpuUnits,
		EventOriginTotalMem:         alert.Events[0].Origin.TotalMem,
		EventOriginIp:               alert.Events[0].Origin.Ip,
		EventOriginHost:             alert.Events[0].Origin.Host,
		EventOriginUser:             alert.Events[0].Origin.User,
		EventOriginGroup:            alert.Events[0].Origin.Group,
		EventOriginPort:             alert.Events[0].Origin.Port,
		EventOriginDomain:           alert.Events[0].Origin.Domain,
		EventOriginFqdn:             alert.Events[0].Origin.Fqdn,
		EventOriginMac:              alert.Events[0].Origin.Mac,
		EventOriginProcess:          alert.Events[0].Origin.Process,
		EventOriginFile:             alert.Events[0].Origin.File,
		EventOriginPath:             alert.Events[0].Origin.Path,
		EventOriginHash:             alert.Events[0].Origin.Hash,
		EventOriginUrl:              alert.Events[0].Origin.Url,
		EventOriginEmail:            alert.Events[0].Origin.Email,

		EventOriginGeolocationCountry:     alert.Events[0].Origin.Geolocation.GeolocationCountry,
		EventOriginGeolocationCity:        alert.Events[0].Origin.Geolocation.GeolocationCity,
		EventOriginGeolocationLatitude:    alert.Events[0].Origin.Geolocation.GeolocationLatitude,
		EventOriginGeolocationLongitude:   alert.Events[0].Origin.Geolocation.GeolocationLongitude,
		EventOriginGeolocationAsn:         alert.Events[0].Origin.Geolocation.GeolocationAsn,
		EventOriginGeolocationAso:         alert.Events[0].Origin.Geolocation.GeolocationAso,
		EventOriginGeolocationCountryCode: alert.Events[0].Origin.Geolocation.GeolocationCountryCode,
		EventOriginGeolocationAccuracy:    alert.Events[0].Origin.Geolocation.GeolocationAccuracy,

		EventTargetBytesSent:        alert.Events[0].Target.BytesSent,
		EventTargetBytesReceived:    alert.Events[0].Target.BytesReceived,
		EventTargetPackagesSent:     alert.Events[0].Target.PackagesSent,
		EventTargetPackagesReceived: alert.Events[0].Target.PackagesReceived,
		EventTargetConnections:      alert.Events[0].Target.Connections,
		EventTargetUsedCpuPercent:   alert.Events[0].Target.UsedCpuPercent,
		EventTargetUsedMemPercent:   alert.Events[0].Target.UsedMemPercent,
		EventTargetTotalCpuUnits:    alert.Events[0].Target.TotalCpuUnits,
		EventTargetTotalMem:         alert.Events[0].Target.TotalMem,
		EventTargetIp:               alert.Events[0].Target.Ip,
		EventTargetHost:             alert.Events[0].Target.Host,
		EventTargetUser:             alert.Events[0].Target.User,
		EventTargetGroup:            alert.Events[0].Target.Group,
		EventTargetPort:             alert.Events[0].Target.Port,
		EventTargetDomain:           alert.Events[0].Target.Domain,
		EventTargetFqdn:             alert.Events[0].Target.Fqdn,
		EventTargetMac:              alert.Events[0].Target.Mac,
		EventTargetProcess:          alert.Events[0].Target.Process,
		EventTargetFile:             alert.Events[0].Target.File,
		EventTargetPath:             alert.Events[0].Target.Path,
		EventTargetHash:             alert.Events[0].Target.Hash,
		EventTargetUrl:              alert.Events[0].Target.Url,
		EventTargetEmail:            alert.Events[0].Target.Email,

		EventTargetGeolocationCountry:     alert.Events[0].Target.Geolocation.GeolocationCountry,
		EventTargetGeolocationCity:        alert.Events[0].Target.Geolocation.GeolocationCity,
		EventTargetGeolocationLatitude:    alert.Events[0].Target.Geolocation.GeolocationLatitude,
		EventTargetGeolocationLongitude:   alert.Events[0].Target.Geolocation.GeolocationLongitude,
		EventTargetGeolocationAsn:         alert.Events[0].Target.Geolocation.GeolocationAsn,
		EventTargetGeolocationAso:         alert.Events[0].Target.Geolocation.GeolocationAso,
		EventTargetGeolocationCountryCode: alert.Events[0].Target.Geolocation.GeolocationCountryCode,
		EventTargetGeolocationAccuracy:    alert.Events[0].Target.Geolocation.GeolocationAccuracy,

		LastEventOriginBytesSent:        alert.LastEvent.Origin.BytesSent,
		LastEventOriginBytesReceived:    alert.LastEvent.Origin.BytesReceived,
		LastEventOriginPackagesSent:     alert.LastEvent.Origin.PackagesSent,
		LastEventOriginPackagesReceived: alert.LastEvent.Origin.PackagesReceived,
		LastEventOriginConnections:      alert.LastEvent.Origin.Connections,
		LastEventOriginUsedCpuPercent:   alert.LastEvent.Origin.UsedCpuPercent,
		LastEventOriginUsedMemPercent:   alert.LastEvent.Origin.UsedMemPercent,
		LastEventOriginTotalCpuUnits:    alert.LastEvent.Origin.TotalCpuUnits,
		LastEventOriginTotalMem:         alert.LastEvent.Origin.TotalMem,
		LastEventOriginIp:               alert.LastEvent.Origin.Ip,
		LastEventOriginHost:             alert.LastEvent.Origin.Host,
		LastEventOriginUser:             alert.LastEvent.Origin.User,
		LastEventOriginGroup:            alert.LastEvent.Origin.Group,
		LastEventOriginPort:             alert.LastEvent.Origin.Port,
		LastEventOriginDomain:           alert.LastEvent.Origin.Domain,
		LastEventOriginFqdn:             alert.LastEvent.Origin.Fqdn,
		LastEventOriginMac:              alert.LastEvent.Origin.Mac,
		LastEventOriginProcess:          alert.LastEvent.Origin.Process,
		LastEventOriginFile:             alert.LastEvent.Origin.File,
		LastEventOriginPath:             alert.LastEvent.Origin.Path,
		LastEventOriginHash:             alert.LastEvent.Origin.Hash,
		LastEventOriginUrl:              alert.LastEvent.Origin.Url,
		LastEventOriginEmail:            alert.LastEvent.Origin.Email,

		LastEventOriginGeolocationCountry:     alert.LastEvent.Origin.Geolocation.GeolocationCountry,
		LastEventOriginGeolocationCity:        alert.LastEvent.Origin.Geolocation.GeolocationCity,
		LastEventOriginGeolocationLatitude:    alert.LastEvent.Origin.Geolocation.GeolocationLatitude,
		LastEventOriginGeolocationLongitude:   alert.LastEvent.Origin.Geolocation.GeolocationLongitude,
		LastEventOriginGeolocationAsn:         alert.LastEvent.Origin.Geolocation.GeolocationAsn,
		LastEventOriginGeolocationAso:         alert.LastEvent.Origin.Geolocation.GeolocationAso,
		LastEventOriginGeolocationCountryCode: alert.LastEvent.Origin.Geolocation.GeolocationCountryCode,
		LastEventOriginGeolocationAccuracy:    alert.LastEvent.Origin.Geolocation.GeolocationAccuracy,

		LastEventTargetBytesSent:        alert.LastEvent.Target.BytesSent,
		LastEventTargetBytesReceived:    alert.LastEvent.Target.BytesReceived,
		LastEventTargetPackagesSent:     alert.LastEvent.Target.PackagesSent,
		LastEventTargetPackagesReceived: alert.LastEvent.Target.PackagesReceived,
		LastEventTargetConnections:      alert.LastEvent.Target.Connections,
		LastEventTargetUsedCpuPercent:   alert.LastEvent.Target.UsedCpuPercent,
		LastEventTargetUsedMemPercent:   alert.LastEvent.Target.UsedMemPercent,
		LastEventTargetTotalCpuUnits:    alert.LastEvent.Target.TotalCpuUnits,
		LastEventTargetTotalMem:         alert.LastEvent.Target.TotalMem,
		LastEventTargetIp:               alert.LastEvent.Target.Ip,
		LastEventTargetHost:             alert.LastEvent.Target.Host,
		LastEventTargetUser:             alert.LastEvent.Target.User,
		LastEventTargetGroup:            alert.LastEvent.Target.Group,
		LastEventTargetPort:             alert.LastEvent.Target.Port,
		LastEventTargetDomain:           alert.LastEvent.Target.Domain,
		LastEventTargetFqdn:             alert.LastEvent.Target.Fqdn,
		LastEventTargetMac:              alert.LastEvent.Target.Mac,
		LastEventTargetProcess:          alert.LastEvent.Target.Process,
		LastEventTargetFile:             alert.LastEvent.Target.File,
		LastEventTargetPath:             alert.LastEvent.Target.Path,
		LastEventTargetHash:             alert.LastEvent.Target.Hash,
		LastEventTargetUrl:              alert.LastEvent.Target.Url,
		LastEventTargetEmail:            alert.LastEvent.Target.Email,

		LastEventTargetGeolocationCountry:     alert.LastEvent.Target.Geolocation.GeolocationCountry,
		LastEventTargetGeolocationCity:        alert.LastEvent.Target.Geolocation.GeolocationCity,
		LastEventTargetGeolocationLatitude:    alert.LastEvent.Target.Geolocation.GeolocationLatitude,
		LastEventTargetGeolocationLongitude:   alert.LastEvent.Target.Geolocation.GeolocationLongitude,
		LastEventTargetGeolocationAsn:         alert.LastEvent.Target.Geolocation.GeolocationAsn,
		LastEventTargetGeolocationAso:         alert.LastEvent.Target.Geolocation.GeolocationAso,
		LastEventTargetGeolocationCountryCode: alert.LastEvent.Target.Geolocation.GeolocationCountryCode,
		LastEventTargetGeolocationAccuracy:    alert.LastEvent.Target.Geolocation.GeolocationAccuracy,

		GPTTimestamp:      time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00"),
		GPTClassification: "",
		GPTReasoning:      "",
		GPTNextSteps:      "",
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
		ActivityID:     alertDetails.ID,
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
