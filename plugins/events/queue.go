package main

import (
	"context"
	"fmt"
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
	elasticUrl := plugins.PluginCfg("com.utmstack", false).Get("elasticsearch").String()

	err := opensearch.Connect([]string{elasticUrl})
	if err != nil {
		_ = catcher.Error("cannot connect to Elasticsearch/OpenSearch", err, nil)
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
						_ = catcher.Error("failed to send logs to Elasticsearch/OpenSearch", err, nil)
					}

					nd = make([]opensearch.BulkItem, 0, 10)
					ndMutex.Unlock()
				}
			}()

			for {
				l := <-logs

				dataType := gjson.Get(l, "dataType").String()

				id := gjson.Get(l, "id").String()

				index := indexBuilder(fmt.Sprintf("v11-log-%s", dataType), time.Now().UTC())

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
