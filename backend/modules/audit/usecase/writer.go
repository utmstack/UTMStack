package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	queueSize  = 4096
	batchMax   = 200
	flushEvery = time.Second

	writeTimeout = 10 * time.Second
)

type writer struct {
	repo  connectors.Repository
	queue chan *domain.AuditLog

	stop sync.Once
	done chan struct{}
}

func newWriter(repo connectors.Repository) *writer {
	return &writer{
		repo:  repo,
		queue: make(chan *domain.AuditLog, queueSize),
		done:  make(chan struct{}),
	}
}

func (w *writer) enqueue(row *domain.AuditLog) {
	select {
	case w.queue <- row:
	default:
		w.insert([]*domain.AuditLog{row})
	}
}

func (w *writer) Start() {
	go w.run()
}

func (w *writer) run() {
	defer close(w.done)

	batch := make([]*domain.AuditLog, 0, batchMax)
	ticker := time.NewTicker(flushEvery)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		w.insert(batch)
		batch = batch[:0]
	}

	for {
		select {
		case row, ok := <-w.queue:
			if !ok {
				flush()
				return
			}
			batch = append(batch, row)
			if len(batch) >= batchMax {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (w *writer) Stop() {
	w.stop.Do(func() { close(w.queue) })
	select {
	case <-w.done:
	case <-time.After(writeTimeout):
		_ = catcher.Error("audit queue did not drain before shutdown", nil, nil)
	}
}

func (w *writer) insert(rows []*domain.AuditLog) {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	if err := w.repo.InsertBatch(tenancy.WithAllTenants(ctx), rows); err != nil {
		_ = catcher.Error("audit insert failed", err, map[string]any{"rows": len(rows)})
	}
}
