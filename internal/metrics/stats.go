// Package metrics provides statistics computation for stress test results.
package metrics

import "math"

// CalcPercentiles computes latency percentiles from a sorted []float64 slice (milliseconds).
// The slice must be sorted in ascending order before calling.
// Uses linear interpolation: rank = (P/100) * (N-1).
func CalcPercentiles(data []float64, percentiles ...float64) map[float64]float64 {
	if len(data) == 0 {
		return make(map[float64]float64)
	}

	result := make(map[float64]float64, len(percentiles))
	n := float64(len(data))

	for _, p := range percentiles {
		if p <= 0 {
			result[p] = data[0]
			continue
		}
		if p >= 100 {
			result[p] = data[len(data)-1]
			continue
		}

		// Linear interpolation: map [0,100] percentiles to [0, N-1] index
		rank := (p / 100.0) * (n - 1)
		lower := int(math.Floor(rank))
		upper := int(math.Ceil(rank))

		if lower == upper {
			result[p] = data[lower]
		} else {
			frac := rank - float64(lower)
			result[p] = data[lower]*(1-frac) + data[upper]*frac
		}
	}
	return result
}

// CalcStandardDeviation computes the population standard deviation.
func CalcStandardDeviation(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	var sumSq float64
	for _, v := range data {
		diff := v - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

// CalcHistogram creates bucket counts for a histogram.
// Buckets: <10ms, <25ms, <50ms, <100ms, <250ms, <500ms, <1s, <2.5s, <5s, <10s, >=10s.
func CalcHistogram(latencies []float64) map[string]int64 {
	type bucket struct {
		label    string
		threshold float64
	}

	// Ordered slices maintain insertion order for deterministic placement
	buckets := []bucket{
		{"<10ms", 10},
		{"<25ms", 25},
		{"<50ms", 50},
		{"<100ms", 100},
		{"<250ms", 250},
		{"<500ms", 500},
		{"<1s", 1000},
		{"<2.5s", 2500},
		{"<5s", 5000},
		{"<10s", 10000},
		{">=10s", math.MaxFloat64},
	}

	hist := make(map[string]int64)
	for _, b := range buckets {
		hist[b.label] = 0
	}

	for _, ms := range latencies {
		for _, b := range buckets {
			if ms < b.threshold {
				hist[b.label]++
				break
			}
		}
	}
	return hist
}
