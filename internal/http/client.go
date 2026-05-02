package http

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/config"
)

// Client wraps http.Client with stress-test specific features.
type Client struct {
	httpClient *http.Client
	retries    int
	retryDelay time.Duration
	followRedir bool
}

// Option configures the Client.
type Option func(*Client)

// WithRetries sets the number of retry attempts on failure.
func WithRetries(n int) Option {
	return func(c *Client) { c.retries = n }
}

// WithRetryDelay sets the delay between retries.
func WithRetryDelay(d time.Duration) Option {
	return func(c *Client) { c.retryDelay = d }
}

// WithFollowRedirects controls redirect behavior.
func WithFollowRedirects(f bool) Option {
	return func(c *Client) { c.followRedir = f }
}

// WithTimeout sets the request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithTLSConfig sets custom TLS settings.
func WithTLSConfig(tlsCfg *config.TLSConfig) Option {
	return func(c *Client) {
		c.httpClient.Transport = BuildTransportWithConfig(tlsCfg)
	}
}

// WithCookieJar enables cookie persistence.
func WithCookieJar() Option {
	return func(c *Client) {
		jar, err := cookiejar.New(nil)
		if err == nil {
			c.httpClient.Jar = jar
		}
	}
}

// NewClient creates a new stress test HTTP client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Transport: DefaultTransport(),
			Timeout:   30 * time.Second,
		},
		retries:     0,
		retryDelay:  500 * time.Millisecond,
		followRedir: true,
	}

	for _, opt := range opts {
		opt(c)
	}

	if !c.followRedir {
		FollowRedirects(c.httpClient)
	}

	return c
}

// Do executes an HTTP request with optional retries.
func (c *Client) Do(req *http.Request) (*Response, error) {
	reqStart := time.Now()
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryDelay)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read body: %w", err)
			continue
		}

		response := &Response{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			BodySize:   int64(len(body)),
			Headers:    resp.Header,
			Latency:    time.Since(reqStart),
			Body:       body,
		}

		if resp.StatusCode >= 400 {
			return response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		return response, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retries+1, lastErr)
}

// ExecuteRequest runs a config.Request definition through the client.
func (c *Client) ExecuteRequest(req *config.Request, defaults *config.DefaultRequest) (*Response, error) {
	httpReq, err := BuildRequest(req, defaults)
	if err != nil {
		return nil, err
	}
	return c.Do(httpReq)
}
