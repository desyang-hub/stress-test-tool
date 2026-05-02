package metrics

import (
	"net/http"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// ErrorType categorizes the failure reason for statistical grouping.
type ErrorType string

const (
	ErrorTimeout    ErrorType = "timeout"
	ErrorConnection ErrorType = "connection_refused"
	ErrorTLS        ErrorType = "tls_error"
	ErrorHTTP       ErrorType = "http_error"
	ErrorAssertion  ErrorType = "assertion_failed"
	ErrorOther      ErrorType = "other"
)

// Result represents a single completed HTTP request.
type Result struct {
	RequestID     string        // unique request identifier
	Method        string        // HTTP method
	URL           string        // request URL
	StatusCode    int           // HTTP status code
	Status        string        // status text (e.g. "OK")
	Latency       time.Duration // total request latency
	ResponseSize  int64         // response body size in bytes
	Error         string        // error message, empty on success
	ErrorType     ErrorType     // error category
	Timestamp     time.Time     // request start time
	Stage         int           // stage index this request belongs to
	WorkerID      int           // worker goroutine ID
	Headers       http.Header   // response headers
	AssertionFail []string      // names of failed assertions
}

// Stats holds aggregated statistics across all completed requests.
type Stats struct {
	TotalRequests         int64             `json:"total_requests"`
	Successful            int64             `json:"successful"`
	Failed                int64             `json:"failed"`
	TotalBytes            int64             `json:"total_bytes_received"`
	TotalDuration         time.Duration     `json:"total_duration"`
	TPS                   float64           `json:"tps"`
	QPS                   float64           `json:"qps"`
	LatencyMin            float64           `json:"min_ms"`
	LatencyMax            float64           `json:"max_ms"`
	LatencyMean           float64           `json:"mean_ms"`
	LatencyMedian         float64           `json:"p50_ms"`
	LatencyP90            float64           `json:"p90_ms"`
	LatencyP95            float64           `json:"p95_ms"`
	LatencyP99            float64           `json:"p99_ms"`
	LatencyStdDev         float64           `json:"stddev_ms"`
	ErrorBreakdown        map[ErrorType]int64 `json:"error_breakdown"`
	StatusCodeBreakdown   map[int]int64     `json:"status_code_breakdown"`
	Histogram             map[string]int64  `json:"histogram"`
}

// LiveStats provides real-time statistics for the progress bar and live console.
type LiveStats struct {
	TotalRequests  int64
	Successful     int64
	Failed         int64
	CurrentRPS     float64
	CurrentLatency float64 // rolling average ms
	Elapsed        time.Duration
}

// Tracker collects metrics from multiple goroutines concurrently.
// Uses atomic.Int64 for safe concurrent access on all platforms including 32-bit.
type Tracker struct {
	mu             sync.RWMutex
	latencies      []float64           // in milliseconds, unsorted (sorted during Snapshot)
	errorBreakdown map[ErrorType]int64
	statusCodeDist map[int]int64
	totalBytes     atomic.Int64
	totalDuration  atomic.Int64 // nanoseconds
	requestCount   atomic.Int64
	successCount   atomic.Int64
	failureCount   atomic.Int64
	stageTimes     map[int]time.Duration
	stageCounts    map[int]int64
	windowStart    time.Time
	windowLatency  []float64
}

// NewTracker creates an initialized tracker.
func NewTracker() *Tracker {
	return &Tracker{
		latencies:      make([]float64, 0, 1024),
		errorBreakdown: make(map[ErrorType]int64),
		statusCodeDist: make(map[int]int64),
		stageTimes:     make(map[int]time.Duration),
		stageCounts:    make(map[int]int64),
		windowStart:    time.Now(),
		windowLatency:  make([]float64, 0, 256),
	}
}

// Record adds a completed request's result to the tracker.
// Safe to call from multiple goroutines concurrently.
func (t *Tracker) Record(r *Result) {
	t.requestCount.Add(1)

	ms := float64(r.Latency.Microseconds()) / 1000.0

	t.mu.Lock()
	t.latencies = append(t.latencies, ms)
	t.mu.Unlock()

	t.totalDuration.Add(r.Latency.Nanoseconds())

	// Track rolling window for live RPS
	t.mu.Lock()
	t.windowLatency = append(t.windowLatency, ms)
	t.mu.Unlock()

	isError := r.Error != "" || len(r.AssertionFail) > 0
	if isError {
		t.failureCount.Add(1)
		if r.ErrorType != "" {
			t.mu.Lock()
			t.errorBreakdown[r.ErrorType]++
			t.mu.Unlock()
		}
	} else {
		t.successCount.Add(1)
	}

	t.mu.Lock()
	t.statusCodeDist[r.StatusCode]++
	t.mu.Unlock()

	t.totalBytes.Add(r.ResponseSize)

	t.mu.Lock()
	t.stageTimes[r.Stage] += r.Latency
	t.stageCounts[r.Stage]++
	t.mu.Unlock()
}

// Snapshot returns a computed Stats object with all aggregations.
func (t *Tracker) Snapshot() Stats {
	t.mu.RLock()

	total := t.requestCount.Load()
	success := t.successCount.Load()
	failure := t.failureCount.Load()

	latCopy := make([]float64, len(t.latencies))
	copy(latCopy, t.latencies)
	slices.Sort(latCopy)

	p := CalcPercentiles(latCopy, 50, 90, 95, 99)

	var mean float64
	totalDuration := t.totalDuration.Load()
	if total > 0 {
		mean = float64(totalDuration) / float64(total) / float64(time.Millisecond)
	}

	stats := Stats{
		TotalRequests:         total,
		Successful:            success,
		Failed:                failure,
		TotalBytes:            t.totalBytes.Load(),
		TotalDuration:         time.Duration(totalDuration),
		LatencyMin:            minOrZero(latCopy),
		LatencyMax:            maxOrZero(latCopy),
		LatencyMean:           mean,
		LatencyMedian:         p[50],
		LatencyP90:            p[90],
		LatencyP95:            p[95],
		LatencyP99:            p[99],
		ErrorBreakdown:        make(map[ErrorType]int64),
		StatusCodeBreakdown:   make(map[int]int64),
		Histogram:             CalcHistogram(latCopy),
	}

	for k, v := range t.errorBreakdown {
		stats.ErrorBreakdown[k] = v
	}
	for k, v := range t.statusCodeDist {
		stats.StatusCodeBreakdown[k] = v
	}

	if total > 0 && stats.TotalDuration > 0 {
		stats.TPS = float64(total) / stats.TotalDuration.Seconds()
		stats.QPS = stats.TPS
	}

	if len(latCopy) > 1 {
		stats.LatencyStdDev = CalcStandardDeviation(latCopy, mean)
	}

	t.mu.RUnlock()
	return stats
}

// LiveSnapshot returns a LiveStats for real-time progress display.
func (t *Tracker) LiveSnapshot() LiveStats {
	total := t.requestCount.Load()
	success := t.successCount.Load()
	failure := t.failureCount.Load()

	t.mu.RLock()
	windowLat := make([]float64, len(t.windowLatency))
	copy(windowLat, t.windowLatency)
	t.mu.RUnlock()

	var avgLat, currentRPS float64
	if len(windowLat) > 0 {
		var sum float64
		for _, v := range windowLat {
			sum += v
		}
		avgLat = sum / float64(len(windowLat))
	}

	elapsed := time.Since(t.windowStart)
	if elapsed > 0 {
		currentRPS = float64(total) / elapsed.Seconds()
	}

	return LiveStats{
		TotalRequests:  total,
		Successful:     success,
		Failed:         failure,
		CurrentRPS:     currentRPS,
		CurrentLatency: avgLat,
		Elapsed:        elapsed,
	}
}

// Reset clears all collected metrics.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.latencies = t.latencies[:0]
	t.errorBreakdown = make(map[ErrorType]int64)
	t.statusCodeDist = make(map[int]int64)
	t.totalBytes.Store(0)
	t.totalDuration.Store(0)
	t.requestCount.Store(0)
	t.successCount.Store(0)
	t.failureCount.Store(0)
	t.stageTimes = make(map[int]time.Duration)
	t.stageCounts = make(map[int]int64)
	t.windowStart = time.Now()
	t.windowLatency = t.windowLatency[:0]
}

func minOrZero(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	return a[0]
}

func maxOrZero(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	return a[len(a)-1]
}
