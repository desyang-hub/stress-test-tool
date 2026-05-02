package reporters

import (
	"encoding/csv"
	"os"
	"time"
)

// CSVReporter writes per-request results to a CSV file.
type CSVReporter struct {
	OutputDir string
}

// NewCSVReporter creates a CSV reporter.
func NewCSVReporter(outputDir string) *CSVReporter {
	return &CSVReporter{OutputDir: outputDir}
}

// Write writes the CSV report.
func (r *CSVReporter) Write(results []CSVRow) error {
	filename := "results.csv"
	if r.OutputDir != "" {
		if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
			return err
		}
		filename = r.OutputDir + "/" + filename
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Header
	if err := w.Write([]string{"timestamp", "request_id", "method", "url", "status_code", "latency_ms", "response_size", "error"}); err != nil {
		return err
	}

	for _, row := range results {
		if err := w.Write([]string{
			row.Timestamp.Format(time.RFC3339),
			row.RequestID,
			row.Method,
			row.URL,
			row.StatusCode,
			row.LatencyMS,
			row.ResponseSize,
			row.Error,
		}); err != nil {
			return err
		}
	}

	return nil
}

// CSVRow holds one row of CSV data.
type CSVRow struct {
	Timestamp    time.Time
	RequestID    string
	Method       string
	URL          string
	StatusCode   string
	LatencyMS    string
	ResponseSize string
	Error        string
}
