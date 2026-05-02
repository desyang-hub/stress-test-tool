package http

import (
	"net/http"
	"time"
)

// Response wraps an http.Response with timing and error information.
type Response struct {
	StatusCode    int           // HTTP status code
	Status        string        // status text
	BodySize      int64         // response body size in bytes
	Headers       http.Header   // response headers
	Latency       time.Duration // total round-trip time
	TLSHandshake  time.Duration // TLS handshake duration (0 if not TLS)
	DNSResolve    time.Duration // DNS resolution duration (0 if cached)
	Error         string        // error message, empty on success
	ErrorCategory string        // error category for grouping
	Body          []byte        // response body (consumed by client)
}

// NewResponse creates a Response from an http.Response and timing info.
func NewResponse(resp *http.Response, reqStart time.Time, timings *ResponseTimings) *Response {
	r := &Response{
		StatusCode:   resp.StatusCode,
		Status:       resp.Status,
		BodySize:     0,
		Headers:      resp.Header,
		Latency:      time.Since(reqStart),
		TLSHandshake: timings.TLSHandshake,
		DNSResolve:   timings.DNS,
	}

	if resp.Body != nil {
		// We can't read the body here without consuming it.
		// The size will be set by the caller after reading.
		r.BodySize = resp.ContentLength
	}

	return r
}

// IsSuccess checks if the response indicates a successful request.
func (r *Response) IsSuccess() bool {
	return r.Error == "" && r.StatusCode >= 200 && r.StatusCode < 400
}

// IsServerError checks if the response is a server error (5xx).
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// ResponseTimings captures detailed timing breakdown.
type ResponseTimings struct {
	DNS          time.Duration
	TCPConnect   time.Duration
	TLSHandshake time.Duration
	Wait         time.Duration // time waiting for response
	Download     time.Duration // body download time
}
