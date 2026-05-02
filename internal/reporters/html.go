package reporters

import (
	"html/template"
	"os"
	"time"
)

// HTMLReporter generates an HTML report.
type HTMLReporter struct {
	OutputDir string
}

// NewHTMLReporter creates an HTML reporter.
func NewHTMLReporter(outputDir string) *HTMLReporter {
	return &HTMLReporter{OutputDir: outputDir}
}

// Write saves the HTML report to a file.
func (r *HTMLReporter) Write(s Stats) error {
	if r.OutputDir != "" {
		if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
			return err
		}
	}

	filename := "report.html"
	if r.OutputDir != "" {
		filename = r.OutputDir + "/" + filename
	}

	tmpl := template.Must(template.New("report").Parse(htmlTemplate))

	data := reportData{
		Title:         s.Name,
		GeneratedAt:   time.Now().Format(time.RFC3339),
		TotalRequests: s.TotalRequests,
		Successful:    s.Successful,
		Failed:        s.Failed,
		SuccessRate:   successRateFloat(s),
		TPS:           s.TPS,
		Duration:      s.TotalDuration.Round(time.Millisecond).String(),
		LatencyMin:    s.LatencyMin,
		LatencyMean:   s.LatencyMean,
		LatencyMedian: s.LatencyMedian,
		LatencyP90:    s.LatencyP90,
		LatencyP95:    s.LatencyP95,
		LatencyP99:    s.LatencyP99,
		LatencyMax:    s.LatencyMax,
		StatusCodes:   statusRows(s.StatusCodeBreakdown, s.TotalRequests),
		Histogram:     histogramRows(s.Histogram, s.TotalRequests),
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

type reportData struct {
	Title         string
	GeneratedAt   string
	TotalRequests int64
	Successful    int64
	Failed        int64
	SuccessRate   float64
	TPS           float64
	Duration      string
	LatencyMin    float64
	LatencyMean   float64
	LatencyMedian float64
	LatencyP90    float64
	LatencyP95    float64
	LatencyP99    float64
	LatencyMax    float64
	StatusCodes   []statusRow
	Histogram     []histRow
}

type statusRow struct {
	Code  int
	Count int64
	Pct   float64
}

type histRow struct {
	Key   string
	Count int64
	Pct   float64
}

func statusRows(breakdown map[int]int64, total int64) []statusRow {
	rows := make([]statusRow, 0, len(breakdown))
	for code, count := range breakdown {
		rows = append(rows, statusRow{
			Code:  code,
			Count: count,
			Pct:   float64(count) / float64(total) * 100,
		})
	}
	return rows
}

func histogramRows(buckets map[string]int64, total int64) []histRow {
	rows := make([]histRow, 0, len(buckets))
	for key, count := range buckets {
		rows = append(rows, histRow{
			Key:   key,
			Count: count,
			Pct:   float64(count) / float64(total) * 100,
		})
	}
	return rows
}

func successRateFloat(s Stats) float64 {
	if s.TotalRequests == 0 {
		return 0
	}
	return float64(s.Successful) / float64(s.TotalRequests) * 100
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Stress Test Report{{if .Title}} - {{.Title}}{{end}}</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #f5f5f5; color: #333; padding: 20px; }
.container { max-width: 960px; margin: 0 auto; }
h1 { font-size: 1.5rem; margin-bottom: 8px; color: #1a1a1a; }
.subtitle { color: #666; margin-bottom: 24px; font-size: 0.9rem; }
.card { background: #fff; border-radius: 8px; padding: 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 16px; }
.stat-item { text-align: center; padding: 12px; }
.stat-value { font-size: 1.8rem; font-weight: 700; color: #1a73e8; }
.stat-label { font-size: 0.8rem; color: #666; margin-top: 4px; }
.stat-value.fail { color: #d32f2f; }
.stat-value.success { color: #2e7d32; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #eee; }
th { font-weight: 600; color: #555; font-size: 0.85rem; }
td { font-size: 0.9rem; }
.pct-bar { background: #1a73e8; height: 20px; border-radius: 4px; min-width: 2px; }
.histogram-bar { background: #2e7d32; height: 16px; border-radius: 3px; }
.footer { text-align: center; color: #999; font-size: 0.8rem; margin-top: 24px; }
</style>
</head>
<body>
<div class="container">
<h1>Stress Test Report{{if .Title}} - {{.Title}}{{end}}</h1>
<p class="subtitle">Generated at {{.GeneratedAt}}</p>

<div class="card">
<div class="stats-grid">
<div class="stat-item"><div class="stat-value">{{.TotalRequests}}</div><div class="stat-label">Total Requests</div></div>
<div class="stat-item"><div class="stat-value success">{{.Successful}}</div><div class="stat-label">Successful</div></div>
<div class="stat-item"><div class="stat-value fail">{{.Failed}}</div><div class="stat-label">Failed</div></div>
<div class="stat-item"><div class="stat-value">{{printf "%.1f" .TPS}}</div><div class="stat-label">Req/s</div></div>
<div class="stat-item"><div class="stat-value">{{.Duration}}</div><div class="stat-label">Duration</div></div>
</div>
</div>

<div class="card">
<h2 style="margin-bottom:12px;font-size:1rem;">Latency Percentiles (ms)</h2>
<table>
<tr><th>Percentile</th><th>Value (ms)</th></tr>
<tr><td>Min</td><td>{{printf "%.1f" .LatencyMin}}</td></tr>
<tr><td>Mean</td><td>{{printf "%.1f" .LatencyMean}}</td></tr>
<tr><td>P50</td><td>{{printf "%.1f" .LatencyMedian}}</td></tr>
<tr><td>P90</td><td>{{printf "%.1f" .LatencyP90}}</td></tr>
<tr><td>P95</td><td>{{printf "%.1f" .LatencyP95}}</td></tr>
<tr><td>P99</td><td>{{printf "%.1f" .LatencyP99}}</td></tr>
<tr><td>Max</td><td>{{printf "%.1f" .LatencyMax}}</td></tr>
</table>
</div>

<div class="card">
<h2 style="margin-bottom:12px;font-size:1rem;">Status Codes</h2>
<table>
<tr><th>Code</th><th>Count</th><th>Distribution</th></tr>
{{range .StatusCodes}}
<tr><td>{{.Code}}</td><td>{{.Count}}</td><td><div class="pct-bar" style="width:{{printf "%.0f" .Pct}}%"></div> {{printf "%.1f" .Pct}}%</td></tr>
{{end}}
</table>
</div>

{{if .Histogram}}
<div class="card">
<h2 style="margin-bottom:12px;font-size:1rem;">Latency Histogram</h2>
<table>
<tr><th>Bucket</th><th>Count</th><th>Distribution</th></tr>
{{range .Histogram}}
<tr><td>{{.Key}}</td><td>{{.Count}}</td><td><div class="histogram-bar" style="width:{{printf "%.0f" .Pct}}%"></div></td></tr>
{{end}}
</table>
</div>
{{end}}

<div class="footer">Generated by stress-test-utils</div>
</div>
</body>
</html>`
