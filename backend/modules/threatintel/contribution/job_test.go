package contribution

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
)

type countingSource struct {
	reads atomic.Int32
	panic bool
	err   error
}

func (c *countingSource) RecentIncidents(context.Context) ([]*models.Incident, error) {
	c.reads.Add(1)

	if c.panic {
		panic("boom")
	}

	return nil, c.err
}

func (c *countingSource) IncidentAlerts(context.Context, string) ([]*models.IncidentAlert, error) {
	return nil, nil
}

type fakeLease struct {
	mine bool
	err  error
	name string
}

func (f *fakeLease) Acquire(_ context.Context, name string, _ time.Duration) (bool, error) {
	f.name = name

	return f.mine, f.err
}

func TestADisabledJobReadsNothing(t *testing.T) {
	src := &countingSource{}
	NewJob(src, nil, func() bool { return false }, &fakeLease{mine: true}).runOnce(context.Background())

	if src.reads.Load() != 0 {
		t.Error("a disabled contribution must not touch the incidents")
	}
}

func TestAnEnabledJobThatHoldsTheLeaseRuns(t *testing.T) {
	src, lease := &countingSource{}, &fakeLease{mine: true}
	NewJob(src, nil, func() bool { return true }, lease).runOnce(context.Background())

	if src.reads.Load() != 1 || lease.name != "threatintel-contribution" {
		t.Errorf("reads=%d lease=%q", src.reads.Load(), lease.name)
	}
}

func TestAReplicaWithoutTheLeaseDoesNothing(t *testing.T) {
	for _, l := range []*fakeLease{{mine: false}, {err: errors.New("db down")}} {
		src := &countingSource{}
		NewJob(src, nil, func() bool { return true }, l).runOnce(context.Background())

		if src.reads.Load() != 0 {
			t.Errorf("lease %+v: ran without it", l)
		}
	}
}

func TestAJobWithoutLeasesStillRuns(t *testing.T) {
	src := &countingSource{}
	NewJob(src, nil, func() bool { return true }, nil).runOnce(context.Background())

	if src.reads.Load() != 1 {
		t.Error("expected a cycle")
	}
}

func TestAPanicInACycleDoesNotTakeTheBackendDown(t *testing.T) {
	src := &countingSource{panic: true}
	NewJob(src, nil, func() bool { return true }, nil).runOnce(context.Background())
}

func TestAFailedReadEndsTheCycleQuietly(t *testing.T) {
	src := &countingSource{err: errors.New("db down")}
	NewJob(src, nil, func() bool { return true }, nil).runOnce(context.Background())

	if src.reads.Load() != 1 {
		t.Error("expected one attempt")
	}
}

func TestStartRunsNowAndStopsWithTheContext(t *testing.T) {
	src := &countingSource{}
	j := NewJob(src, nil, func() bool { return true }, nil)
	j.interval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() { j.Start(ctx); close(done) }()

	time.Sleep(60 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start did not stop")
	}

	if src.reads.Load() < 2 {
		t.Errorf("expected repeated cycles, got %d", src.reads.Load())
	}
}

func TestANilJobStartIsHarmless(t *testing.T) {
	var j *Job
	j.Start(context.Background())
}
