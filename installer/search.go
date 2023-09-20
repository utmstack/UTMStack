package main

import (
	"time"

	"github.com/levigross/grequests"
)

func InitOpenSearch() error {
	baseURL := "http://127.0.0.1:9200/"
	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		_, err := grequests.Get(baseURL+"_cluster/healt", &grequests.RequestOptions{
			Params: map[string]string{
				"wait_for_status": "green",
				"timeout":         "50s",
			},
		})

		if err != nil {
			if intent >= 10 {
				return err
			}
		} else {
			break
		}
	}

	_, err := grequests.Put(baseURL+"_snapshot/.utm_geoip", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"type": "fs",
			"settings": map[string]interface{}{
				"location": "/usr/share/opensearch/.utm-geoip/",
				"readonly": true,
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = grequests.Put(baseURL+"_index_template/utmstack_indexes", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"alert-*", "log-*", ".utm-*", ".utmstack-*"},
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

	_, err = grequests.Post(baseURL+"_snapshot/.utm_geoip/.utm-geoip/_restore?wait_for_completion=true", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"indices": ".utm-geoip",
		},
	})
	if err != nil {
		return err
	}

	return nil
}
