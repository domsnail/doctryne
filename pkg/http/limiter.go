package http

import (
	"sync"
	"time"
)

type HostLimiter struct {
	mu sync.Mutex

	lastSeen map[string]time.Time

	delay time.Duration
}

func NewHostLimiter(delay time.Duration) *HostLimiter {
	return &HostLimiter{
		lastSeen: make(map[string]time.Time),
		delay:    delay,
	}
}
