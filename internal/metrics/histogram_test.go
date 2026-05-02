package metrics

import (
	"testing"
)

func TestBuildHistogram(t *testing.T) {
	latencies := []float64{5, 15, 35, 75, 150, 350, 1500, 3000, 15000}
	h := BuildHistogram(latencies, nil)

	// Buckets: 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000
	// 5 -> bucket 0 (<10)
	// 15 -> bucket 1 (<25)
	// 35 -> bucket 2 (<50)
	// 75 -> bucket 3 (<100)
	// 150 -> bucket 4 (<250)
	// 350 -> bucket 5 (<500)
	// 1500 -> bucket 6 (<1s/1000ms)
	// 3000 -> bucket 7 (<2.5s/2500ms)
	// 15000 -> overflow

	if h.Counts[0] != 1 {
		t.Errorf("bucket 0 (<10ms): count = %d, want 1", h.Counts[0])
	}
	if h.Counts[1] != 1 {
		t.Errorf("bucket 1 (<25ms): count = %d, want 1", h.Counts[1])
	}
	if h.Counts[5] != 1 {
		t.Errorf("bucket 5 (<500ms): count = %d, want 1", h.Counts[5])
	}
	if h.Overflow != 1 {
		t.Errorf("overflow: count = %d, want 1", h.Overflow)
	}
}

func TestBuildHistogramEmpty(t *testing.T) {
	h := BuildHistogram([]float64{}, nil)
	for _, c := range h.Counts {
		if c != 0 {
			t.Errorf("empty histogram: bucket count = %d, want 0", c)
		}
	}
}
