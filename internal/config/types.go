// Package config defines the test configuration structures and provides
// YAML loading, environment variable resolution, and validation.
package config

import "time"

// TestConfig is the top-level configuration for a stress test.
type TestConfig struct {
	Name           string          `yaml:"name"`
	Description    string          `yaml:"description"`
	DefaultRequest DefaultRequest  `yaml:"default_request"`
	Stages         []Stage         `yaml:"stages"`
	Timeout        *Duration       `yaml:"timeout"`
	TLS            TLSConfig       `yaml:"tls"`
	CookieJar      bool            `yaml:"cookie_jar"`
	Output         OutputConfig    `yaml:"output"`
	Requests       []Request       `yaml:"requests"`
}

// DefaultRequest provides default values for all requests in the test.
type DefaultRequest struct {
	Method          string            `yaml:"method"`
	URL             string            `yaml:"url"`
	Headers         map[string]string `yaml:"headers"`
	Body            string            `yaml:"body"`
	AuthToken       string            `yaml:"auth_token"`
	Timeout         *Duration         `yaml:"timeout"`
	FollowRedirects bool              `yaml:"follow_redirects"`
	Retries         int               `yaml:"retries"`
	RetryDelay      *Duration         `yaml:"retry_delay"`
	QueryParams     map[string]string `yaml:"query_params"`
}

// Stage defines one phase of the concurrent gradient test.
type Stage struct {
	Duration    Duration `yaml:"duration"`
	Concurrency int      `yaml:"concurrency"`
	RateLimit   *uint64  `yaml:"rate_limit"` // nil = no limit
	Name        string   `yaml:"name"`
}

// Duration wraps time.Duration for YAML unmarshaling.
type Duration struct {
	time.Duration
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}

// TLSConfig holds TLS client settings.
type TLSConfig struct {
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	CertFile           string `yaml:"cert_file"`
	KeyFile            string `yaml:"key_file"`
	CAFile             string `yaml:"ca_file"`
	MinVersion         string `yaml:"min_version"`
}

// OutputConfig controls report generation.
type OutputConfig struct {
	Directory string   `yaml:"directory"`
	Formats   []string `yaml:"formats"`
}

// Request is a single HTTP request definition within a test.
type Request struct {
	ID              string            `yaml:"id"`
	Method          string            `yaml:"method"`
	URL             string            `yaml:"url"`
	Headers         map[string]string `yaml:"headers"`
	Body            string            `yaml:"body"`
	AuthToken       string            `yaml:"auth_token"`
	Timeout         *Duration         `yaml:"timeout"`
	FollowRedirects *bool             `yaml:"follow_redirects"`
	Retries         *int              `yaml:"retries"`
	RetryDelay      *Duration         `yaml:"retry_delay"`
	QueryParams     map[string]string `yaml:"query_params"`
	Assertions      []Assertion       `yaml:"assertions"`
	Tags            map[string]string `yaml:"tags"`
}

// Assertion defines a validation rule on the response.
type Assertion struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // status, latency, body, header, json
	Target   string `yaml:"target"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
}
