package statistics

import (
	"sync"
	"time"

	"github.com/utmstack/UTMStack/correlation/sqldb"
)

var stats = make(map[string]Stat)
var statsMutex = &sync.Mutex{}

type Stat struct{
	Id string `json:"id"`
	Source string `json:"source"`
	Type string `json:"type"`
	Count int64 `json:"count"`
}

func GetStats() map[string]Stat {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	
	return stats
}

func One(t, s string) {
	id := t + s
	
	statsMutex.Lock()
	defer statsMutex.Unlock()

	if _, ok := stats[id]; !ok {
		stats[id] = Stat{
			Id:     id,
			Source: s,
			Type:   t,
			Count:  1,
		}
	} else {
		stats[id] = Stat{
			Id:     id,
			Source: s,
			Type:   t,
			Count:  stats[id].Count + 1,
		}
	}
}

func Update() {
	for {
		statsMutex.Lock()
		tmp := stats
		stats = nil
		stats = make(map[string]Stat)
		statsMutex.Unlock()
		for _, s := range tmp {
			sqldb.UpdateStatistics(
				s.Id,
				s.Source,
				s.Type,
				s.Count,
			)
		}
		time.Sleep(10 * time.Minute)
	}
}
