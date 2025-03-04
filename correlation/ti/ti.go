package ti

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/correlation/correlation"
	"github.com/utmstack/UTMStack/correlation/utils"
	"runtime"
	"strings"
)

var blockList map[string]string
var channel chan string

func init() {
	blockList = make(map[string]string, 10000)
	channel = make(chan string, 10000)
}

func IsBlocklisted() {
	saveFields := []utils.SavedField{
		{
			Field: "logx.*.proto",
			Alias: "Protocol",
		},
		{
			Field: "logx.*.src_ip",
			Alias: "SourceIP",
		},
		{
			Field: "logx.*.dest_ip",
			Alias: "DestinationIP",
		},
		{
			Field: "logx.*.src_port",
			Alias: "SourcePort",
		},
		{
			Field: "logx.*.dest_port",
			Alias: "DestinationPort",
		},
	}

	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				log := <-channel

				for key, value := range blockList {
					var stop bool

					switch value {
					case "IP":
						sourceIp := gjson.Get(log, "logx.*.src_ip")
						destinationIp := gjson.Get(log, "logx.*.dest_ip")

						if sourceIp.String() == key {
							correlation.Alert(
								"Connection attempt from a malicious IP",
								"Low",
								"A blocklisted element has been identified in the logs. Further investigation is recommended.",
								"",
								"Threat Intelligence",
								"",
								[]string{"https://threatwinds.com"},
								gjson.Get(log, "dataType").String(),
								gjson.Get(log, "dataSource").String(),
								utils.ExtractDetails(saveFields, log),
							)

							stop = true
						}

						if destinationIp.String() == key {
							correlation.Alert(
								"Connection attempt to a malicious IP",
								"High",
								"A blocklisted element has been identified in the logs. Further investigation is recommended.",
								"",
								"Threat Intelligence",
								"",
								[]string{"https://threatwinds.com"},
								gjson.Get(log, "dataType").String(),
								gjson.Get(log, "dataSource").String(),
								utils.ExtractDetails(saveFields, log),
							)

							stop = true
						}
					}

					if stop {
						break
					}

					if strings.Contains(log, key) {
						correlation.Alert(
							fmt.Sprintf("Malicious %s found in log: %s", value, key),
							"Low",
							"A blocklisted element has been identified in the logs. Further investigation is recommended.",
							"",
							"Threat Intelligence",
							"",
							[]string{"https://threatwinds.com"},
							gjson.Get(log, "dataType").String(),
							gjson.Get(log, "dataSource").String(),
							utils.ExtractDetails(saveFields, log),
						)

						break
					}
				}
			}
		}()
	}
}

func Enqueue(log string) {
	channel <- log
}
