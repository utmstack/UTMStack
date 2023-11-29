package correlation

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"log"

	"github.com/google/uuid"
	"github.com/levigross/grequests"
	"github.com/utmstack/UTMStack/correlation/search"
	"github.com/utmstack/UTMStack/correlation/utils"
)

type Host struct {
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
	Source            Host           `json:"source"`
	Destination       Host           `json:"destination"`
	Logs              []string       `json:"logs"`
	Tags              []string       `json:"tags"`
	Notes             string         `json:"notes"`
	TagRulesApplied   []int          `json:"TagRulesApplied"`
}

func Alert(name, severity, description, solution, category, tactic string, reference []string, dataType, dataSource string,
	details map[string]string) {

	log.Printf("Reporting alert: %s", name)

	if !UpdateAlert(name, severity, details) {
		NewAlert(name, severity, description, solution, category, tactic, reference, dataType, dataSource,
			details)
	}
}

func UpdateAlert(name, severity string, details map[string]string) bool {
	var updated bool

	index, err := search.IndexBuilder("alert", time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		log.Printf("Could not build index name: %v", err)
		return true
	}

	cnf := utils.GetConfig()

	var filter = []map[string]interface{}{
		0: {
			"term": map[string]interface{}{
				"name.keyword": name,
			},
		},
		1: {
			"term": map[string]interface{}{
				"severityLabel.keyword": severity,
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

	response, err := grequests.Post(cnf.Elasticsearch+"/"+index+"/_search", &grequests.RequestOptions{
		JSON: request,
	})
	if err != nil {
		log.Printf("Could not check existent alert: %v", err)
		return false
	}

	resultStr := response.String()

	_ = response.Close()

	var resultObj map[string]interface{}

	err = json.Unmarshal([]byte(resultStr), &resultObj)

	if err != nil {
		log.Printf("Could not check existent alert: %v", err)
		return false
	}

	if hits, ok := resultObj["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if val, ok := total["value"].(float64); ok {
				if int(val) != 0 {
					updated = true
					r, err := grequests.Post(cnf.Elasticsearch+"/"+index+"/_update/"+resultObj["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_id"].(string), &grequests.RequestOptions{
						JSON: map[string]interface{}{
							"script": map[string]interface{}{
								"source": "ctx._source.logs.add(params.log)",
								"lang":   "painless",
								"params": map[string]interface{}{
									"log": details["id"],
								},
							},
						},
					})
					_ = r.Close()
					if err != nil {
						log.Printf("Could not update existent alert: %v", err)
						return false
					}
				}
			}
		}
	}
	return updated
}

func NewAlert(name, severity, description, solution, category, tactic string, reference []string, dataType, dataSource string,
	details map[string]string) {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	sourceASN, _ := strconv.Atoi(details["SourceASN"])
	sourcePort, _ := strconv.Atoi(details["SourcePort"])
	sourceLat, _ := strconv.ParseFloat(details["SourceLat"], 64)
	sourceLon, _ := strconv.ParseFloat(details["SourceLon"], 64)
	sourceAccuracyRadius, _ := strconv.Atoi(details["SourceAccuracyRadius"])
	sourceIsSatelliteProvider, _ := strconv.Atoi(details["SourceIsSatelliteProvider"])
	sourceIsAnonymousProxy, _ := strconv.Atoi(details["SourceIsAnonymousProxy"])

	var sisp bool
	if sourceIsSatelliteProvider == 1 {
		sisp = true
	}

	var siap bool
	if sourceIsAnonymousProxy == 1 {
		siap = true
	}

	destinationASN, _ := strconv.Atoi(details["DestinationASN"])
	destinationPort, _ := strconv.Atoi(details["DestinationPort"])
	destinationLat, _ := strconv.ParseFloat(details["DestinationLat"], 64)
	destinationLon, _ := strconv.ParseFloat(details["DestinationLon"], 64)
	destinationAccuracyRadius, _ := strconv.Atoi(details["DestinationAccuracyRadius"])
	destinationIsSatelliteProvider, _ := strconv.Atoi(details["DestinationIsSatelliteProvider"])
	destinationIsAnonymousProxy, _ := strconv.Atoi(details["DestinationIsAnonymousProxy"])

	var disp bool
	if destinationIsSatelliteProvider == 1 {
		disp = true
	}

	var diap bool
	if destinationIsAnonymousProxy == 1 {
		diap = true
	}

	var severityN int
	switch severity {
	case "Low":
		severityN = 1
	case "Medium":
		severityN = 2
	case "High":
		severityN = 3
	}

	a := AlertFields{
		Timestamp:     timestamp,
		ID:            uuid.NewString(),
		Status:        1,
		StatusLabel:   "Automatic review",
		Name:          name,
		Category:      category,
		Severity:      severityN,
		SeverityLabel: severity,
		Protocol:      details["Protocol"],
		Description:   description,
		Solution:      solution,
		Tactic:        tactic,
		Reference:     reference,
		DataType:      dataType,
		DataSource:    dataSource,
		Source: Host{
			User:                details["SourceUser"],
			Host:                details["SourceHost"],
			IP:                  details["SourceIP"],
			Port:                sourcePort,
			Country:             details["SourceCountry"],
			CountryCode:         details["SourceCountryCode"],
			City:                details["SourceCity"],
			Coordinates:         [2]float64{sourceLat, sourceLon},
			AccuracyRadius:      sourceAccuracyRadius,
			ASN:                 sourceASN,
			ASO:                 details["SourceASO"],
			IsSatelliteProvider: sisp,
			IsAnonymousProxy:    siap,
		},
		Destination: Host{
			User:                details["DestinationUser"],
			Host:                details["DestinationHost"],
			IP:                  details["DestinationIP"],
			Port:                destinationPort,
			Country:             details["DestinationCountry"],
			CountryCode:         details["DestinationCountryCode"],
			City:                details["DestinationCity"],
			Coordinates:         [2]float64{destinationLat, destinationLon},
			AccuracyRadius:      destinationAccuracyRadius,
			ASN:                 destinationASN,
			ASO:                 details["DestinationASO"],
			IsSatelliteProvider: disp,
			IsAnonymousProxy:    diap,
		},
		Logs: []string{details["id"]},
	}

	aj, err := json.Marshal(a)
	body := strings.NewReader(string(aj))
	if err == nil {
		index, err := search.IndexBuilder("alert", time.Now().UTC().Format(time.RFC3339Nano))
		if err == nil {
			cnf := utils.GetConfig()
			url := cnf.Elasticsearch + "/" + index + "/_doc"
			_, err := utils.DoPost(url, "application/json", body)
			if err != nil {
				log.Printf("Could not send alert to Elasticsearch: %v", err)
			}
		} else {
			log.Printf("Could not build index name: %v", err)
		}
	} else {
		log.Printf("Could not encode alert in JSON: %v", err)
	}
	time.Sleep(3 * time.Second)
}
