package registry

import (
	"sync"
	"time"
)

type Dataset struct {
	ID         string
	Separator  rune
	Headers    []string
	ColIndex   map[string]int
	RawRows    [][]string
	Indices    map[string]map[string][]int
	indicesMu  sync.RWMutex
	UploadedAt time.Time
	SizeBytes  int64
	RowCount   int
}
