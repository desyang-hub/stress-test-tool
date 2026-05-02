package reporters

import (
	"bytes"
	"strings"
	"unicode/utf8"
	"testing"
	"time"
)

func TestPrintSummary(t *testing.T) {
	s := Stats{
		Name:            "test-load",
		TotalRequests:   1000,
		Successful:      950,
		Failed:          50,
		TPS:             100.5,
		TotalDuration:   10 * time.Second,
		LatencyMin:      1.0,
		LatencyMean:     25.5,
		LatencyMedian:   20.0,
		LatencyP90:      50.0,
		LatencyP95:      75.0,
		LatencyP99:      120.0,
		LatencyMax:      500.0,
		LatencyStdDev:   30.0,
		StatusCodeBreakdown: map[int]int64{200: 900, 500: 50},
		ErrorBreakdown:    map[string]int64{"timeout": 30, "connection": 20},
		Histogram: map[string]int64{
			"0-10ms":  200,
			"10-50ms": 500,
			"50-100ms": 200,
			">100ms":  100,
		},
	}

	var buf bytes.Buffer
	r := &ConsoleReporter{w: &buf}
	r.PrintSummary(s)

	output := buf.String()
	if !strings.Contains(output, "STRESS TEST REPORT") {
		t.Error("missing header")
	}
	if !strings.Contains(output, "test-load") {
		t.Error("missing name")
	}
	if !strings.Contains(output, "Total Requests:") {
		t.Error("missing total requests")
	}
	if !strings.Contains(output, "100.5") {
		t.Error("missing TPS")
	}
	if !strings.Contains(output, "Latency (ms):") {
		t.Error("missing latency section")
	}
	if !strings.Contains(output, "Status Codes:") {
		t.Error("missing status codes")
	}
	if !strings.Contains(output, "Errors:") {
		t.Error("missing errors")
	}
	if !strings.Contains(output, "Latency Histogram:") {
		t.Error("missing histogram")
	}
	if !strings.Contains(output, "█") {
		t.Error("missing histogram bars")
	}
}

func TestPrintSummaryEmpty(t *testing.T) {
	var buf bytes.Buffer
	r := &ConsoleReporter{w: &buf}
	r.PrintSummary(Stats{})

	output := buf.String()
	if !strings.Contains(output, "STRESS TEST REPORT") {
		t.Error("missing header for empty stats")
	}
}

func TestSuccessRate(t *testing.T) {
	if got := successRate(Stats{TotalRequests: 100, Successful: 80}); got != 80.0 {
		t.Errorf("successRate = %f, want 80.0", got)
	}
	if got := successRate(Stats{}); got != 0 {
		t.Errorf("successRate on empty = %f, want 0", got)
	}
}

func TestFailRate(t *testing.T) {
	if got := failRate(Stats{TotalRequests: 100, Failed: 20}); got != 20.0 {
		t.Errorf("failRate = %f, want 20.0", got)
	}
}

func TestHistogramBar(t *testing.T) {
	bar := histogramBar(4, 500, 1000)
	if utf8.RuneCountInString(bar) == 0 {
		t.Error("expected non-empty bar")
	}
	if utf8.RuneCountInString(bar) > 40 {
		t.Error("bar exceeds max length")
	}

	// 50% of 40 = 20 runes
	if utf8.RuneCountInString(bar) != 20 {
		t.Errorf("bar runes = %d, want 20", utf8.RuneCountInString(bar))
	}

	// Zero count should give empty bar
	bar = histogramBar(4, 0, 1000)
	if bar != "" {
		t.Errorf("expected empty bar for zero count, got %q", bar)
	}

	// Zero total should give empty bar
	bar = histogramBar(4, 100, 0)
	if bar != "" {
		t.Errorf("expected empty bar for zero total, got %q", bar)
	}
}
