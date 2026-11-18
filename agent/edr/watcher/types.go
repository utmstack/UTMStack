package watcher

import "errors"

var ErrUnavailable = errors.New("file monitoring is unavailable")
var ErrWatchLimit = errors.New("file monitoring scope exceeds its resource limit")
var ErrResync = errors.New("file monitoring requires reconciliation")

type Sink interface{ Enqueue(path, op string) }
type CursorStore interface {
	LoadUSN(volume string) (uint64, int64)
	SaveUSN(volume string, journalID uint64, nextUSN int64)
}
