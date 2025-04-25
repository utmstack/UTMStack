package main

import (
	"context"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/opensearch"
	"github.com/tidwall/gjson"
)

var logs = make(chan string, 100*runtime.NumCPU())

func addToQueue(l string) {
	logs <- l
}

func startQueue() {
	osUrl := plugins.PluginCfg("com.utmstack", false).Get("opensearch").String()

	err := opensearch.Connect([]string{osUrl})
	if err != nil {
		_ = catcher.Error("cannot connect to OpenSearch", err, nil)
		os.Exit(1)
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
