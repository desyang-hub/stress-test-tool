// Package assertions provides response validation against configurable rules.
package assertions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/config"
)

// EvalResult holds the result of a single assertion evaluation.
type EvalResult struct {
	Name  string `json:"name"`
	Pass  bool   `json:"pass"`
	Actual  string `json:"actual"`
	Expected string `json:"expected"`
	Error   string `json:"error,omitempty"`
}

// Errorf creates a failed EvalResult with an error message.
func Errorf(name, expected, actual, format string, args ...any) EvalResult {
	msg := fmt.Sprintf(format, args...)
	return EvalResult{
		Name:     name,
		Pass:     false,
		Expected: expected,
		Actual:   actual,
		Error:    msg,
	}
}

// PassResult creates a passed EvalResult.
func PassResult(name string) EvalResult {
	return EvalResult{
		Name: name,
		Pass: true,
	}
}

// FailResult creates a failed EvalResult.
func FailResult(name, expected, actual string) EvalResult {
	return EvalResult{
		Name:     name,
		Pass:     false,
		Expected: expected,
		Actual:   actual,
	}
}

// Evaluator evaluates a config.Assertion against a response.
type Evaluator struct{}

// Evaluate runs a single assertion.
func (e *Evaluator) Evaluate(assert *config.Assertion, resp *ResponseInfo) EvalResult {
	switch assert.Type {
	case "status":
		return e.evalStatus(assert, resp)
	case "latency":
		return e.evalLatency(assert, resp)
	case "body":
		return e.evalBody(assert, resp)
	case "header":
		return e.evalHeader(assert, resp)
	case "json":
		return e.evalJSON(assert, resp)
	default:
		return Errorf(assert.Name, assert.Type, "unknown", "unsupported assertion type: %s", assert.Type)
	}
}

// EvaluateAll runs all assertions on a response.
func (e *Evaluator) EvaluateAll(asserts []*config.Assertion, resp *ResponseInfo) []EvalResult {
	var results []EvalResult
	for _, a := range asserts {
		results = append(results, e.Evaluate(a, resp))
	}
	return results
}

func (e *Evaluator) evalStatus(assert *config.Assertion, resp *ResponseInfo) EvalResult {
	statusStr := fmt.Sprintf("%d", resp.StatusCode)
	switch assert.Operator {
	case "equal", "==":
		if statusStr != assert.Value {
			return FailResult(assert.Name, assert.Value, statusStr)
		}
	case "not_equal", "!=":
		if statusStr == assert.Value {
			return FailResult(assert.Name, assert.Value+" (not equal)", statusStr)
		}
	case "in_range", "between":
		parts := strings.Split(assert.Value, "-")
		if len(parts) != 2 {
			return Errorf(assert.Name, assert.Value, statusStr, "invalid range: %s", assert.Value)
		}
		min, err := strconv.Atoi(parts[0])
		if err != nil {
			return Errorf(assert.Name, assert.Value, statusStr, "invalid min value: %v", err)
		}
		max, err := strconv.Atoi(parts[1])
		if err != nil {
			return Errorf(assert.Name, assert.Value, statusStr, "invalid max value: %v", err)
		}
		code := resp.StatusCode
		if code < min || code > max {
			return FailResult(assert.Name, fmt.Sprintf("%d-%d", min, max), fmt.Sprintf("%d", code))
		}
	default:
		return Errorf(assert.Name, assert.Operator, "status", "unsupported operator: %s", assert.Operator)
	}
	return PassResult(assert.Name)
}

