// Package reporters provides output formatters for stress test results.
package reporters

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ConsoleReporter prints live progress and final summary to stdout.
type ConsoleReporter struct {
	w io.Writer
}

// NewConsoleReporter creates a console reporter.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{w: os.Stdout}
}

// PrintSummary renders the final stats summary.
func (r *ConsoleReporter) PrintSummary(s Stats) {
	w := r.w

	// Header
	fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 70))
	fmt.Fprintf(w, "  STRESS TEST REPORT")
	if s.Name != "" {
		fmt.Fprintf(w, " - %s", s.Name)
	}
	fmt.Fprintf(w, "\n%s\n", strings.Repeat("=", 70))

	// Key metrics
	fmt.Fprintf(w, "\n  Total Requests:    %d\n", s.TotalRequests)
	fmt.Fprintf(w, "  Successful:        %d (%.1f%%)\n", s.Successful, successRate(s))
	fmt.Fprintf(w, "  Failed:            %d (%.1f%%)\n", s.Failed, failRate(s))
	fmt.Fprintf(w, "  Throughput:        %.1f req/s\n", s.TPS)
	fmt.Fprintf(w, "  Total Duration:    %s\n\n", s.TotalDuration.Round(time.Millisecond))

	// Latency
	fmt.Fprintf(w, "  Latency (ms):\n")
	fmt.Fprintf(w, "    Min:     %.1f\n", s.LatencyMin)
	fmt.Fprintf(w, "    Mean:    %.1f\n", s.LatencyMean)
	fmt.Fprintf(w, "    P50:     %.1f\n", s.LatencyMedian)
	fmt.Fprintf(w, "    P90:     %.1f\n", s.LatencyP90)
	fmt.Fprintf(w, "    P95:     %.1f\n", s.LatencyP95)
	fmt.Fprintf(w, "    P99:     %.1f\n", s.LatencyP99)
	fmt.Fprintf(w, "    Max:     %.1f\n", s.LatencyMax)
	fmt.Fprintf(w, "    StdDev:  %.1f\n\n", s.LatencyStdDev)

	// Status code breakdown
	if len(s.StatusCodeBreakdown) > 0 {
		fmt.Fprintf(w, "  Status Codes:\n")
		for code := range s.StatusCodeBreakdown {
			count := s.StatusCodeBreakdown[code]
			pct := float64(count) / float64(s.TotalRequests) * 100
			fmt.Fprintf(w, "    %d: %d (%.1f%%)\n", code, count, pct)
		}
		fmt.Fprintln(w)
	}

	// Error breakdown
	if len(s.ErrorBreakdown) > 0 {
		fmt.Fprintf(w, "  Errors:\n")
		for errType, count := range s.ErrorBreakdown {
			pct := float64(count) / float64(s.Failed) * 100
			fmt.Fprintf(w, "    %s: %d (%.1f%%)\n", errType, count, pct)
		}
		fmt.Fprintln(w)
	}

	// Histogram
	if len(s.Histogram) > 0 {
		fmt.Fprintf(w, "  Latency Histogram:\n")
		for label, count := range s.Histogram {
			b := histogramBar(int64(len(s.Histogram)), count, s.TotalRequests)
			fmt.Fprintf(w, "    %-8s | %s %d\n", label, b, count)
		}
		fmt.Fprintln(w)
	}

	// Footer
	fmt.Fprintf(w, "%s\n", strings.Repeat("=", 70))
}

func successRate(s Stats) float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.Successful) / float64(s.TotalRequests) * 100
}

func failRate(s Stats) float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.Failed) / float64(s.TotalRequests) * 100
}

func histogramBar(_ int64, count, total int64) string {
	if total == 0 {
		return ""
	}
	// Scale to max 40 chars
	const maxLen = 40
	pct := float64(count) / float64(total)
	barLen := int(pct * float64(maxLen))
	if barLen > maxLen {
		barLen = maxLen
	}
	if barLen < 1 && count > 0 {
		barLen = 1
	}
	return strings.Repeat("█", barLen)
}

// Stats holds the data needed for report rendering.
type Stats struct {
	Name                string
	TotalRequests       int64
	Successful          int64
	Failed              int64
	TPS                 float64
	TotalDuration       time.Duration
	LatencyMin          float64
	LatencyMean         float64
	LatencyMedian       float64
	LatencyP90          float64
	LatencyP95          float64
	LatencyP99          float64
	LatencyMax          float64
	LatencyStdDev       float64
	StatusCodeBreakdown map[int]int64
	ErrorBreakdown      map[string]int64
	Histogram           map[string]int64
}
