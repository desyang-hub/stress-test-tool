package reporters

import (
	"encoding/json"
	"os"
	"time"
)

// JSONReporter writes machine-readable JSON output.
type JSONReporter struct {
	OutputDir string
}

// NewJSONReporter creates a JSON reporter.
func NewJSONReporter(outputDir string) *JSONReporter {
	return &JSONReporter{OutputDir: outputDir}
}

// Write writes the JSON report.
func (r *JSONReporter) Write(s Stats) error {
	filename := "report.json"
	if r.OutputDir != "" {
		if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
			return err
		}
		filename = r.OutputDir + "/" + filename
	}

	data := jsonReport{
		GeneratedAt:    time.Now().Format(time.RFC3339),
		Name:           s.Name,
		TotalRequests:  s.TotalRequests,
		Successful:     s.Successful,
		Failed:         s.Failed,
		TPS:            s.TPS,
		TotalDuration:  s.TotalDuration.String(),
		LatencyMin:     s.LatencyMin,
		LatencyMean:    s.LatencyMean,
		LatencyMedian:  s.LatencyMedian,
		LatencyP90:     s.LatencyP90,
		LatencyP95:     s.LatencyP95,
		LatencyP99:     s.LatencyP99,
		LatencyMax:     s.LatencyMax,
		LatencyStdDev:  s.LatencyStdDev,
		StatusCodeDist: s.StatusCodeBreakdown,
		ErrorBreakdown: s.ErrorBreakdown,
		Histogram:      s.Histogram,
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

type jsonReport struct {
	GeneratedAt    string           `json:"generated_at"`
	Name           string           `json:"name,omitempty"`
	TotalRequests  int64            `json:"total_requests"`
	Successful     int64            `json:"successful"`
	Failed         int64            `json:"failed"`
	TPS            float64          `json:"tps"`
	TotalDuration  string           `json:"total_duration"`
	LatencyMin     float64          `json:"latency_min_ms"`
	LatencyMean    float64          `json:"latency_mean_ms"`
	LatencyMedian  float64          `json:"latency_p50_ms"`
	LatencyP90     float64          `json:"latency_p90_ms"`
	LatencyP95     float64          `json:"latency_p95_ms"`
	LatencyP99     float64          `json:"latency_p99_ms"`
	LatencyMax     float64          `json:"latency_max_ms"`
	LatencyStdDev  float64          `json:"latency_stddev_ms"`
	StatusCodeDist map[int]int64    `json:"status_code_breakdown"`
	ErrorBreakdown map[string]int64 `json:"error_breakdown,omitempty"`
	Histogram      map[string]int64 `json:"histogram,omitempty"`
}
