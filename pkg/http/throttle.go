package http

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type ThrottleMap struct {
	mu sync.Mutex
	m  map[string]*fqdnThrottle

	opts *ThrottleOptions
}

func NewThrottleMap(opts *ThrottleOptions) *ThrottleMap {
	return &ThrottleMap{
		m:    make(map[string]*fqdnThrottle),
		mu:   sync.Mutex{},
		opts: opts,
	}
}

type ThrottleOptions struct {
	RefreshPeriod time.Duration
	MaxRequests   uint64

	MinDelay time.Duration
}

func (tm *ThrottleMap) Wait(ctx context.Context, fqdn string) error {
	if fqdn == "" {
		return errors.New("empty fqdn provided")
	}

	tm.mu.Lock()
	w, ok := tm.m[fqdn]
	if !ok {
		w = newThrottle(fqdn, tm.opts)
		tm.m[fqdn] = w
	}

	tm.mu.Unlock()
	return w.wait(ctx)
}

type fqdnThrottle struct {
	fqdn string

	mu sync.Mutex
	c  sync.Cond

	refreshAt time.Time
	count     atomic.Uint64

	opts *ThrottleOptions
}

func newThrottle(fqdn string, opts *ThrottleOptions) *fqdnThrottle {
	if opts == nil {
		panic("nil throttle options")
	}

	if opts.MaxRequests == 0 {
		panic("throttle options: max requests cannot be 0")
	}

	if opts.MinDelay < 0 {
		panic("throttle min delay cannot be less then 0")
	}

	t := fqdnThrottle{
		fqdn:  fqdn,
		mu:    sync.Mutex{},
		count: atomic.Uint64{},
		opts:  opts,
	}

	t.c = sync.Cond{L: &t.mu}
	t.refreshAt = time.Now().Truncate(time.Second * 60).Add(opts.RefreshPeriod)
	slog.Debug("created new throttling pool",
		slog.String("fqdn", t.fqdn),
		slog.Duration("refresh_duration", opts.RefreshPeriod),
		slog.Duration("min_delay", opts.MinDelay),
	)

	return &t
}

func (t *fqdnThrottle) isRequestAllowed() bool {
	var now = time.Now()
	if now.After(t.refreshAt) {
		slog.Debug("refreshing requests count",
			slog.String("fqdn", t.fqdn),
			slog.Uint64("refreshing_counter", t.count.Load()),
		)

		t.mu.Lock()

		t.refreshAt = now.Truncate(time.Second * 60).Add(t.opts.RefreshPeriod)
		t.count.Swap(uint64(0))

		t.mu.Unlock()
	}

	if t.count.Load() < t.opts.MaxRequests {
		return true
	}

	return false
}

func (t *fqdnThrottle) wait(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			t.mu.Lock()
			t.c.Broadcast() // broadcast wake up all waiters if context canceled for anyone
			t.mu.Unlock()
		case <-done:
		}
	}()
	defer close(done)

	for {
		if t.isRequestAllowed() {
			t.count.Add(1)
			time.Sleep(t.opts.MinDelay)

			return nil
		}

		// if request rejected wait for next window or context cancel
		now := time.Now()
		waitFor := t.refreshAt.Sub(now)

		slog.WarnContext(ctx, "throttling request...",
			slog.String("fqdn", t.fqdn),
			slog.Duration("wait_for", waitFor),
		)

		timer := time.NewTimer(waitFor)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			t.mu.Lock()
			t.c.Broadcast()
			t.mu.Unlock()
		}
	}
}
