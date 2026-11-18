package ransomware

import (
	"context"
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/fanotify"
	"sync"
)

// FanotifyFeed reports actual attributed writes. Rename/delete/attribute events
// require a separate file-handle/name group and are not synthesized here.
type FanotifyFeed struct {
	monitor          *fanotify.Monitor
	mu               sync.RWMutex
	sink             func(FileEvent)
	interest         func(string) bool
	identityInterest func(uint64, uint64) bool
}

func NewFanotifyFeed(cfg fanotify.Config) (*FanotifyFeed, error) {
	if cfg.Mode != fanotify.ModeNotify {
		return nil, errors.New("ransomware activity feed requires fanotify notify mode")
	}
	f := &FanotifyFeed{}
	m, err := fanotify.New(cfg, nil, f.observe)
	if err != nil {
		return nil, err
	}
	f.monitor = m
	return f, nil
}
func (f *FanotifyFeed) setInterest(interest func(string) bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.interest = interest
}
func (f *FanotifyFeed) setIdentityInterest(interest func(uint64, uint64) bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.identityInterest = interest
}
func (f *FanotifyFeed) observe(o fanotify.Outcome) {
	e := o.Event
	if e.Mask&fanotify.Modify == 0 || e.PID <= maxKernelPID {
		return
	}
	f.mu.RLock()
	sink, interest, identityInterest := f.sink, f.interest, f.identityInterest
	f.mu.RUnlock()
	if sink == nil || (interest != nil || identityInterest != nil) &&
		(interest == nil || !interest(e.Path)) &&
		(identityInterest == nil || !identityInterest(e.Device, e.Inode)) {
		return
	}
	sink(FileEvent{PID: e.PID, Path: e.Path, Op: OpWrite,
		Device: e.Device, Inode: e.Inode, FileIdentityRequired: true,
		OSIdentity: e.OSIdentity, IdentityRequired: true, PPID: e.PPID, Image: e.Image, Cmdline: e.Cmdline})
}
func (f *FanotifyFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	f.mu.Lock()
	f.sink = sink
	f.mu.Unlock()
	return f.monitor.Run(ctx)
}
func (f *FanotifyFeed) Health() error { return f.monitor.Health() }
