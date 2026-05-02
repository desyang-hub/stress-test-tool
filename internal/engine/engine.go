// Package engine orchestrates multi-stage stress tests.
package engine

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/assertions"
	"github.com/desyang-hub/stress-test-utils/internal/config"
	"github.com/desyang-hub/stress-test-utils/internal/concurrency"
	stresstesthttp "github.com/desyang-hub/stress-test-utils/internal/http"
	"github.com/desyang-hub/stress-test-utils/internal/metrics"
)

// Engine orchestrates a multi-stage stress test.
type Engine struct {
	cfg           *config.TestConfig
	defaultReq    *config.DefaultRequest
	httpClient    *stresstesthttp.Client
	evaluator     *assertions.Evaluator
	tracker       *metrics.Tracker
	coordinator   *Coordinator
	requestPool   *RequestPool
	defaultTimeout time.Duration
}

// RequestPool provides sequential access to configured requests.
type RequestPool struct {
	reqs []*config.Request
	idx  int
	mu   sync.Mutex
}

// NewRequestPool creates a request pool from the test config.
func NewRequestPool(cfg *config.TestConfig) *RequestPool {
	if len(cfg.Requests) == 0 {
		if cfg.DefaultRequest.URL != "" {
			retries := cfg.DefaultRequest.Retries
			return &RequestPool{
				reqs: []*config.Request{
					{
						ID:          "default",
						Method:      cfg.DefaultRequest.Method,
						URL:         cfg.DefaultRequest.URL,
						Headers:     cfg.DefaultRequest.Headers,
						Body:        cfg.DefaultRequest.Body,
						AuthToken:   cfg.DefaultRequest.AuthToken,
						Timeout:     cfg.DefaultRequest.Timeout,
						Retries:     &retries,
						QueryParams: cfg.DefaultRequest.QueryParams,
					},
				},
			}
		}
		return &RequestPool{reqs: []*config.Request{}}
	}

	poolReqs := make([]*config.Request, len(cfg.Requests))
	for i := range cfg.Requests {
		poolReqs[i] = &cfg.Requests[i]
	}
	return &RequestPool{reqs: poolReqs}
}

// Next returns the next request in round-robin order.
func (rp *RequestPool) Next() *config.Request {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	if len(rp.reqs) == 0 {
		return nil
	}
	req := rp.reqs[rp.idx%len(rp.reqs)]
	rp.idx++
	return req
}

// New creates a new Engine.
func New(cfg *config.TestConfig) *Engine {
	timeout := 30 * time.Second
	if cfg.Timeout != nil {
		timeout = cfg.Timeout.Duration
	}

	opts := []stresstesthttp.Option{
		stresstesthttp.WithTimeout(timeout),
		stresstesthttp.WithRetries(cfg.DefaultRequest.Retries),
	}
	if cfg.CookieJar {
		opts = append(opts, stresstesthttp.WithCookieJar())
	}
	if cfg.TLS.InsecureSkipVerify || cfg.TLS.CertFile != "" || cfg.TLS.CAFile != "" {
		opts = append(opts, stresstesthttp.WithTLSConfig(&cfg.TLS))
	}

	e := &Engine{
		cfg:            cfg,
		defaultReq:     &cfg.DefaultRequest,
		httpClient:     stresstesthttp.NewClient(opts...),
		evaluator:      &assertions.Evaluator{},
		tracker:        metrics.NewTracker(),
		requestPool:    NewRequestPool(cfg),
		defaultTimeout: timeout,
	}
	return e
}

// Run executes all stages sequentially, returning final stats.
func (e *Engine) Run() (metrics.Stats, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.coordinator = NewCoordinator(ctx, cancel)
	e.coordinator.Start()

	var wg sync.WaitGroup

	for stageIdx, stageCfg := range e.cfg.Stages {
		if e.coordinator.ShutdownRequested() {
			break
		}

		var limiter *concurrency.TokenBucket
		if stageCfg.RateLimit != nil {
			limiter = concurrency.NewTokenBucket(*stageCfg.RateLimit, *stageCfg.RateLimit)
		}
		sem := concurrency.NewSemaphore(stageCfg.Concurrency)

		workers := make([]*Worker, stageCfg.Concurrency)
		for i := 0; i < stageCfg.Concurrency; i++ {
			workers[i] = NewWorker(i, e, e.requestPool, limiter, sem, int64(stageIdx))
			wg.Add(1)
			go func(w *Worker) {
				defer wg.Done()
				w.Run(ctx)
			}(workers[i])
		}

		select {
		case <-time.After(stageCfg.Duration.Duration):
		case <-e.coordinator.ShutdownChannel():
			cancel()
			wg.Wait()
			goto done
		}

		cancel()
		wg.Wait()

		// Reset for next stage (skip after last stage)
		if stageIdx < len(e.cfg.Stages)-1 {
			ctx, cancel = context.WithCancel(context.Background())
			e.coordinator.ResetContext(ctx, cancel)
			e.tracker.Reset()
		}

		wg = sync.WaitGroup{}
		for i := 0; i < stageCfg.Concurrency; i++ {
			workers[i] = NewWorker(i, e, e.requestPool, limiter, sem, int64(stageIdx))
			wg.Add(1)
			go func(w *Worker) {
				defer wg.Done()
				w.Run(ctx)
			}(workers[i])
		}
	}

done:
	cancel()
	e.coordinator.Stop()
	return e.tracker.Snapshot(), nil
}

// RandInt returns a random int for body templating.
func RandInt() int {
	return rand.Intn(1000000)
}

// classifyNetworkError categorizes network errors.
func classifyNetworkError(err error) metrics.ErrorType {
	if err == nil {
		return ""
	}
	errStr := err.Error()
	switch {
	case containsStr(errStr, "timeout"), containsStr(errStr, "i/o timeout"):
		return metrics.ErrorTimeout
	case containsStr(errStr, "connection refused"), containsStr(errStr, "dial"):
		return metrics.ErrorConnection
	case containsStr(errStr, "tls"):
		return metrics.ErrorTLS
	default:
		return metrics.ErrorOther
	}
}

func classifyError(err string) metrics.ErrorType {
	switch {
	case len(err) == 0:
		return ""
	case containsStr(err, "timeout"), containsStr(err, "i/o timeout"):
		return metrics.ErrorTimeout
	case containsStr(err, "connection refused"), containsStr(err, "dial"):
		return metrics.ErrorConnection
	case containsStr(err, "tls"):
		return metrics.ErrorTLS
	case containsStr(err, "assertion"):
		return metrics.ErrorAssertion
	default:
		return metrics.ErrorOther
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubStr(s, substr))
}

func findSubStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
