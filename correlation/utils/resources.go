package utils

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/threatwinds/go-sdk/catcher"
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
		catcher.Info("Memory used by engine", map[string]any{"used_by_engine": usedByEngine})
		catcher.Info("Free memory", map[string]any{"free_memory": FreeMemory()})
		catcher.Info("Physical memory", map[string]any{"physical_memory": TotalMemory()})
		AssignedMemory = float32(usedByEngine) / float32(TotalMemory()/4) * 100
		catcher.Info("Assigned memory used", map[string]any{"assigned_memory_used": AssignedMemory})
		time.Sleep(60 * time.Second)
	}
}
