package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestRegistryPrometheusExportsHTTPMetrics(t *testing.T) {
	registry := New()
	registry.RecordHTTP("GET", "/health/ready", 200, 150*time.Millisecond)
	registry.RecordHTTP("GET", "/health/ready", 200, 50*time.Millisecond)

	out := registry.Prometheus()

	assertContains(t, out, `goapp_http_requests_total{method="GET",route="/health/ready",status="200"} 2`)
	assertContains(t, out, `goapp_http_request_duration_seconds_sum{method="GET",route="/health/ready",status="200"} 0.200000`)
	assertContains(t, out, `goapp_http_request_duration_seconds_count{method="GET",route="/health/ready",status="200"} 2`)
}

func TestRegistryEscapesLabelValues(t *testing.T) {
	registry := New()
	registry.RecordHTTP("GET", `/bad"path`, 404, time.Millisecond)

	assertContains(t, registry.Prometheus(), `route="/bad\"path"`)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("output does not contain %q:\n%s", want, got)
	}
}
