package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/helpers"
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/plugins/events/statistics"
	"github.com/utmstack/UTMStack/plugins/events/utils"
)

var logs = make(chan string, 10000)

func AddToQueue(l string) {
	logs <- l
}

func init() {
	pCfg, e := helpers.PluginCfg[utils.Config]("com.utmstack")
	if e != nil {
		return
	}

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			var ndMutex = &sync.Mutex{}
			var nd string

			go func() {
				for {
					if nd == "" {
						time.Sleep(10 * time.Second)
						continue
					}

					url := fmt.Sprintf("%s/_bulk", pCfg.Elasticsearch)

					ndMutex.Lock()
					tmp := nd
					nd = ""
					ndMutex.Unlock()

					if tmp == "" {
						time.Sleep(10 * time.Second)
						continue
					}

					bodyBuffer := bytes.NewBufferString(tmp)

					bJson := bodyBuffer.Bytes()

					_, _, e := helpers.DoReq[interface{}](url, bJson, "POST", map[string]string{"Content-Type": "application/json"})
					if e != nil {
						time.Sleep(10 * time.Second)
						continue
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

				index, e := IndexBuilder("log-"+dataType, timestamp)
				if e != nil {
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