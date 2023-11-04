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

var cacheStorageMutex = &sync.RWMutex{}

var CacheStorage []string

func Status() {
	for {
		log.Printf("Logs in cache: %v", len(CacheStorage))
		if len(CacheStorage) != 0 {
			est := gjson.Get(CacheStorage[0], "@timestamp").String()
			log.Printf("Old document in cache: %s", est)
		}
		time.Sleep(60 * time.Second)
	}

}

func Search(allOf []rules.AllOf, oneOf []rules.OneOf, seconds int) []string {
	var elements []string
	cacheStorageMutex.RLock()
	cToBreak := 0
	ait := time.Now().UTC().Unix() - func()int64{
		switch seconds{
		case 0:
			return 60
		default:
			return int64(seconds)
		}
	}()
	for i := len(CacheStorage) - 1; i >= 0; i-- {
		est := gjson.Get(CacheStorage[i], "@timestamp").String()
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
				oneCatch = evalElement(CacheStorage[i], of.Field, of.Operator, of.Value)
				if oneCatch {
					break
				}
			}
			for _, af := range allOf {
				allCatch = evalElement(CacheStorage[i], af.Field, af.Operator, af.Value)
				if !allCatch {
					break
				}
			}
			if (len(allOf) == 0 || allCatch) && (len(oneOf) == 0 || oneCatch) {
				elements = append(elements, CacheStorage[i])
			}
		}
	}
	cacheStorageMutex.RUnlock()
	return elements
}

var logs = make(chan string, 10000)

func AddToCache(l string) {
	if len(l) == 10000 {
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
				cacheStorageMutex.Lock()
				CacheStorage = append(CacheStorage, l)
				cacheStorageMutex.Unlock()
			}
		}()
	}
}

func Clean() {
	for {
		if utils.AssignedMemory >= 80 && len(CacheStorage) > 500 {
			cacheStorageMutex.Lock()
			CacheStorage = CacheStorage[500:]
			cacheStorageMutex.Unlock()
		} else {
			time.Sleep(5 * time.Second)
		}
	}
}
