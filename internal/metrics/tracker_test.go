package metrics

import (
	"sync"
	"testing"
	"time"
)

func TestTrackerRecord(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{
		RequestID:    "req-1",
		Method:       "GET",
		URL:          "http://example.com",
		StatusCode:   200,
		Status:       "OK",
		Latency:      50 * time.Millisecond,
		ResponseSize: 1024,
		Stage:        0,
	})

	stats := tracker.Snapshot()
	if stats.TotalRequests != 1 {
		t.Errorf("total requests = %d, want 1", stats.TotalRequests)
	}
	if stats.Successful != 1 {
		t.Errorf("successful = %d, want 1", stats.Successful)
	}
	if stats.Failed != 0 {
		t.Errorf("failed = %d, want 0", stats.Failed)
	}
	if stats.LatencyMean != 50 {
		t.Errorf("latency mean = %v, want 50", stats.LatencyMean)
	}
}

func TestTrackerRecordError(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{
		RequestID: "req-err",
		Method:    "GET",
		URL:       "http://example.com/broken",
		StatusCode: 0,
		Latency:    5000 * time.Millisecond,
		Error:      "connection refused",
		ErrorType:  ErrorConnection,
		Stage:      0,
	})

	stats := tracker.Snapshot()
	if stats.Failed != 1 {
		t.Errorf("failed = %d, want 1", stats.Failed)
	}
	if stats.ErrorBreakdown[ErrorConnection] != 1 {
		t.Errorf("error breakdown connection = %d, want 1", stats.ErrorBreakdown[ErrorConnection])
	}
}

func TestTrackerRecordAssertionFail(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{
		RequestID:   "req-assert",
		Method:      "GET",
		URL:         "http://example.com/api",
		StatusCode:  200,
		Latency:     100 * time.Millisecond,
		AssertionFail: []string{"status mismatch"},
		Stage:       0,
	})

	stats := tracker.Snapshot()
	if stats.Failed != 1 {
		t.Errorf("failed = %d, want 1", stats.Failed)
	}
}

func TestTrackerConcurrent(t *testing.T) {
	tracker := NewTracker()
	var wg sync.WaitGroup
	n := 1000

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tracker.Record(&Result{
				RequestID:    string(rune('a' + idx%26)),
				Method:       "GET",
				URL:          "http://example.com",
				StatusCode:   200 + idx%5,
				Latency:      time.Duration(10+idx%100) * time.Millisecond,
				ResponseSize: 512,
				Stage:        idx % 3,
			})
		}(i)
	}
	wg.Wait()

	stats := tracker.Snapshot()
	if stats.TotalRequests != int64(n) {
		t.Errorf("total requests = %d, want %d", stats.TotalRequests, n)
	}
}

func TestTrackerSnapshot(t *testing.T) {
	tracker := NewTracker()
	for i := 0; i < 10; i++ {
		tracker.Record(&Result{
			RequestID:    string(rune('a' + i)),
			Method:       "GET",
			URL:          "http://example.com",
			StatusCode:   200,
			Latency:      time.Duration(10*(i+1)) * time.Millisecond,
			ResponseSize: 100,
			Stage:        0,
		})
	}

	stats := tracker.Snapshot()
	if stats.LatencyMin != 10 {
		t.Errorf("latency min = %v, want 10", stats.LatencyMin)
	}
	if stats.LatencyMax != 100 {
		t.Errorf("latency max = %v, want 100", stats.LatencyMax)
	}
	if stats.LatencyMedian < 30 || stats.LatencyMedian > 70 {
		t.Errorf("P50 = %v, expected ~45-55", stats.LatencyMedian)
	}
}

func TestTrackerReset(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{
		RequestID: "req-1",
		StatusCode: 200,
		Latency:   50 * time.Millisecond,
		Stage:     0,
	})
	tracker.Reset()

	stats := tracker.Snapshot()
	if stats.TotalRequests != 0 {
		t.Errorf("after reset, total requests = %d, want 0", stats.TotalRequests)
	}
}

