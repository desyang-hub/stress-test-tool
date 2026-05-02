package reporters

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCSVReporterWrite(t *testing.T) {
	dir := t.TempDir()

	results := []CSVRow{
		{
			Timestamp:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			RequestID:    "req-1",
			Method:       "GET",
			URL:          "http://example.com/api",
			StatusCode:   "200",
			LatencyMS:    "15.5",
			ResponseSize: "1024",
			Error:        "",
		},
		{
			Timestamp:    time.Date(2025, 1, 1, 0, 0, 1, 0, time.UTC),
			RequestID:    "req-2",
			Method:       "POST",
			URL:          "http://example.com/api",
			StatusCode:   "500",
			LatencyMS:    "120.3",
			ResponseSize: "256",
			Error:        "timeout",
		},
	}

	r := NewCSVReporter(dir)
	if err := r.Write(results); err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(dir, "results.csv")
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	// Header + 2 data rows
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2 data), got %d", len(rows))
	}

	// Check header
	expectedHeader := []string{"timestamp", "request_id", "method", "url", "status_code", "latency_ms", "response_size", "error"}
	for i, col := range expectedHeader {
		if rows[0][i] != col {
			t.Errorf("header[%d] = %q, want %q", i, rows[0][i], col)
		}
	}

	// Check first data row
	if rows[1][2] != "GET" {
		t.Errorf("row[1].method = %q, want %q", rows[1][2], "GET")
	}
	if rows[1][4] != "200" {
		t.Errorf("row[1].status_code = %q, want %q", rows[1][4], "200")
	}

	// Check second data row
	if rows[2][7] != "timeout" {
		t.Errorf("row[2].error = %q, want %q", rows[2][7], "timeout")
	}
}

func TestCSVReporterNoOutputDir(t *testing.T) {
	results := []CSVRow{
		{RequestID: "r1", Method: "GET", URL: "http://test", StatusCode: "200"},
	}
	r := NewCSVReporter("")
	// Creates in current directory; just verify no panic
	if err := r.Write(results); err != nil {
		t.Fatal(err)
	}
	// Clean up the file we created
	os.Remove("results.csv")
}
