package assertions

import (
	"testing"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/config"
)

var evaluator = &Evaluator{}

func newResp(status int, latency time.Duration, headers map[string]string, body []byte) *ResponseInfo {
	h := make(map[string]string)
	for k, v := range headers {
		h[k] = v
	}
	return &ResponseInfo{StatusCode: status, Latency: latency, Headers: h, Body: body}
}

func mkAssert(name, typ, op, val, target string) *config.Assertion {
	return &config.Assertion{Name: name, Type: typ, Operator: op, Value: val, Target: target}
}

func TestEvalStatusEqual(t *testing.T) {
	r := evaluator.Evaluate(mkAssert("s1", "status", "equal", "200", ""), newResp(200, 0, nil, nil))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
	r = evaluator.Evaluate(mkAssert("s2", "status", "equal", "201", ""), newResp(200, 0, nil, nil))
	if r.Pass {
		t.Error("expected fail, got pass")
	}
}

func TestEvalStatusInRange(t *testing.T) {
	r := evaluator.Evaluate(mkAssert("s3", "status", "between", "200-299", ""), newResp(201, 0, nil, nil))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
	r = evaluator.Evaluate(mkAssert("s4", "status", "between", "200-299", ""), newResp(500, 0, nil, nil))
	if r.Pass {
		t.Error("expected fail, got pass")
	}
}

func TestEvalLatencyLessThan(t *testing.T) {
	r := evaluator.Evaluate(mkAssert("l1", "latency", "<", "100ms", ""), newResp(200, 50*time.Millisecond, nil, nil))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
	r = evaluator.Evaluate(mkAssert("l2", "latency", "<", "50ms", ""), newResp(200, 100*time.Millisecond, nil, nil))
	if r.Pass {
		t.Error("expected fail, got pass")
	}
}

func TestEvalLatencySeconds(t *testing.T) {
	r := evaluator.Evaluate(mkAssert("l3", "latency", "<", "1s", ""), newResp(200, 500*time.Millisecond, nil, nil))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
}

func TestEvalBodyContains(t *testing.T) {
	r := evaluator.Evaluate(mkAssert("b1", "body", "contains", "hello", ""), newResp(200, 0, nil, []byte("hello world")))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
	r = evaluator.Evaluate(mkAssert("b2", "body", "contains", "xyz", ""), newResp(200, 0, nil, []byte("hello world")))
	if r.Pass {
		t.Error("expected fail, got pass")
	}
}

func TestEvalHeaderPresent(t *testing.T) {
	r := evaluator.Evaluate(mkAssert("h1", "header", "exists", "", "Content-Type"), newResp(200, 0, map[string]string{"Content-Type": "application/json"}, nil))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
	r = evaluator.Evaluate(mkAssert("h2", "header", "exists", "", "X-Missing"), newResp(200, 0, map[string]string{}, nil))
	if r.Pass {
		t.Error("expected fail, got pass")
	}
}

func TestEvalJSONExists(t *testing.T) {
	body := []byte(`{"name": "test", "count": 5}`)
	r := evaluator.Evaluate(mkAssert("j1", "json", "exists", "", "name"), newResp(200, 0, nil, body))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
}

func TestEvalJSONEqual(t *testing.T) {
	body := []byte(`{"name": "test"}`)
	r := evaluator.Evaluate(mkAssert("j2", "json", "==", "test", "name"), newResp(200, 0, nil, body))
	if !r.Pass {
		t.Errorf("expected pass, got: %v", r)
	}
}

func TestEvalJSONBadData(t *testing.T) {
	body := []byte(`not json`)
	r := evaluator.Evaluate(mkAssert("j3", "json", "exists", "", "missing_key"), newResp(200, 0, nil, body))
	if r.Pass {
		t.Error("expected fail on invalid JSON")
	}
}

func TestEvaluateAll(t *testing.T) {
	asserts := []*config.Assertion{
		{Name: "status ok", Type: "status", Operator: "equal", Value: "200"},
		{Name: "fast", Type: "latency", Operator: "<", Value: "1000ms"},
	}
	body := []byte(`{"ok": true}`)
	results := evaluator.EvaluateAll(asserts, newResp(200, 50*time.Millisecond, map[string]string{"X-Custom": "val"}, body))
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if !results[0].Pass {
		t.Errorf("assert 0: %v", results[0])
	}
	if !results[1].Pass {
		t.Errorf("assert 1: %v", results[1])
	}
}