func TestTrackerLiveSnapshot(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{
		RequestID: "req-1",
		StatusCode: 200,
		Latency:   50 * time.Millisecond,
		Stage:     0,
	})

	ls := tracker.LiveSnapshot()
	if ls.TotalRequests != 1 {
		t.Errorf("live total = %d, want 1", ls.TotalRequests)
	}
	if ls.CurrentLatency != 50 {
		t.Errorf("live latency = %v, want 50", ls.CurrentLatency)
	}
}

func BenchmarkTrackerRecord(b *testing.B) {
	tracker := NewTracker()
	r := &Result{
		RequestID:    "test",
		StatusCode:   200,
		Latency:      10 * time.Millisecond,
		ResponseSize: 1024,
		Stage:        0,
	}
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tracker.Record(r)
		}
	})
}

func BenchmarkTrackerSnapshot(b *testing.B) {
	tracker := NewTracker()
	for i := 0; i < 10000; i++ {
		tracker.Record(&Result{
			RequestID:    "test",
			StatusCode:   200,
			Latency:      time.Duration(i%100) * time.Millisecond,
			ResponseSize: 512,
			Stage:        0,
		})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Snapshot()
	}
}

func BenchmarkCalcPercentiles(b *testing.B) {
	data := make([]float64, 10000)
	for i := range data {
		data[i] = float64(i)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		CalcPercentiles(data, 50, 90, 95, 99)
	}
}

func TestTrackerStatusCodeBreakdown(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{StatusCode: 200, Stage: 0})
	tracker.Record(&Result{StatusCode: 200, Stage: 0})
	tracker.Record(&Result{StatusCode: 404, Stage: 0})
	tracker.Record(&Result{StatusCode: 500, Stage: 0})

	stats := tracker.Snapshot()
	if stats.StatusCodeBreakdown[200] != 2 {
		t.Errorf("200 count = %d, want 2", stats.StatusCodeBreakdown[200])
	}
	if stats.StatusCodeBreakdown[404] != 1 {
		t.Errorf("404 count = %d, want 1", stats.StatusCodeBreakdown[404])
	}
	if stats.StatusCodeBreakdown[500] != 1 {
		t.Errorf("500 count = %d, want 1", stats.StatusCodeBreakdown[500])
	}
}

func TestTrackerTotalBytes(t *testing.T) {
	tracker := NewTracker()
	tracker.Record(&Result{ResponseSize: 100, Stage: 0})
	tracker.Record(&Result{ResponseSize: 200, Stage: 0})
	tracker.Record(&Result{ResponseSize: 300, Stage: 0})

	stats := tracker.Snapshot()
	if stats.TotalBytes != 600 {
		t.Errorf("total bytes = %d, want 600", stats.TotalBytes)
	}
}

func TestTrackerTPS(t *testing.T) {
	tracker := NewTracker()
	for i := 0; i < 10; i++ {
		tracker.Record(&Result{
			StatusCode: 200,
			Latency:    10 * time.Millisecond,
			Stage:      0,
		})
	}

	stats := tracker.Snapshot()
	if stats.TPS <= 0 {
		t.Errorf("TPS should be > 0, got %v", stats.TPS)
	}
	if stats.QPS != stats.TPS {
		t.Errorf("QPS = %v, should equal TPS = %v", stats.QPS, stats.TPS)
	}
}

func TestTrackerEmptySnapshot(t *testing.T) {
	tracker := NewTracker()
	stats := tracker.Snapshot()
	if stats.TotalRequests != 0 {
		t.Errorf("empty tracker: total = %d, want 0", stats.TotalRequests)
	}
	if stats.LatencyMin != 0 {
		t.Errorf("empty tracker: min = %v, want 0", stats.LatencyMin)
	}
}
