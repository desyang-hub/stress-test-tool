package engine

import (
	"context"
	"net/http"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/assertions"
	"github.com/desyang-hub/stress-test-utils/internal/config"
	"github.com/desyang-hub/stress-test-utils/internal/concurrency"
	stresstesthttp "github.com/desyang-hub/stress-test-utils/internal/http"
	"github.com/desyang-hub/stress-test-utils/internal/metrics"
)

// Worker is a goroutine that repeatedly sends HTTP requests.
type Worker struct {
	ID          int
	engine      *Engine
	requestPool *RequestPool
	limiter     *concurrency.TokenBucket
	sem         *concurrency.Semaphore
	stageIdx    int64
}

// NewWorker creates a new worker.
func NewWorker(id int, engine *Engine, pool *RequestPool, limiter *concurrency.TokenBucket, sem *concurrency.Semaphore, stageIdx int64) *Worker {
	return &Worker{
		ID:          id,
		engine:      engine,
		requestPool: pool,
		limiter:     limiter,
		sem:         sem,
		stageIdx:    stageIdx,
	}
}

// Run is the worker's main loop. It exits when the context is cancelled.
func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Apply rate limiting
		if w.limiter != nil {
			w.limiter.Wait()
		}

		// Acquire concurrency slot
		w.sem.Acquire(1)

		// Get next request
		req := w.requestPool.Next()
		if req == nil {
			w.sem.Release(1)
			continue
		}

		// Build and send HTTP request
		start := time.Now()
		httpReq, err := stresstesthttp.BuildRequest(req, w.engine.defaultReq)
		if err != nil {
			result := &metrics.Result{
				RequestID: req.ID,
				Method:    req.Method,
				URL:       req.URL,
				Latency:   time.Since(start),
				Error:     err.Error(),
				ErrorType: metrics.ErrorOther,
				Timestamp: start,
				Stage:     int(w.stageIdx),
				WorkerID:  w.ID,
			}
			w.engine.tracker.Record(result)
			w.sem.Release(1)
			continue
		}

		resp, err := w.engine.httpClient.Do(httpReq)
		latency := time.Since(start)

		if err != nil {
			result := &metrics.Result{
				RequestID: req.ID,
				Method:    req.Method,
				URL:       req.URL,
				Latency:   latency,
				Error:     err.Error(),
				ErrorType: classifyNetworkError(err),
				Timestamp: start,
				Stage:     int(w.stageIdx),
				WorkerID:  w.ID,
			}
			w.engine.tracker.Record(result)
			w.sem.Release(1)
			continue
		}

		// Evaluate assertions
		result := &metrics.Result{
			RequestID:    req.ID,
			Method:       req.Method,
			URL:          req.URL,
			StatusCode:   resp.StatusCode,
			Status:       http.StatusText(resp.StatusCode),
			Latency:      latency,
			Timestamp:    start,
			Stage:        int(w.stageIdx),
			WorkerID:     w.ID,
		}

		if len(req.Assertions) > 0 {
			body := resp.Body
			if body == nil {
				body = []byte{}
			}
			respInfo := &assertions.ResponseInfo{
				StatusCode: resp.StatusCode,
				Latency:    latency,
				Headers:    respHeadersMap(resp),
				Body:       body,
			}

			asserts := make([]*config.Assertion, len(req.Assertions))
			for i := range req.Assertions {
				asserts[i] = &req.Assertions[i]
			}
			failures := w.engine.evaluator.EvaluateAll(asserts, respInfo)
			for _, f := range failures {
				if !f.Pass {
					result.AssertionFail = append(result.AssertionFail, f.Name)
					if f.Error != "" {
						result.Error = f.Error
					} else {
						result.Error = "assertion failed: " + f.Name
					}
				}
			}
		}

		if result.Error != "" {
			result.ErrorType = classifyError(result.Error)
		}

		w.engine.tracker.Record(result)
		w.sem.Release(1)
	}
}

func respHeadersMap(resp *stresstesthttp.Response) map[string]string {
	h := make(map[string]string)
	for k, v := range resp.Headers {
		if len(v) > 0 {
			h[k] = v[0]
		}
	}
	return h
}
