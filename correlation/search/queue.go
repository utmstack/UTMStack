package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/correlation/statistics"
	"github.com/utmstack/UTMStack/correlation/utils"
)

var logs = make(chan string, 10000)

func AddToQueue(l string) {
	logs <- l
}

func ProcessQueue() {
	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			var ndMutex = &sync.Mutex{}
			var nd string

			go func() {
				for {
					if nd != "" {
						url := fmt.Sprintf("%s/_bulk", utils.GetConfig().Elasticsearch)
						ndMutex.Lock()
						tmp := nd
						nd = ""
						ndMutex.Unlock()
						
						body, err := utils.DoPost(url, "application/x-ndjson", strings.NewReader(tmp))
						if err != nil {
							log.Fatalf("Could not send logs to Elasticsearch: %v. %s", err, body)
						}
					}
					time.Sleep(1 * time.Second)
				}
			}()

			for {
				l := <-logs
				var cl *bytes.Buffer = new(bytes.Buffer)
				dataType := gjson.Get(l, "dataType").String()
				dataSource := gjson.Get(l, "dataSource").String()

				statistics.One(dataType, dataSource)

				timestamp := gjson.Get(l, "@timestamp").String()
				id := gjson.Get(l, "id").String()

				index, err := IndexBuilder("log-"+dataType, timestamp)
				if err != nil {
					log.Printf("Error trying to build index name: %v", err)
					continue
				}

				json.Compact(cl, []byte(l))

				ndMutex.Lock()
				nd += fmt.Sprintf(`{"index": {"_index": "%s", "_id": "%s"}}`, index, id) + "\n" + cl.String() + "\n"
				ndMutex.Unlock()
			}
		}()
	}
}

