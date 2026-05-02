package config

import (
	"fmt"
	"net/url"
	"slices"
)

// Validate checks the configuration for correctness.
// Returns a list of all validation errors (does not return on first error).
func Validate(cfg *TestConfig) []error {
	var errs []error

	if cfg.Name == "" {
		errs = append(errs, fmt.Errorf("config: name is required"))
	}

	if len(cfg.Stages) == 0 {
		errs = append(errs, fmt.Errorf("config: at least one stage is required"))
	}

	for i, stage := range cfg.Stages {
		if stage.Duration.Duration == 0 {
			errs = append(errs, fmt.Errorf("config: stage %d (%s) duration must be positive", i, stage.Name))
		}
		if stage.Concurrency <= 0 {
			errs = append(errs, fmt.Errorf("config: stage %d (%s) concurrency must be > 0", i, stage.Name))
		}
	}

	// Validate request URLs
	for _, req := range cfg.Requests {
		if req.ID == "" {
			errs = append(errs, fmt.Errorf("config: request id is required"))
		}
		if req.URL != "" {
			if _, uerr := url.ParseRequestURI(req.URL); uerr != nil {
				errs = append(errs, fmt.Errorf("config: request %q invalid URL %q: %w", req.ID, req.URL, uerr))
			}
		}
		if req.Method != "" && !slices.Contains([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}, req.Method) {
			errs = append(errs, fmt.Errorf("config: request %q unsupported method %q", req.ID, req.Method))
		}

		for _, assert := range req.Assertions {
			if assert.Type != "" && !slices.Contains([]string{"status", "latency", "body", "header", "json"}, assert.Type) {
				errs = append(errs, fmt.Errorf("config: request %q assertion %q unsupported type %q", req.ID, assert.Name, assert.Type))
			}
		}
	}

	// Validate output formats
	validFormats := []string{"console", "html", "csv", "json"}
	for _, f := range cfg.Output.Formats {
		if !slices.Contains(validFormats, f) {
			errs = append(errs, fmt.Errorf("config: unsupported output format %q, valid: %v", f, validFormats))
		}
	}

	return errs
}
