package main

import (
	"context"
	"fmt"
	"os"

	"github.com/desyang-hub/stress-test-utils/internal/config"
	"github.com/desyang-hub/stress-test-utils/internal/engine"
	"github.com/desyang-hub/stress-test-utils/internal/metrics"
	"github.com/desyang-hub/stress-test-utils/internal/reporters"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "stresstest",
		Usage: "High-performance API stress testing tool",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run a stress test from YAML config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to YAML/JSON config file",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "output-dir",
						Usage: "Directory for report output files",
						Value: ".",
					},
					&cli.StringSliceFlag{
						Name:  "formats",
						Usage: "Report formats: console,html,csv,json",
						Value: []string{"console"},
					},
				},
				Action: runAction,
			},
			{
				Name:  "validate",
				Usage: "Validate a config file without running tests",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "config",
						Aliases:  []string{"c"},
						Usage:    "Path to YAML/JSON config file",
						Required: true,
					},
				},
				Action: validateAction,
			},
		},
		Version: "1.2.1",
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runAction(ctx context.Context, c *cli.Command) error {
	cfgPath := c.String("config")
	outputDir := c.String("output-dir")
	formats := c.StringSlice("formats")

	// Load config
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Validate config
	if errs := config.Validate(cfg); len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		return fmt.Errorf("invalid config (%d errors)", len(errs))
	}

	// Update output config from flags
	if outputDir != "" {
		cfg.Output.Directory = outputDir
	}
	if len(formats) > 0 {
		cfg.Output.Formats = formats
	}

	// Run engine
	e := engine.New(cfg)
	mStats, err := e.Run()
	if err != nil {
		return fmt.Errorf("run test: %w", err)
	}

	// Convert to reporter stats
	stats := toReporterStats(mStats)

	// Write reports
	if err := writeReports(stats, cfg.Output); err != nil {
		return fmt.Errorf("write reports: %w", err)
	}

	return nil
}

func validateAction(ctx context.Context, c *cli.Command) error {
	cfgPath := c.String("config")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	errs := config.Validate(cfg)
	if len(errs) > 0 {
		fmt.Printf("Config %s: INVALID\n", cfgPath)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		return fmt.Errorf("%d validation errors", len(errs))
	}

	fmt.Printf("Config %s: valid\n", cfgPath)
	fmt.Printf("  Name: %s\n", cfg.Name)
	fmt.Printf("  Stages: %d\n", len(cfg.Stages))
	fmt.Printf("  Requests: %d\n", len(cfg.Requests))
	return nil
}

func toReporterStats(s metrics.Stats) reporters.Stats {
	sb := make(map[int]int64, len(s.StatusCodeBreakdown))
	for k, v := range s.StatusCodeBreakdown {
		sb[k] = v
	}
	eb := make(map[string]int64, len(s.ErrorBreakdown))
	for k, v := range s.ErrorBreakdown {
		eb[string(k)] = v
	}
	return reporters.Stats{
		TotalRequests:       s.TotalRequests,
		Successful:          s.Successful,
		Failed:              s.Failed,
		TPS:                 s.TPS,
		TotalDuration:       s.TotalDuration,
		LatencyMin:          s.LatencyMin,
		LatencyMean:         s.LatencyMean,
		LatencyMedian:       s.LatencyMedian,
		LatencyP90:          s.LatencyP90,
		LatencyP95:          s.LatencyP95,
		LatencyP99:          s.LatencyP99,
		LatencyMax:          s.LatencyMax,
		LatencyStdDev:       s.LatencyStdDev,
		StatusCodeBreakdown: sb,
		ErrorBreakdown:      eb,
		Histogram:           s.Histogram,
	}
}

func writeReports(stats reporters.Stats, output config.OutputConfig) error {
	if output.Directory != "" {
		if err := os.MkdirAll(output.Directory, 0755); err != nil {
			return err
		}
	}

	for _, format := range output.Formats {
		switch format {
		case "console":
			r := reporters.NewConsoleReporter()
			r.PrintSummary(stats)
		case "html":
			r := reporters.NewHTMLReporter(output.Directory)
			if err := r.Write(stats); err != nil {
				return err
			}
		case "json":
			r := reporters.NewJSONReporter(output.Directory)
			if err := r.Write(stats); err != nil {
				return err
			}
		case "csv":
			fmt.Fprintf(os.Stderr, "  CSV export: not yet supported in this version\n")
		default:
			fmt.Fprintf(os.Stderr, "  Unknown format: %s\n", format)
		}
	}

	return nil
}
