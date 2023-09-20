package statistics

import (
	"sync"
	"time"

	"github.com/AtlasInsideCorp/UTMStackCorrelationEngine/sqldb"
)

var stats = make(map[string]interface{})
var statsMutex = &sync.Mutex{}

func GetStats() map[string]interface{} {
	return stats
}

func One(t, s string) {
	id := t + s
	statsMutex.Lock()
	if stats[id] == nil {
		stats[id] = map[string]interface{}{
			"id":     id,
			"source": s,
			"type":   t,
			"count":  1,
		}
	} else {
		stats[id] = map[string]interface{}{
			"id":     id,
			"source": s,
			"type":   t,
			"count":  stats[id].(map[string]interface{})["count"].(int) + 1,
		}
	}
	statsMutex.Unlock()
}

func Update() {
	for {
		statsMutex.Lock()
		tmp := stats
		stats = nil
		stats = make(map[string]interface{})
		statsMutex.Unlock()
		for _, s := range tmp {
			sqldb.UpdateStatistics(
				s.(map[string]interface{})["id"].(string),
				s.(map[string]interface{})["source"].(string),
				s.(map[string]interface{})["type"].(string),
				s.(map[string]interface{})["count"].(int),
			)
		}
		time.Sleep(1 * time.Second)
	}
}
