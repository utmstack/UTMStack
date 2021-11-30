package utils

import (
	"time"

	"github.com/levigross/grequests"
)

func initializeElastic() error {
	// wait for elastic to be ready
	baseURL := "http://127.0.0.1:9200/"
	for {
		time.Sleep(2 * time.Minute)

		_, err := grequests.Get(baseURL+"_cluster/healt", &grequests.RequestOptions{
			Params: map[string]string{
				"wait_for_status": "green",
				"timeout":         "50s",
			},
		})

		if err == nil {
			break
		}
	}

	_, err := grequests.Put(baseURL+"_snapshot/utm_geoip", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"type": "fs",
			"settings": map[string]interface{}{
				"location": "utm-geoip",
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = grequests.Put(baseURL+"_index_template/utmstack_indexes", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"alert-*", "log-*", "dc-*", ".utm-*", ".utmstack-*"},
			"template": map[string]interface{}{
				"settings": map[string]interface{}{
					"index.number_of_shards":           1,
					"index.number_of_replicas":         0,
					"index.mapping.total_fields.limit": 50000,
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// restore geoip snapshot
	_, err = grequests.Post(baseURL+"_snapshot/utm_geoip/utm-geoip/_restore?wait_for_completion=false", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"indices":              "utm-geoip",
			"ignore_unavailable":   true,
			"include_global_state": false,
			"rename_pattern":       "utm-geoip",
			"rename_replacement":   ".utm-geoip",
		},
	})
	if err != nil {
		return err
	}

	return nil
}
