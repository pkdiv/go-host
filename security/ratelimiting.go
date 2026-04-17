package security

import (
	"sync"
	"time"
)

type ClientLimiter struct {
	mu           sync.Mutex
	counts       map[string]int
	lastRequests map[string]time.Time
	window       time.Duration
	limit        int
}

func (c *ClientLimiter) Allow(clientIP string) bool {

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	if c.lastRequests[clientIP].Add(c.window).Before(now) {
		c.lastRequests[clientIP] = now
		c.counts[clientIP] = 0
	}

	c.counts[clientIP]++
	c.lastRequests[clientIP] = now

	return c.counts[clientIP] <= c.limit

}

func NewClientLimiter(window time.Duration, limit int) *ClientLimiter {
	return &ClientLimiter{
		counts:       make(map[string]int),
		lastRequests: make(map[string]time.Time),
		window:       window,
		limit:        limit,
	}
}
