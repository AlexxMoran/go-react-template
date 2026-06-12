// Package metrics provides a tiny Prometheus-compatible metrics hook.
//
// The template keeps this deliberately small: it exposes HTTP request counters
// and duration totals when METRICS_ENABLED=true. Production services can replace
// or extend this package with a full observability stack without touching
// handlers or domain code.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Registry struct {
	mu       sync.Mutex
	requests map[requestKey]requestStats
}

type requestKey struct {
	Method string
	Route  string
	Status int
}

type requestStats struct {
	Count       uint64
	DurationSum float64
}

func New() *Registry {
	return &Registry{requests: make(map[requestKey]requestStats)}
}

func (r *Registry) RecordHTTP(method, route string, status int, duration time.Duration) {
	if route == "" {
		route = "unknown"
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := requestKey{Method: method, Route: route, Status: status}
	stats := r.requests[key]
	stats.Count++
	stats.DurationSum += duration.Seconds()
	r.requests[key] = stats
}

func Handler(r *Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(r.Prometheus()))
	}
}

func Middleware(r *Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		r.RecordHTTP(c.Request.Method, route, c.Writer.Status(), time.Since(started))
	}
}

func (r *Registry) Prometheus() string {
	r.mu.Lock()
	rows := make([]struct {
		Key   requestKey
		Stats requestStats
	}, 0, len(r.requests))
	for key, stats := range r.requests {
		rows = append(rows, struct {
			Key   requestKey
			Stats requestStats
		}{Key: key, Stats: stats})
	}
	r.mu.Unlock()

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Key.Route != rows[j].Key.Route {
			return rows[i].Key.Route < rows[j].Key.Route
		}
		if rows[i].Key.Method != rows[j].Key.Method {
			return rows[i].Key.Method < rows[j].Key.Method
		}
		return rows[i].Key.Status < rows[j].Key.Status
	})

	var b strings.Builder
	b.WriteString("# HELP goapp_http_requests_total Total HTTP requests handled by route.\n")
	b.WriteString("# TYPE goapp_http_requests_total counter\n")
	for _, row := range rows {
		fmt.Fprintf(&b, "goapp_http_requests_total{%s} %d\n", labels(row.Key), row.Stats.Count)
	}

	b.WriteString("# HELP goapp_http_request_duration_seconds_sum Total HTTP request duration by route.\n")
	b.WriteString("# TYPE goapp_http_request_duration_seconds_sum counter\n")
	for _, row := range rows {
		fmt.Fprintf(&b, "goapp_http_request_duration_seconds_sum{%s} %.6f\n", labels(row.Key), row.Stats.DurationSum)
	}

	b.WriteString("# HELP goapp_http_request_duration_seconds_count Total HTTP request duration observations by route.\n")
	b.WriteString("# TYPE goapp_http_request_duration_seconds_count counter\n")
	for _, row := range rows {
		fmt.Fprintf(&b, "goapp_http_request_duration_seconds_count{%s} %d\n", labels(row.Key), row.Stats.Count)
	}

	return b.String()
}

func labels(key requestKey) string {
	return fmt.Sprintf(`method="%s",route="%s",status="%d"`,
		escapeLabel(key.Method),
		escapeLabel(key.Route),
		key.Status,
	)
}

func escapeLabel(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return strings.ReplaceAll(s, `"`, `\"`)
}
