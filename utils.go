package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/levigross/grequests"
)

func runEnvCmd(env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Executing command:", command, "with args:", arg)
	return cmd.Run()
}

func runCmd(command string, arg ...string) error {
	return runEnvCmd([]string{}, command, arg...)
}

func checkOutput(command string, arg ...string) string {
	log.Println("Executing command:", command, "with args:", arg)
	out, err := exec.Command(command, arg...).Output()
	check(err)
	return strings.TrimSpace(string(out))
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func initializeElastic(secret string) {
	// wait for elastic to be ready
	baseURL := "http://127.0.0.1:9200/"
	for {
		log.Println("Waiting for the search engine")
		time.Sleep(50 * time.Second)

		_, err := grequests.Get(baseURL + "_cluster/healt", &grequests.RequestOptions{
			Params: map[string]string{
				"wait_for_status": "green",
				"timeout":         "50s",
			},
		})

		if err == nil {
			break
		}

		log.Println("The search engine is taking more than expected to run, retrying...")
	}

	// configure elastic
	initialIndex := "index-" + secret + "-000001"
	// create alias
	_, err := grequests.Post(baseURL + "_aliases", &grequests.RequestOptions{
		JSON: map[string][]interface{}{
			"actions": []interface{}{
				map[string]interface{}{
					"add": map[string]string{
						"index": initialIndex,
						"alias": "index-" + secret,
					},
				},
			},
		},
	})
	check(err)

	// create main index template
	_, err = grequests.Put(baseURL + "_template/main_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"index-*"},
			"settings": map[string]interface{}{
				"index.mapping.total_fields.limit": 10000,
				"opendistro.index_state_management.policy_id": "main_index_policy",
				"opendistro.index_state_management.rollover_alias": "index-" + secret,
				"number_of_shards": 3,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// create template for generic index
	_, err = grequests.Put(baseURL + "_template/generic_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"generic-*"},
			"settings": map[string]interface{}{
				"index.mapping.total_fields.limit": 10000,
				"number_of_shards": 1,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// create template for DC index
	_, err = grequests.Put(baseURL + "_template/dc_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"dc-*"},
			"settings": map[string]interface{}{
				"number_of_shards": 1,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// create utmstack_index template
	_, err = grequests.Put(baseURL + "_template/utmstack_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"utmstack-*"},
			"settings": map[string]interface{}{
				"number_of_shards": 1,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// crate template for geolocation index
	_, err = grequests.Put(baseURL + "_template/utm_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"index_patterns": []string{"utm-*"},
			"settings": map[string]interface{}{
				"number_of_shards": 1,
				"number_of_replicas": 0,
			},
		},
	})
	check(err)

	// enable snapshots
	_, err = grequests.Put(baseURL + "_snapshot/main_index", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"type": "fs",
			"settings": map[string]interface{}{
				"location": "main_index",
			},
		},
	})
	check(err)

	// create ISM policy
	_, err = grequests.Put(baseURL + "_opendistro/_ism/policies/main_index_policy", &grequests.RequestOptions{
		JSON: map[string]interface{}{
			"policy": map[string]interface{}{
				"description": "Main Index Lifecycle",
				"default_state": "ingest",
				"states": []interface{}{
					map[string]interface{}{
						"name": "ingest",
						"actions": []interface{}{
							map[string]interface{}{
								"rollover": map[string]interface{}{
									"min_doc_count": 30000000,
									"min_size": "15gb",
								},
							},
						},
						"transitions": []interface{}{
							map[string]string{
								"state_name": "search",
							},
						},
					},
					map[string]interface{}{
						"name": "search",
						"actions": []interface{}{
							map[string]interface{}{
								"snapshot": map[string]string{
									"repository": "main_index",
									"snapshot": "incremental",
								},
							},
						},
						"transitions": []interface{}{
							map[string]interface{}{
								"state_name": "read",
								"conditions": map[string]string{
									"min_index_age": "30d",
								},
							},
						},
					},
					map[string]interface{}{
						"name": "read",
						"actions": []interface{}{
							map[string]interface{}{
								"force_merge": map[string]interface{}{
									"max_num_segments": 1,
								},
							},
							map[string]interface{}{
								"snapshot": map[string]interface{}{
									"repository": "main_index",
									"snapshot": "incremental",
								},
							},
						},
						"transitions": []interface{}{},
					},
				},
			},
		},
	})
	check(err)

	// create initial index
	_, err = grequests.Put(baseURL + initialIndex, &grequests.RequestOptions{
		JSON: map[string]interface{}{},
	})
	check(err)
}
