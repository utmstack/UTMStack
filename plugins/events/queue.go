package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	go_sdk "github.com/threatwinds/go-sdk"
	go_sdk_os "github.com/threatwinds/go-sdk/opensearch"
	"github.com/tidwall/gjson"
)

var logs = make(chan string, 100*runtime.NumCPU())

func addToQueue(l string) {
	logs <- l
}

func init() {
	pCfg, e := go_sdk.PluginCfg[Config]("com.utmstack")
	if e != nil {
		return
	}

	go_sdk_os.Connect([]string{pCfg.Elasticsearch})

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			var ndMutex = &sync.Mutex{}
			var nd = make([]go_sdk_os.BulkItem, 0, 10)

			go func() {
				for {
					if len(nd) == 0 {
						time.Sleep(10 * time.Second)
						continue
					}

					ndMutex.Lock()
					err := go_sdk_os.Bulk(context.Background(), nd)
					if err != nil {
						go_sdk.Logger().ErrorF(err.Error())
					}

					nd = make([]go_sdk_os.BulkItem, 0, 10)
					ndMutex.Unlock()
				}
			}()

			for {
				l := <-logs

				dataType := gjson.Get(l, "dataType").String()

				id := gjson.Get(l, "id").String()

				index := indexBuilder(fmt.Sprintf("v11-log-%s", dataType), time.Now().UTC())

				ndMutex.Lock()
				nd = append(nd, go_sdk_os.BulkItem{
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
