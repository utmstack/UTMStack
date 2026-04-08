package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

func ElasticQuery(index string, query interface{}, op string) error {
	var url string
	var method string
	cnf := config.GetConfig()

	switch op {
	case "create":
		if gptResp, ok := query.(schema.GPTAlertResponse); ok && gptResp.ActivityID != "" {
			url = fmt.Sprintf("%s/%s/_doc/%s", cnf.OpensearchURL, index, gptResp.ActivityID)
			method = "PUT"
		} else {
			url = fmt.Sprintf("%s/%s/_doc", cnf.OpensearchURL, index)
			method = "POST"
		}
	case "update":
		if gptResp, ok := query.(schema.GPTAlertResponse); ok && gptResp.ActivityID != "" {
			url = fmt.Sprintf("%s/%s/_doc/%s", cnf.OpensearchURL, index, gptResp.ActivityID)
			method = "PUT"
		} else {
			url = fmt.Sprintf("%s/%s%s", cnf.OpensearchURL, index, config.ELASTIC_UPDATE_BY_QUERY_ENDPOINT)
			method = "POST"
		}
	default:
		return fmt.Errorf("unsupported operation: %s", op)
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("error marshalling query: %v", err)
	}

	resp, statusCode, err := utils.DoOpenSearchReq(url, queryBytes, method, headers, cnf.OpensearchUser, cnf.OpensearchPassword, config.HTTP_TIMEOUT)
	if err != nil || (statusCode != http.StatusOK && statusCode != http.StatusCreated) {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	return nil
}

// ElasticSearch performs a search with default pagination (uses top=10000)
func ElasticSearch(index, field, value string) ([]byte, error) {
	return ElasticSearchWithLimit(index, field, value, 10000)
}

// ElasticSearchWithLimit performs a search with a specified result limit
func ElasticSearchWithLimit(index, field, value string, limit int) ([]byte, error) {
	cnf := config.GetConfig()
	// Build URL with explicit limit parameter
	url := fmt.Sprintf("%s%s?page=0&size=%d&top=%d&indexPattern=%s",
		cnf.Backend, config.API_ALERT_ENDPOINT, limit, limit, index)

	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": cnf.InternalKey,
	}

	body := schema.SearchDetailsRequest{{Field: field, Operator: "IS", Value: value}}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshalling body: %v", err)
	}

	resp, statusCode, err := utils.DoReq(url, bodyBytes, "POST", headers, config.HTTP_TIMEOUT)
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
		// For update operations, pass the GPTAlertResponse directly to use _doc endpoint
		return ElasticQuery(config.SOC_AI_INDEX, doc, op)
	} else {
		// Handle create operations with index creation retry logic
		err := ElasticQuery(config.SOC_AI_INDEX, doc, op)
		if err != nil {
			// Check if the error is due to index not existing
			if strings.Contains(err.Error(), "index_not_found_exception") || strings.Contains(err.Error(), "no such index") {
				// Try to create the index first
				if createErr := CreateIndexIfNotExist(config.SOC_AI_INDEX); createErr != nil {
					return fmt.Errorf("error creating document in elastic: %v (failed to create index: %v)", err, createErr)
				}

				// Retry the create operation
				if retryErr := ElasticQuery(config.SOC_AI_INDEX, doc, op); retryErr != nil {
					return fmt.Errorf("error creating document in elastic after index creation: %v", retryErr)
				}
			} else {
				return fmt.Errorf("error creating document in elastic: %v", err)
			}
		}
		return nil
	}
}

func CreateIndexIfNotExist(index string) error {
	cnf := config.GetConfig()
	url := fmt.Sprintf("%s/%s", cnf.OpensearchURL, index)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, statusCode, err := utils.DoOpenSearchReq(url, nil, "HEAD", headers, cnf.OpensearchUser, cnf.OpensearchPassword, config.HTTP_TIMEOUT)
	if err != nil {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	if statusCode == 404 {
		resp, statusCode, err = utils.DoOpenSearchReq(url, nil, "PUT", headers, cnf.OpensearchUser, cnf.OpensearchPassword, config.HTTP_TIMEOUT)
		if err != nil || (statusCode != http.StatusOK && statusCode != http.StatusCreated) {
			return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
		}
	}

	return nil
}
