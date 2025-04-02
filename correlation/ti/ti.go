package ti

import (
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/correlation/correlation"
	"github.com/utmstack/UTMStack/correlation/utils"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Cache map[string]bool

var blockList map[string]string
var channel chan string
var cache Cache
var cacheLock *sync.RWMutex

func init() {
	blockList = make(map[string]string, 10000)
	channel = make(chan string, 10000)
	cache = make(Cache, 10000)
	cacheLock = &sync.RWMutex{}
	cache.AutoClean()
}

func blocked(log string) bool {
	log = strings.ToLower(log)

	exclusionList := []string{
		"block",
		"denied",
		"drop",
		"reject",
		"deny",
	}

	for _, e := range exclusionList {
		if strings.Contains(log, e) {
			return true
		}
	}

	return false
}

func (c *Cache) Add(k string) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	(*c)[k] = false
}

func (c *Cache) IsCached(k string) bool {
	cacheLock.RLock()
	defer cacheLock.RUnlock()
	_, ok := (*c)[k]
	return ok
}

func (c *Cache) AutoClean() {
	go func() {
		for {
			cacheLock.Lock()
			cache = make(Cache, 10000)
			cacheLock.Unlock()
			time.Sleep(8 * time.Hour)
		}
	}()
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

	numCPU := runtime.NumCPU()
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				log := <-channel

				sourceIp := gjson.Get(log, "logx.*.src_ip")
				destinationIp := gjson.Get(log, "logx.*.dest_ip")

				if !cache.IsCached(sourceIp.String()) {
					if _, ok := blockList[sourceIp.String()]; ok && !blocked(log) {
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

					}

					cache.Add(sourceIp.String())
				}

				if !cache.IsCached(destinationIp.String()) {
					if _, ok := blockList[destinationIp.String()]; ok && !blocked(log) {
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
					}

					cache.Add(destinationIp.String())
				}
			}
		}()
	}
}

func Enqueue(log string) {
	if len(channel) >= 10000 {
		return
	}

	channel <- log
}
