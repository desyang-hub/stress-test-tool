package reporters

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJSONReporterWrite(t *testing.T) {
	dir := t.TempDir()

	s := Stats{
		Name:            "json-test",
		TotalRequests:   200,
		Successful:      190,
		Failed:          10,
		TPS:             80.5,
		TotalDuration:   5 * time.Second,
		LatencyMin:      0.5,
		LatencyMean:     15.0,
		LatencyMedian:   12.0,
		LatencyP90:      30.0,
		LatencyP95:      45.0,
		LatencyP99:      80.0,
		LatencyMax:      200.0,
		LatencyStdDev:   10.0,
		StatusCodeBreakdown: map[int]int64{200: 180, 404: 10},
		ErrorBreakdown:    map[string]int64{"not_found": 10},
		Histogram: map[string]int64{
			"<10ms":  100,
			"10-50ms": 80,
			">50ms":  20,
		},
	}

	r := NewJSONReporter(dir)
	if err := r.Write(s); err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(dir, "report.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	var report jsonReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}

	if report.Name != "json-test" {
		t.Errorf("name = %q, want %q", report.Name, "json-test")
	}
	if report.TotalRequests != 200 {
		t.Errorf("total_requests = %d, want 200", report.TotalRequests)
	}
	if report.Successful != 190 {
		t.Errorf("successful = %d, want 190", report.Successful)
	}
	if report.Failed != 10 {
		t.Errorf("failed = %d, want 10", report.Failed)
	}
	if report.TPS != 80.5 {
		t.Errorf("tps = %f, want 80.5", report.TPS)
	}
	if report.LatencyMean != 15.0 {
		t.Errorf("latency_mean_ms = %f, want 15.0", report.LatencyMean)
	}
	if report.StatusCodeDist[200] != 180 {
		t.Errorf("status_code_dist[200] = %d, want 180", report.StatusCodeDist[200])
	}
	if report.ErrorBreakdown["not_found"] != 10 {
		t.Errorf("error_breakdown[not_found] = %d, want 10", report.ErrorBreakdown["not_found"])
	}
	if report.Histogram["<10ms"] != 100 {
		t.Errorf("histogram[<10ms] = %d, want 100", report.Histogram["<10ms"])
	}
	if report.GeneratedAt == "" {
		t.Error("generated_at should not be empty")
	}
}

func TestJSONReporterOutputDir(t *testing.T) {
	s := Stats{TotalRequests: 10}
	r := NewJSONReporter("")
	if err := r.Write(s); err != nil {
		t.Fatal(err)
	}
	os.Remove("report.json")
}
