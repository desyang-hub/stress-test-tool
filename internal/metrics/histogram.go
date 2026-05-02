package metrics

import (
	"fmt"
	"strconv"
)

// DefaultBuckets defines the standard latency histogram bucket thresholds in milliseconds.
var DefaultBuckets = []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// Histogram represents a latency histogram with bucket counts.
type Histogram struct {
	Buckets  []float64   // bucket boundaries in ms
	Labels   []string    // display labels
	Counts   []int64     // count per bucket
	Boundary float64     // last boundary (for overflow)
	Overflow int64       // count beyond last boundary
}

// BuildHistogram creates a histogram from a []float64 latency slice (in ms).
// Values are placed into consecutive buckets; values beyond the last boundary go to Overflow.
func BuildHistogram(latencies []float64, buckets []float64) *Histogram {
	if len(buckets) == 0 {
		buckets = DefaultBuckets
	}

	labels := make([]string, len(buckets))
	for i, b := range buckets {
		labels[i] = formatBucketLabel(b)
	}

	h := &Histogram{
		Buckets:  buckets,
		Labels:   labels,
		Counts:   make([]int64, len(buckets)),
		Boundary: buckets[len(buckets)-1],
	}

	for _, ms := range latencies {
		placed := false
		for i, b := range buckets {
			if ms <= b {
				h.Counts[i]++
				placed = true
				break
			}
		}
		if !placed {
			h.Overflow++
		}
	}
	return h
}

func formatBucketLabel(ms float64) string {
	if ms < 1000 {
		return formatMs(ms) + "ms"
	}
	s := ms / 1000
	return formatSec(s) + "s"
}

func formatMs(ms float64) string {
	if ms == float64(int(ms)) {
		return formatInt(float64(int(ms)))
	}
	return formatFloat(ms, 1)
}

func formatSec(s float64) string {
	if s == float64(int(s)) {
		return formatInt(s)
	}
	return formatFloat(s, 2)
}

func formatInt(v float64) string {
	return fmt.Sprintf("%d", int(v))
}

func formatFloat(v float64, precision int) string {
	return fmt.Sprintf("%."+strconv.Itoa(precision)+"f", v)
}
