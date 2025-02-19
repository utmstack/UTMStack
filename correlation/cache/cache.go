package cache

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/correlation/rules"
	"github.com/utmstack/UTMStack/correlation/utils"
)

const bufferSize int = 1000000

var storageMutex = &sync.RWMutex{}

var storage []string

func Status() {
	for {
		log.Printf("Logs in cache: %v", len(storage))
		if len(storage) != 0 {
			est := gjson.Get(storage[0], "@timestamp").String()
			log.Printf("Old document in cache: %s", est)
		}
		time.Sleep(60 * time.Second)
	}

}

func Search(allOf []rules.AllOf, oneOf []rules.OneOf, seconds int64) []string {
	var elements []string
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	cToBreak := 0
	ait := time.Now().UTC().Unix() - func() int64 {
		switch seconds {
		case 0:
			return 60
		default:
			return seconds
		}
	}()
	for i := len(storage) - 1; i >= 0; i-- {
		est := gjson.Get(storage[i], "@timestamp").String()
		eit, err := time.Parse(time.RFC3339Nano, est)
		if err != nil {
			log.Printf("Could not parse @timestamp: %v", err)
			continue
		}
		if eit.Unix() < ait {
			if cToBreak <= 5 {
				cToBreak++
			} else {
				break
			}
		} else {
			cToBreak = 0
			var allCatch bool
			var oneCatch bool
			for _, of := range oneOf {
				oneCatch = evalElement(storage[i], of.Field, of.Operator, of.Value)
				if oneCatch {
					break
				}
			}
			for _, af := range allOf {
				allCatch = evalElement(storage[i], af.Field, af.Operator, af.Value)
				if !allCatch {
					break
				}
			}
			if (len(allOf) == 0 || allCatch) && (len(oneOf) == 0 || oneCatch) {
				elements = append(elements, storage[i])
			}
		}
	}

	return elements
}

var logs = make(chan string, bufferSize)

func AddToCache(l string) {
	if len(logs) == bufferSize {
		log.Printf("Buffer is full, you could be lossing events")
		return
	}
	logs <- l
}

func ProcessQueue() {
	numCPU := runtime.NumCPU() * 2
	for i := 0; i < numCPU; i++ {
		go func() {
			for {
				l := <-logs
				storageMutex.Lock()
				storage = append(storage, l)
				storageMutex.Unlock()
			}
		}()
	}
}

func Clean() {
	for {
		var clean bool

		if len(storage) > 1 {
			if utils.AssignedMemory >= 80 {
				clean = true
			} else {
				old := gjson.Get(storage[0], "@timestamp").String()
				oldTime, err := time.Parse(time.RFC3339Nano, old)
				if err != nil {
					log.Printf("Could not parse old log timestamp. Cleaning up")
					clean = true
				}

				hourAgo := time.Now().UTC().Add(-1 * time.Hour)

				if oldTime.Before(hourAgo) {
					clean = true
				}
			}
		}

		if clean {
			storageMutex.Lock()
			storage = storage[1:]
			storageMutex.Unlock()
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
