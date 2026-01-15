package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	twos "github.com/threatwinds/go-sdk/os"
	"github.com/tidwall/gjson"
)

var logs = make(chan string, 100*runtime.NumCPU())
var bulkQueue *twos.BulkQueue

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

		err := twos.Connect([]string{osUrl}, "", "")
		if err == nil {
			break
		}

		_ = catcher.Error("cannot connect to OpenSearch, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2
		} else {
			_ = catcher.Error("all retries failed when connecting to OpenSearch", err, nil)
			return
		}
	}

	bulkQueue = twos.NewBulkQueue(twos.BulkQueueConfig{
		FlushInterval: 10 * time.Second,
		OnError: func(failedItems []twos.BulkItem, err error) {
			_ = catcher.Error("failed to send logs to OpenSearch", err, map[string]any{
				"failedCount": len(failedItems),
			})
		},
	})

	if bulkQueue == nil {
		_ = catcher.Error("failed to create bulk queue", fmt.Errorf("OpenSearch connection not established"), nil)
		return
	}

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			for l := range logs {
				dataType := gjson.Get(l, "dataType").String()
				id := gjson.Get(l, "id").String()
				index := twos.BuildCurrentIndex("v11", "log", dataType)

				bulkQueue.AddWithID(index, id, l)
			}
		}()
	}
}
