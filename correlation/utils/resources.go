package utils

import (
	"log"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

var AssignedMemory float32

func TotalMemory() uint64 {
	v, _ := mem.VirtualMemory()
	return v.Total / 1024 / 1024
}

func FreeMemory() uint64 {
	v, _ := mem.VirtualMemory()
	return v.Free / 1024 / 1024
}

func UsedByEngineMemory() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024
}

func Status() {
	for {
		usedByEngine := UsedByEngineMemory()
		log.Printf("Memory used by engine: %v MB", usedByEngine)
		log.Printf("Free memory: %v MB", FreeMemory())
		log.Printf("Physical memory: %v MB", TotalMemory())
		AssignedMemory = float32(usedByEngine) / float32(TotalMemory()/4) * 100
		log.Printf("Assigned memory used: %v %%", AssignedMemory)
		time.Sleep(60 * time.Second)
	}
}