func (e *Evaluator) evalLatency(assert *config.Assertion, resp *ResponseInfo) EvalResult {
	latencyMs := resp.Latency.Milliseconds()
	var valueMs int64
	if strings.HasSuffix(assert.Value, "ms") {
		v := strings.TrimSuffix(assert.Value, "ms")
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Errorf(assert.Name, assert.Value, fmt.Sprintf("%dms", latencyMs), "invalid latency: %v", err)
		}
		valueMs = int64(val)
	} else if strings.HasSuffix(assert.Value, "s") {
		v := strings.TrimSuffix(assert.Value, "s")
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Errorf(assert.Name, assert.Value, fmt.Sprintf("%dms", latencyMs), "invalid latency: %v", err)
		}
		valueMs = int64(val * 1000)
	} else {
		val, err := strconv.ParseInt(assert.Value, 10, 64)
		if err != nil {
			return Errorf(assert.Name, assert.Value, fmt.Sprintf("%dms", latencyMs), "invalid latency: %v", err)
		}
		valueMs = val
	}

	switch assert.Operator {
	case "less_than", "<":
		if latencyMs >= valueMs {
			return FailResult(assert.Name, fmt.Sprintf("<%dms", valueMs), fmt.Sprintf("%dms", latencyMs))
		}
	case "greater_than", ">":
		if latencyMs <= valueMs {
			return FailResult(assert.Name, fmt.Sprintf(">%dms", valueMs), fmt.Sprintf("%dms", latencyMs))
		}
	case "equal", "==":
		if latencyMs != valueMs {
			return FailResult(assert.Name, fmt.Sprintf("%dms", valueMs), fmt.Sprintf("%dms", latencyMs))
		}
	default:
		return Errorf(assert.Name, assert.Operator, "latency", "unsupported operator: %s", assert.Operator)
	}
	return PassResult(assert.Name)
}

func (e *Evaluator) evalBody(assert *config.Assertion, resp *ResponseInfo) EvalResult {
	body := string(resp.Body)
	switch assert.Operator {
	case "contains":
		if !strings.Contains(body, assert.Value) {
			return FailResult(assert.Name, fmt.Sprintf("contains %q", assert.Value), "body")
		}
	case "not_contains":
		if strings.Contains(body, assert.Value) {
			return FailResult(assert.Name, fmt.Sprintf("not contains %q", assert.Value), "body")
		}
	case "equal", "==":
		if body != assert.Value {
			return FailResult(assert.Name, assert.Value, body)
		}
	case "length":
		actualLen := strconv.Itoa(len(body))
		if actualLen != assert.Value {
			return FailResult(assert.Name, assert.Value, actualLen)
		}
	default:
		return Errorf(assert.Name, assert.Operator, "body", "unsupported operator: %s", assert.Operator)
	}
	return PassResult(assert.Name)
}

func (e *Evaluator) evalHeader(assert *config.Assertion, resp *ResponseInfo) EvalResult {
	actual := resp.Headers[assert.Target]
	switch assert.Operator {
	case "equal", "==":
		if actual != assert.Value {
			return FailResult(assert.Name, assert.Value, actual)
		}
	case "contains":
		if !strings.Contains(actual, assert.Value) {
			return FailResult(assert.Name, fmt.Sprintf("contains %q", assert.Value), actual)
		}
	case "present", "exists":
		if actual == "" {
			return FailResult(assert.Name, "present", "missing")
		}
	default:
		return Errorf(assert.Name, assert.Operator, "header", "unsupported operator: %s", assert.Operator)
	}
	return PassResult(assert.Name)
}

func (e *Evaluator) evalJSON(assert *config.Assertion, resp *ResponseInfo) EvalResult {
	var data map[string]any
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		return Errorf(assert.Name, assert.Value, "", "invalid JSON: %v", err)
	}

	// Simple JSON path resolution (supports dot notation)
	parts := strings.Split(assert.Target, ".")
	var current any = data
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return FailResult(assert.Name, "object", fmt.Sprintf("%T", current))
		}
		current = m[part]
	}

	switch assert.Operator {
	case "exists":
		if current == nil {
			return FailResult(assert.Name, "exists", "null")
		}
	case "equal", "==":
		actual := fmt.Sprintf("%v", current)
		if actual != assert.Value {
			return FailResult(assert.Name, assert.Value, actual)
		}
	case "type":
		switch current.(type) {
		case string, float64, bool, map[string]any, []any:
			// OK
		default:
			return FailResult(assert.Name, assert.Value, fmt.Sprintf("%T", current))
		}
	}
	return PassResult(assert.Name)
}

// ResponseInfo holds the data needed for assertion evaluation.
type ResponseInfo struct {
	StatusCode int
	Latency    time.Duration
	Headers    map[string]string
	Body       []byte
}
