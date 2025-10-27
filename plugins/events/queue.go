package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/threatwinds/go-sdk/opensearch"
	"github.com/tidwall/gjson"
)

var logs = make(chan string, 100*runtime.NumCPU())

func addToQueue(l string) {
	if len(logs) >= 100*runtime.NumCPU() {
		_ = catcher.Error("cannot enqueue log", fmt.Errorf("queue is full"), map[string]any{
			"queue": "logs",
		})

		return
	}

	logs <- l
}

func startQueue() {
	// Retry logic for connecting to OpenSearch
	maxRetries := 3
	retryDelay := 2 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		osUrl := plugins.PluginCfg("com.utmstack", false).Get("opensearch").String()

		err := opensearch.Connect([]string{osUrl})
		if err == nil {
			break
		}

		_ = catcher.Error("cannot connect to OpenSearch, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when connecting to OpenSearch", err, nil)
			return
		}
	}

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			var ndMutex = &sync.Mutex{}
			var nd = make([]opensearch.BulkItem, 0, 10)

			go func() {
				for {
					if len(nd) == 0 {
						time.Sleep(10 * time.Second)
						continue
					}

					ndMutex.Lock()

					err := opensearch.Bulk(context.Background(), nd)
					if err != nil {
						_ = catcher.Error("failed to send logs to OpenSearch", err, nil)
					}

					nd = make([]opensearch.BulkItem, 0, 10)

					ndMutex.Unlock()
				}
			}()

			for {
				l := <-logs

				dataType := gjson.Get(l, "dataType").String()
				id := gjson.Get(l, "id").String()
				index := opensearch.BuildCurrentIndex("v11", "log", dataType)

				ndMutex.Lock()

				nd = append(nd, opensearch.BulkItem{
					Index:  index,
					Id:     id,
					Body:   []byte(l),
					Action: "index",
				})

				ndMutex.Unlock()
			}
		}()
	}
}
