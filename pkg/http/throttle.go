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
	t.refreshAt = time.Now().Add(opts.RefreshPeriod)
	slog.Debug("created new throttling pool",
		slog.String("fqdn", t.fqdn),
		slog.Duration("refresh_duration", opts.RefreshPeriod),
		slog.Duration("min_delay", opts.MinDelay),
	)

	go func() {
		interval := time.NewTicker(opts.RefreshPeriod)
		defer interval.Stop()

		for {
			select {
			case <-interval.C:
				slog.Debug("refreshing requests count",
					slog.String("fqdn", t.fqdn),
					slog.Uint64("refreshing_counter", t.count.Load()),
				)

				t.mu.Lock()

				t.refreshAt = time.Now().Add(t.opts.RefreshPeriod)
				t.count.Swap(uint64(0))

				t.c.Broadcast()
				t.mu.Unlock()
			}
		}
	}()

	return &t
}

func (t *fqdnThrottle) isRequestAllowed() bool {
	return t.count.Load() < t.opts.MaxRequests
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
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !t.isRequestAllowed() {
			slog.WarnContext(ctx, "throttling request, waiting for refresh...",
				slog.String("fqdn", t.fqdn),
				slog.Time("refresh_at", t.refreshAt),
			)

			t.c.L.Lock()
			t.c.Wait()
		} else {
			t.count.Add(1)
			return nil
		}
	}
}
