package ratelimit

import (
	"sync"
	"time"
)

type record struct {
	count        int
	firstFailure time.Time
	blockedUntil time.Time
}

type LoginLimiter struct {
	mu          sync.Mutex
	attempts    map[string]*record
	maxFailures int
	blockTTL    time.Duration
	windowTTL   time.Duration
}

func NewLoginLimiter(maxFailures int, blockTTL, windowTTL time.Duration) *LoginLimiter {
	return &LoginLimiter{
		attempts:    make(map[string]*record),
		maxFailures: maxFailures,
		blockTTL:    blockTTL,
		windowTTL:   windowTTL,
	}
}

func (l *LoginLimiter) IsBlocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.attempts[key]
	if !ok {
		return false
	}
	if !r.blockedUntil.IsZero() && time.Now().Before(r.blockedUntil) {
		return true
	}
	if !r.blockedUntil.IsZero() {
		delete(l.attempts, key)
	}
	return false
}

func (l *LoginLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	r, ok := l.attempts[key]
	if !ok || now.Sub(r.firstFailure) > l.windowTTL {
		l.attempts[key] = &record{count: 1, firstFailure: now}
		return
	}
	r.count++
	if r.count >= l.maxFailures {
		r.blockedUntil = now.Add(l.blockTTL)
	}
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}
