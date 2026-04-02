package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/tidwall/gjson"
)

var logs = make(chan string, 100*runtime.NumCPU())

func addToQueue(l string) {
	if len(logs) >= 100*runtime.NumCPU() {
		_ = catcher.Error("cannot enqueue log", fmt.Errorf("queue is full"), map[string]any{
			"process": "plugin_com.utmstack.events",
			"queue":   "logs",
		})

		return
	}

	logs <- l
}

func startQueue() {
	// Retry logic for connecting to OpenSearch
	maxRetries := 3
	retryDelay := 2 * time.Second

	cfg := plugins.PluginCfg("org.opensearch").Get("opensearch")
	host := cfg.Get("host").String()
	port := cfg.Get("port").String()
	user := cfg.Get("user").String()
	password := cfg.Get("password").String()
	osUrl := "https://" + host + ":" + port

	for retry := 0; retry < maxRetries; retry++ {
		err := sdkos.Connect([]string{osUrl}, user, password)
		if err == nil {
			break
		}

		_ = catcher.Error("cannot connect to OpenSearch, retrying", err, map[string]any{
			"process":    "plugin_com.utmstack.events",
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		} else {
			// If all retries failed, log the error and return
			_ = catcher.Error("all retries failed when connecting to OpenSearch", err, map[string]any{"process": "plugin_com.utmstack.events"})
			return
		}
	}

	queue := sdkos.NewBulkQueue("plugin_com.utmstack.events", sdkos.BulkQueueConfig{
		FlushInterval:  10 * time.Second,
		FlushThreshold: 50,
		MaxRetries:     0,
		RetryDelay:     time.Second,
	})

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				l := <-logs

				dataType := gjson.Get(l, "dataType").String()
				id := gjson.Get(l, "id").String()
				index := sdkos.BuildCurrentDayIndex("v11", "log", dataType)

				queue.AddItem(sdkos.BulkItem{
					Index:      index,
					DocumentID: id,
					Document:   []byte(l),
					Operation:  "index",
				})
			}
		}()
	}
}
