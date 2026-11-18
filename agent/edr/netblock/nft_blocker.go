package netblock

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"time"
)

// nftBlocker owns desired state while the controller retains only verified
// kernel state. Failed updates never publish successful enforcer ownership.
type nftBlocker struct {
	mu        sync.Mutex
	c         *nftController
	safety    SafetyPolicy
	safetyErr error
	ips       map[netip.Addr]string
	prefixes  map[netip.Prefix]string
	unlock    func()
	closed    bool
}

func newNFTBlocker(c *nftController, unlock func()) *nftBlocker {
	return &nftBlocker{c: c, ips: map[netip.Addr]string{}, prefixes: map[netip.Prefix]string{}, safetyErr: ErrNetworkSafety, unlock: unlock}
}

func (b *nftBlocker) SetSafetyPolicy(p SafetyPolicy) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return errors.New("nftables backend is closed")
	}
	_, err := blockPlan(b.c.owner, p, nil, nil, b.c.now())
	if err != nil {
		b.safetyErr = err
		// If the new control plane cannot be resolved, fail open rather than
		// maintaining deny rules without a verified management exception.
		return errors.Join(err, b.c.apply(context.Background(), nil))
	}
	p.Management = append([]netip.Addr(nil), p.Management...)
	p.Resolvers = append([]netip.Addr(nil), p.Resolvers...)
	p.NeverBlock = append([]netip.Prefix(nil), p.NeverBlock...)
	b.safety = p
	b.safetyErr = nil
	// Also verify/remove a previous process's leased owned table before claiming
	// an initially empty policy is healthy. Its old exceptions may differ from
	// the newly resolved control plane after a service restart.
	return b.commitLocked(b.ips, b.prefixes)
}

func validDirection(dir string) bool { return dir == "in" || dir == "out" || dir == "both" }

// ApplyDelta reconciles a complete feed in one transaction, instead of launching
// one subprocess for every IOC. The same map also holds reactive exact IPs.
func (b *nftBlocker) ApplyDelta(add, remove []netip.Addr, addP, removeP []netip.Prefix, dir string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return errors.New("nftables backend is closed")
	}
	if b.safetyErr != nil {
		return b.safetyErr
	}
	if !validDirection(dir) {
		return errors.New("invalid network direction")
	}
	if len(add)+len(remove)+len(addP)+len(removeP) > 2*nftMaxEntries {
		return ErrNetworkBudget
	}
	ipCapacity := len(b.ips) + len(add)
	if ipCapacity > nftMaxEntries {
		ipCapacity = nftMaxEntries
	}
	ips := make(map[netip.Addr]string, ipCapacity)
	for a, d := range b.ips {
		ips[a] = d
	}
	prefixCapacity := len(b.prefixes) + len(addP)
	if prefixCapacity > nftMaxEntries {
		prefixCapacity = nftMaxEntries
	}
	prefixes := make(map[netip.Prefix]string, prefixCapacity)
	for p, d := range b.prefixes {
		prefixes[p] = d
	}
	for _, a := range remove {
		a, err := normalizedAddr(a)
		if err != nil {
			return err
		}
		delete(ips, a)
	}
	for _, p := range removeP {
		p, err := normalizedPrefix(p)
		if err != nil {
			return err
		}
		delete(prefixes, p)
	}
	for _, a := range add {
		a, err := normalizedAddr(a)
		if err != nil {
			return err
		}
		if _, exists := ips[a]; !exists && len(ips)+len(prefixes) >= nftMaxEntries {
			return ErrNetworkBudget
		}
		ips[a] = dir
	}
	for _, p := range addP {
		p, err := normalizedPrefix(p)
		if err != nil {
			return err
		}
		if _, exists := prefixes[p]; !exists && len(ips)+len(prefixes) >= nftMaxEntries {
			return ErrNetworkBudget
		}
		prefixes[p] = dir
	}
	if err := b.commitLocked(ips, prefixes); err != nil {
		return err
	}
	b.ips = ips
	b.prefixes = prefixes
	return nil
}

func (b *nftBlocker) commitLocked(ips map[netip.Addr]string, prefixes map[netip.Prefix]string) error {
	if len(ips)+len(prefixes) == 0 {
		return b.c.apply(context.Background(), nil)
	}
	p, err := blockPlan(b.c.owner, b.safety, ips, prefixes, b.c.now())
	if err != nil {
		return err
	}
	return b.c.apply(context.Background(), p)
}

func (b *nftBlocker) AddIP(a netip.Addr, dir string) error {
	return b.ApplyDelta([]netip.Addr{a}, nil, nil, nil, dir)
}
func (b *nftBlocker) RemoveIP(a netip.Addr) error {
	return b.ApplyDelta(nil, []netip.Addr{a}, nil, nil, "both")
}
func (b *nftBlocker) AddPrefix(p netip.Prefix, dir string) error {
	return b.ApplyDelta(nil, nil, []netip.Prefix{p}, nil, dir)
}
func (b *nftBlocker) RemovePrefix(p netip.Prefix) error {
	return b.ApplyDelta(nil, nil, nil, []netip.Prefix{p}, "both")
}

func (b *nftBlocker) Reset() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return errors.New("nftables backend is closed")
	}
	if err := b.c.apply(context.Background(), nil); err != nil {
		return err
	}
	b.ips = map[netip.Addr]string{}
	b.prefixes = map[netip.Prefix]string{}
	return nil
}

func (b *nftBlocker) Count() int { b.mu.Lock(); defer b.mu.Unlock(); return b.c.count() }
func (b *nftBlocker) Health() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return errors.Join(b.safetyErr, b.c.health())
}

// Run refreshes the finite blocklist lease. Isolation has no refresher: a CLI
// exit or stalled service cannot extend its explicit deadline.
func (b *nftBlocker) Run(ctx context.Context) {
	t := time.NewTicker(nftRefresh)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			b.mu.Lock()
			if !b.closed && b.safetyErr == nil && len(b.ips)+len(b.prefixes) > 0 {
				_ = b.commitLocked(b.ips, b.prefixes)
			}
			b.mu.Unlock()
		}
	}
}

func (b *nftBlocker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return b.c.health()
	}
	err := b.c.apply(context.Background(), nil)
	b.closed = true
	if b.unlock != nil {
		b.unlock()
		b.unlock = nil
	}
	return err
}
