package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/config"
	"github.com/desyang-hub/stress-test-utils/internal/engine"
)

func dur(d time.Duration) config.Duration {
	return config.Duration{Duration: d}
}

func TestEngineBasicRun(t *testing.T) {
	var count int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&count, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	}))
	defer server.Close()

	cfg := config.TestConfig{
		DefaultRequest: config.DefaultRequest{
			URL:    server.URL + "/api/test",
			Method: "GET",
		},
		Stages: []config.Stage{
			{
				Concurrency: 10,
				Duration:    dur(time.Second * 2),
			},
		},
	}

	e := engine.New(&cfg)
	stats, err := e.Run()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalRequests == 0 {
		t.Error("expected at least some requests")
	}
	if stats.Failed > 0 {
		t.Errorf("expected no failures, got %d", stats.Failed)
	}
	if stats.TPS < 1 {
		t.Errorf("unexpectedly low TPS: %.1f", stats.TPS)
	}
}

func TestEngineWithAssertions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Response-Time", "50")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"hello"}`)
	}))
	defer server.Close()

	cfg := config.TestConfig{
		DefaultRequest: config.DefaultRequest{
			URL:    server.URL + "/api/assert",
			Method: "GET",
		},
		Requests: []config.Request{
			{
				URL: server.URL + "/api/assert",
				Assertions: []config.Assertion{
					{
						Name:     "status_ok",
						Type:     "status",
						Target:   "status",
						Operator: "equal",
						Value:    "200",
					},
					{
						Name:     "latency_ok",
						Type:     "latency",
						Target:   "latency",
						Operator: "less_than",
						Value:    "5000",
					},
					{
						Name:     "header_present",
						Type:     "header",
						Target:   "X-Response-Time",
						Operator: "exists",
					},
				},
			},
		},
		Stages: []config.Stage{
			{
				Concurrency: 5,
				Duration:    dur(time.Second * 2),
			},
		},
	}

	e := engine.New(&cfg)
	stats, err := e.Run()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalRequests == 0 {
		t.Error("expected at least some requests")
	}
}

func TestEngineMultipleStagesAccumulate(t *testing.T) {
	// Verify that stats from multiple stages are accumulated (not wiped)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TestConfig{
		DefaultRequest: config.DefaultRequest{
			URL:    server.URL + "/api/accumulate",
			Method: "GET",
		},
		Stages: []config.Stage{
			{
				Concurrency: 5,
				Duration:    dur(time.Second * 1),
			},
			{
				Concurrency: 10,
				Duration:    dur(time.Second * 1),
			},
			{
				Concurrency: 15,
				Duration:    dur(time.Second * 1),
			},
		},
	}

	e := engine.New(&cfg)
	stats, err := e.Run()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalRequests == 0 {
		t.Error("expected at least some requests")
	}
	// Three stages should accumulate meaningful requests
	if stats.TotalRequests < 30 {
		t.Errorf("expected at least 30 total requests across 3 stages, got %d", stats.TotalRequests)
	}
}

func TestEngineMultipleStages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TestConfig{
		DefaultRequest: config.DefaultRequest{
			URL:    server.URL + "/api/stages",
			Method: "GET",
		},
		Stages: []config.Stage{
			{
				Concurrency: 5,
				Duration:    dur(time.Second * 1),
			},
			{
				Concurrency: 10,
				Duration:    dur(time.Second * 1),
			},
		},
	}

	e := engine.New(&cfg)
	stats, err := e.Run()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalRequests == 0 {
		t.Error("expected at least some requests")
	}
	// Two stages of ~1 second each should produce meaningful total
	if stats.TotalRequests < 10 {
		t.Errorf("expected at least 10 total requests across 2 stages, got %d", stats.TotalRequests)
	}
}

func TestEngineWithRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rateLimit := uint64(20)
	cfg := config.TestConfig{
		DefaultRequest: config.DefaultRequest{
			URL:    server.URL + "/api/rate",
			Method: "GET",
		},
		Stages: []config.Stage{
			{
				Concurrency: 50,
				RateLimit:   &rateLimit,
				Duration:    dur(time.Second * 2),
			},
		},
	}

	e := engine.New(&cfg)
	stats, err := e.Run()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalRequests == 0 {
		t.Error("expected at least some requests")
	}
	// With 20 RPS for 2 seconds, expect approximately 40 requests
	if stats.TPS < 5 {
		t.Errorf("unexpectedly low TPS: %.1f (expected ~20)", stats.TPS)
	}
}

func TestEnginePOSTWithBody(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.TestConfig{
		DefaultRequest: config.DefaultRequest{
			URL:    server.URL + "/api/post",
			Method: "POST",
			Body:   `{"key":"value"}`,
		},
		Requests: []config.Request{
			{
				Method: "POST",
				URL:    server.URL + "/api/post",
				Body:   `{"key":"value"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
		Stages: []config.Stage{
			{
				Concurrency: 3,
				Duration:    dur(time.Second * 1),
			},
		},
	}

	e := engine.New(&cfg)
	stats, err := e.Run()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalRequests == 0 {
		t.Error("expected at least some requests")
	}
	if receivedBody != `{"key":"value"}` {
		t.Errorf("expected body %q, got %q", `{"key":"value"}`, receivedBody)
	}
}
