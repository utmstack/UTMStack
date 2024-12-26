package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	gosdk "github.com/threatwinds/go-sdk"
	ossdk "github.com/threatwinds/go-sdk/opensearch"
	"github.com/tidwall/gjson"
)

var logs = make(chan string, 100*runtime.NumCPU())

func addToQueue(l string) {
	logs <- l
}

func startQueue() {
	elasticUrl := gosdk.PluginCfg("com.utmstack", false).Get("elasticsearch").String()

	err := ossdk.Connect([]string{elasticUrl})
	if err != nil {
		_ = gosdk.Error("cannot connect to Elasticsearch/OpenSearch", err, nil)
		os.Exit(1)
	}

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			var ndMutex = &sync.Mutex{}
			var nd = make([]ossdk.BulkItem, 0, 10)

			go func() {
				for {
					if len(nd) == 0 {
						time.Sleep(10 * time.Second)
						continue
					}

					ndMutex.Lock()
					err := ossdk.Bulk(context.Background(), nd)
					if err != nil {
						_ = gosdk.Error("failed to send logs to Elasticsearch/OpenSearch", err, nil)
					}

					nd = make([]ossdk.BulkItem, 0, 10)
					ndMutex.Unlock()
				}
			}()

			for {
				l := <-logs

				dataType := gjson.Get(l, "dataType").String()

				id := gjson.Get(l, "id").String()

				index := indexBuilder(fmt.Sprintf("v11-log-%s", dataType), time.Now().UTC())

				ndMutex.Lock()
				nd = append(nd, ossdk.BulkItem{
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
