package middleware

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/yourorg/goapp/internal/platform/httpx"
	"github.com/yourorg/goapp/pkg/apperror"
)

// RateLimit applies a per-client-IP token-bucket limit: each IP gets its own
// limiter allowing rps sustained requests/second with momentary bursts up to
// burst. Exceeding it returns 429 with the standard error envelope. Idle
// limiters are evicted by a background sweeper so memory stays bounded under many
// distinct client IPs.
//
// ClientIP is only trustworthy when the server is configured with the right
// trusted proxies (see SetTrustedProxies in the server wiring); otherwise an
// attacker can spoof X-Forwarded-For to dodge the limit.
func RateLimit(rps float64, burst int, logger *slog.Logger) gin.HandlerFunc {
	limiters := newIPLimiters(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		if !limiters.get(c.ClientIP()).Allow() {
			httpx.WriteError(c, logger, apperror.TooManyRequests(
				"rate_limited", "Too many requests, please slow down"))
			c.Abort()
			return
		}
		c.Next()
	}
}

type ipLimiters struct {
	mu       sync.Mutex
	limiters map[string]*ipEntry
	rps      rate.Limit
	burst    int
}

type ipEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newIPLimiters(rps rate.Limit, burst int) *ipLimiters {
	l := &ipLimiters{
		limiters: make(map[string]*ipEntry),
		rps:      rps,
		burst:    burst,
	}
	go l.sweep()
	return l
}

func (l *ipLimiters) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.limiters[ip]
	if !ok {
		e = &ipEntry{limiter: rate.NewLimiter(l.rps, l.burst)}
		l.limiters[ip] = e
	}
	e.lastSeen = time.Now()
	return e.limiter
}

// sweep evicts limiters idle for longer than the TTL. It runs for the lifetime of
// the process (the limiter is a startup singleton).
func (l *ipLimiters) sweep() {
	const ttl = 3 * time.Minute
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-ttl)
		l.mu.Lock()
		for ip, e := range l.limiters {
			if e.lastSeen.Before(cutoff) {
				delete(l.limiters, ip)
			}
		}
		l.mu.Unlock()
	}
}
