package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/utmstack/UTMStack/plugins/soc-ai/configurations"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

func ElasticQuery(index string, query interface{}, op string) error {
	var endp string
	switch op {
	case "create":
		endp = configurations.ELASTIC_DOC_ENDPOINT
	case "update":
		endp = configurations.ELASTIC_UPDATE_BY_QUERY_ENDPOINT
	}
	url := configurations.GetOpenSearchHost() + ":" + configurations.GetOpenSearchPort() + "/" + index + endp
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error marshalling query: %v", err)
	}

	resp, statusCode, err := utils.DoReq(url, queryBytes, "POST", headers, configurations.HTTP_TIMEOUT)
	if err != nil || (statusCode != http.StatusOK && statusCode != http.StatusCreated) {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	return nil
}

func ElasticSearch(index, field, value string) ([]byte, error) {
	config := configurations.GetPluginConfig()
	url := config.Backend + configurations.API_ALERT_ENDPOINT + configurations.API_ALERT_INFO_PARAMS + index
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": config.InternalKey,
	}

	body := schema.SearchDetailsRequest{{Field: field, Operator: "IS", Value: value}}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling body: %v", err)
	}

	resp, statusCode, err := utils.DoReq(url, bodyBytes, "POST", headers, configurations.HTTP_TIMEOUT)
	if err != nil || statusCode != http.StatusOK {
		return nil, fmt.Errorf("error while doing request for get Alert Details: %v: %s", err, string(resp))
	}

	return resp, nil
}

func IndexStatus(id, status, op string) error {
	doc := schema.GPTAlertResponse{
		Timestamp:  time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00"),
		Status:     status,
		ActivityID: id,
	}

	if op == "update" {
		query, err := schema.ConvertGPTResponseToUpdateQuery(doc)
		if err != nil {
			return fmt.Errorf("error while converting response to update query: %v", err)
		}
		return ElasticQuery(configurations.SOC_AI_INDEX, query, op)
	} else {
		return ElasticQuery(configurations.SOC_AI_INDEX, doc, op)
	}
}

func CreateIndexIfNotExist(index string) error {
	url := configurations.GetOpenSearchHost() + ":" + configurations.GetOpenSearchPort() + "/" + index
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, statusCode, err := utils.DoReq(url, nil, "HEAD", headers, configurations.HTTP_TIMEOUT)
	if err != nil {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	if statusCode == 404 {
		resp, statusCode, err = utils.DoReq(url, nil, "PUT", headers, configurations.HTTP_TIMEOUT)
		if err != nil || (statusCode != http.StatusOK && statusCode != http.StatusCreated) {
			return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
		}
	}

	return nil
}
