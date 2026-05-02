package metrics

import (
	"math"
	"testing"
)

func TestCalcPercentiles(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	p := CalcPercentiles(data, 50, 90, 95, 99)

	if got := p[50]; got != 5.5 {
		t.Errorf("P50 = %v, want 5.5", got)
	}
	if got := p[90]; got != 9.1 {
		t.Errorf("P90 = %v, want 9.1", got)
	}
	if got := p[95]; math.Abs(got-9.55) > 0.001 {
		t.Errorf("P95 = %v, want 9.55", got)
	}
	if got := p[99]; got != 9.91 {
		t.Errorf("P99 = %v, want 9.91", got)
	}
}

func TestCalcPercentilesEmpty(t *testing.T) {
	p := CalcPercentiles([]float64{}, 50)
	if len(p) != 0 {
		t.Errorf("expected empty result, got %v", p)
	}
}

func TestCalcPercentilesSingle(t *testing.T) {
	data := []float64{42.0}
	p := CalcPercentiles(data, 50, 90, 95, 99)
	for _, v := range p {
		if v != 42.0 {
			t.Errorf("expected 42.0 for all percentiles on single-element slice, got %v", v)
		}
	}
}

func TestCalcPercentilesEdgeCases(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5}
	p := CalcPercentiles(data, 0, 100)
	if p[0] != 1 {
		t.Errorf("P0 = %v, want 1", p[0])
	}
	if p[100] != 5 {
		t.Errorf("P100 = %v, want 5", p[100])
	}
}

func TestCalcStandardDeviation(t *testing.T) {
	data := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	mean := 5.0
	got := CalcStandardDeviation(data, mean)
	want := 2.0
	if got != want {
		t.Errorf("stddev = %v, want %v", got, want)
	}
}

func TestCalcStandardDeviationEmpty(t *testing.T) {
	got := CalcStandardDeviation([]float64{}, 0)
	if got != 0 {
		t.Errorf("expected 0 for empty slice, got %v", got)
	}
}

func TestCalcHistogram(t *testing.T) {
	data := []float64{5, 15, 30, 60, 150, 300, 600, 1500, 3000, 15000}
	h := CalcHistogram(data)

	expected := map[string]int64{
		"<10ms":  1,
		"<25ms":  1,
		"<50ms":  1,
		"<100ms": 1,
		"<250ms": 1,
		"<500ms": 1,
		"<1s":    1,
		"<2.5s":  1,
		"<5s":    1,
		"<10s":   0,
		">=10s":  1,
	}

	for label, want := range expected {
		if got := h[label]; got != want {
			t.Errorf("bucket %s: count = %d, want %d", label, got, want)
		}
	}
}

func TestCalcHistogramEmpty(t *testing.T) {
	h := CalcHistogram([]float64{})
	for label, count := range h {
		if count != 0 {
			t.Errorf("bucket %s: expected 0, got %d", label, count)
		}
	}
}
