package http

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type HostLimiter struct {
	mu sync.RWMutex

	lastSeen map[string]time.Time
	delay    time.Duration
}

func NewHostLimiter(delay time.Duration) *HostLimiter {
	if delay < 0 {
		panic("host limiter delay must greater or equal then zero")
	}

	return &HostLimiter{
		lastSeen: make(map[string]time.Time),
		delay:    delay,
	}
}

func (l *HostLimiter) Wait(ctx context.Context, host string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// todo: add progressive wait

	now := time.Now()
	nextAllowed := l.lastSeen[host].Add(l.delay)
	wait := time.Until(nextAllowed)
	if wait <= 0 {
		l.lastSeen[host] = now
		return nil
	}

	slog.Log(ctx, -8, "rate limiting request", slog.String("host", host))
	timer := time.NewTimer(wait)
	defer timer.Stop()

	l.mu.Unlock()
	select {
	case <-ctx.Done():
		l.mu.Lock()
		return ctx.Err()
	case <-timer.C:
		l.mu.Lock()
		l.lastSeen[host] = time.Now()
		return nil
	}
}
