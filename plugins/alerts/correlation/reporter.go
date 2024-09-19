package correlation

import (
	"encoding/json"
	"strconv"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/alerts/search"
	"github.com/utmstack/UTMStack/plugins/alerts/utils"
)

type Side struct {
	User                string     `json:"user"`
	Host                string     `json:"host"`
	IP                  string     `json:"ip"`
	Port                int        `json:"port"`
	Country             string     `json:"country"`
	CountryCode         string     `json:"countryCode"`
	City                string     `json:"city"`
	Coordinates         [2]float64 `json:"coordinates"`
	AccuracyRadius      int        `json:"accuracyRadius"`
	ASN                 int        `json:"asn"`
	ASO                 string     `json:"aso"`
	IsSatelliteProvider bool       `json:"isSatelliteProvider"`
	IsAnonymousProxy    bool       `json:"isAnonymousProxy"`
}

type IncidentDetail struct {
	CreatedBy    string `json:"createdBy"`
	Observation  string `json:"observation"`
	CreationDate string `json:"creationDate"`
	Source       string `json:"source"`
}

type AlertFields struct {
	Timestamp         string         `json:"@timestamp"`
	ID                string         `json:"id"`
	Status            int            `json:"status"`
	StatusLabel       string         `json:"statusLabel"`
	StatusObservation string         `json:"statusObservation"`
	IsIncident        bool           `json:"isIncident"`
	IncidentDetail    IncidentDetail `json:"incidentDetail"`
	Name              string         `json:"name"`
	Category          string         `json:"category"`
	Severity          int            `json:"severity"`
	SeverityLabel     string         `json:"severityLabel"`
	Protocol          string         `json:"protocol"`
	Description       string         `json:"description"`
	Solution          string         `json:"solution"`
	Tactic            string         `json:"tactic"`
	Reference         []string       `json:"reference"`
	DataType          string         `json:"dataType"`
	DataSource        string         `json:"dataSource"`
	Source            Side           `json:"source"`
	Destination       Side           `json:"destination"`
	Logs              []string       `json:"logs"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"TagRulesApplied"`
}

func Alert(id, name, severity, description, solution, category, tactic string, reference []string, dataType, dataSource string,
	details map[string]string) {

	if !UpdateAlert(name, severity, details) {
		NewAlert(id, name, severity, description, solution, category, tactic, reference, dataType, dataSource,
			details)
	}
}

func UpdateAlert(name, severity string, details map[string]string) bool {
	var updated bool

	var severityLabel string
	switch severity {
	case "low":
		severityLabel = "Low"
	case "medium":
		severityLabel = "Medium"
	case "high":
		severityLabel = "High"
	default:
		severityLabel = "Low"
	}

	index, e := search.IndexBuilder("alert", time.Now().UTC().Format(time.RFC3339Nano))
	if e != nil {
		return false
	}

	pCfg, e := go_sdk.PluginCfg[utils.Config]("com.utmstack")
	if e != nil {
		return false
	}

	var filter = []map[string]interface{}{
		0: {
			"term": map[string]interface{}{
				"name.keyword": name,
			},
		},
		1: {
			"term": map[string]interface{}{
				"severityLabel.keyword": severityLabel,
			},
		},
	}

	if val, ok := details["SourceIP"]; ok {
		if val != "" {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"source.ip.keyword": val,
				},
			})
		}

	}

	if val, ok := details["SourceHost"]; ok {
		if val != "" {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"source.host.keyword": val,
				},
			})
		}

	}

	if val, ok := details["DestinationIP"]; ok {
		if val != "" {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"destination.ip.keyword": val,
				},
			})
		}

	}

	if val, ok := details["DestinationHost"]; ok {
		if val != "" {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"destination.host.keyword": val,
				},
			})
		}
	}

	if val, ok := details["SourceUser"]; ok {
		if val != "" {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"source.user.keyword": val,
				},
			})
		}

	}

	if val, ok := details["DestinationUser"]; ok {
		if val != "" {
			filter = append(filter, map[string]interface{}{
				"term": map[string]interface{}{
					"destination.user.keyword": val,
				},
			})
		}
	}

	var request = map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filter,
			},
		},
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		go_sdk.Logger().ErrorF("could not check existent alert: %v", err)
		return false
	}

	response, _, e := go_sdk.DoReq[map[string]interface{}](pCfg.Elasticsearch+"/"+index+"/_search", requestBytes, "POST", map[string]string{"Content-Type": "application/json"})
	if e != nil {
		return false
	}

	if hits, ok := response["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if val, ok := total["value"].(float64); ok {
				if val > 0 && val <= 25 {
					updated = true

					body := map[string]interface{}{
						"script": map[string]interface{}{
							"source": "ctx._source.logs.add(params.log)",
							"lang":   "painless",
							"params": map[string]interface{}{
								"log": details["id"],
							},
						},
					}

					bodyBytes, err := json.Marshal(body)
					if err != nil {
						go_sdk.Logger().ErrorF("could not update existent alert: %v", err)
						return false
					}

					_, _, e := go_sdk.DoReq[interface{}](pCfg.Elasticsearch+"/"+index+"/_update/"+response["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_id"].(string), bodyBytes, "POST", map[string]string{"Content-Type": "application/json"})
					if e != nil {
						return false
					}
				}
			}
		}
	}
	return updated
}

