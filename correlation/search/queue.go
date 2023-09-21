package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/correlation/statistics"
	"github.com/utmstack/UTMStack/correlation/utils"
	"github.com/tidwall/gjson"
)

var ndMutex = &sync.Mutex{}
var nd string

var logs = make(chan string, 100)

func AddToQueue(l string) {
	logs <- l
}

func NDGenerator() {
	for {
		l := <-logs
		var cl *bytes.Buffer = new(bytes.Buffer)
		start := time.Now()
		dataType := gjson.Get(l, "dataType").String()
		dataSource := gjson.Get(l, "dataSource").String()

		statistics.One(dataType, dataSource)

		timestamp := gjson.Get(l, "@timestamp").String()
		id := gjson.Get(l, "id").String()

		index, err := IndexBuilder("log-"+dataType, timestamp)
		if err != nil {
			h.Error("Error trying to build index name: %v", err)
			continue
		}

		json.Compact(cl, []byte(l))

		ndMutex.Lock()
		nd += fmt.Sprintf(`{"index": {"_index": "%s", "_id": "%s"}}`, index, id) + "\n" + cl.String() + "\n"
		ndMutex.Unlock()

		h.Debug("Generate ND took: %s", time.Since(start))
	}
}

func Bulk() {
	for {
		url := fmt.Sprintf("%s/_bulk", utils.GetConfig().Elasticsearch)
		ndMutex.Lock()
		tmp := nd
		nd = ""
		ndMutex.Unlock()
		body, err := utils.DoPost(url, "application/x-ndjson", strings.NewReader(tmp))
		if err != nil {
			h.FatalError("Could not send logs to Elasticsearch: %v. %s", err, body)
		}
		time.Sleep(1 * time.Second)
	}
}
