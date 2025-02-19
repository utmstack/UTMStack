package search

import (
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/correlation/utils"
)

func Search(query string) []string {
	var result []string
	cnf := utils.GetConfig()
	url := fmt.Sprintf("%s/log-*/_search", cnf.Elasticsearch)
	cnn, err := utils.DoPost(url, "application/json", strings.NewReader(query))
	if err != nil {
		log.Printf("Could not get logs from Elasticsearch: %v", err)
	} else {
		hits := gjson.Get(string(cnn), "hits.hits").Array()
		for _, hit := range hits {
			l := gjson.Get(hit.String(), "_source")
			result = append(result, l.String())
		}
	}
	return result
}