func NewAlert(id, name, severity, description, solution, category, tactic string, reference []string, dataType, dataSource string,
	details map[string]string) {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	sourceASN, _ := strconv.Atoi(details["SourceASN"])
	sourcePort, _ := strconv.Atoi(details["SourcePort"])
	sourceLat, _ := strconv.ParseFloat(details["SourceLat"], 64)
	sourceLon, _ := strconv.ParseFloat(details["SourceLon"], 64)
	sourceAccuracy, _ := strconv.Atoi(details["SourceAccuracyRadius"])

	destinationASN, _ := strconv.Atoi(details["DestinationASN"])
	destinationPort, _ := strconv.Atoi(details["DestinationPort"])
	destinationLat, _ := strconv.ParseFloat(details["DestinationLat"], 64)
	destinationLon, _ := strconv.ParseFloat(details["DestinationLon"], 64)
	destinationAccuracy, _ := strconv.Atoi(details["DestinationAccuracyRadius"])

	var severityN int
	var severityLabel string
	switch severity {
	case "low":
		severityN = 1
		severityLabel = "Low"
	case "medium":
		severityN = 2
		severityLabel = "Medium"
	case "high":
		severityN = 3
		severityLabel = "High"
	default:
		severityN = 1
		severityLabel = "Low"
	}

	a := AlertFields{
		Timestamp:     timestamp,
		ID:            id,
		Status:        1,
		StatusLabel:   "Automatic review",
		Name:          name,
		Category:      category,
		Severity:      severityN,
		SeverityLabel: severityLabel,
		Protocol:      details["Protocol"],
		Description:   description,
		Solution:      solution,
		Tactic:        tactic,
		Reference:     reference,
		DataType:      dataType,
		DataSource:    dataSource,
		Source: Side{
			User:           details["SourceUser"],
			Host:           details["SourceHost"],
			IP:             details["SourceIP"],
			Port:           sourcePort,
			Country:        details["SourceCountry"],
			CountryCode:    details["SourceCountryCode"],
			City:           details["SourceCity"],
			Coordinates:    [2]float64{sourceLat, sourceLon},
			ASN:            sourceASN,
			ASO:            details["SourceASO"],
			AccuracyRadius: sourceAccuracy,
		},
		Destination: Side{
			User:           details["DestinationUser"],
			Host:           details["DestinationHost"],
			IP:             details["DestinationIP"],
			Port:           destinationPort,
			Country:        details["DestinationCountry"],
			CountryCode:    details["DestinationCountryCode"],
			City:           details["DestinationCity"],
			Coordinates:    [2]float64{destinationLat, destinationLon},
			ASN:            destinationASN,
			ASO:            details["DestinationASO"],
			AccuracyRadius: destinationAccuracy,
		},
		Logs: []string{details["id"]},
	}

	pCfg, _ := go_sdk.PluginCfg[utils.Config]("com.utmstack")

	aj, err := json.Marshal(a)
	if err != nil {
		go_sdk.Logger().ErrorF("could not encode alert in JSON: %v", err)
		return
	}

	index, e := search.IndexBuilder("alert", time.Now().UTC().Format(time.RFC3339Nano))
	if e != nil {
		go_sdk.Logger().ErrorF("could not build index name: %v", err)
		return
	}

	url := pCfg.Elasticsearch + "/" + index + "/_doc"

	_, _, e = go_sdk.DoReq[interface{}](url, aj, "POST", map[string]string{"Content-Type": "application/json"})
	if e != nil {
		go_sdk.Logger().ErrorF("could not send alert to Elasticsearch: %v", err)
	}
}
