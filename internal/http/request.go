package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/config"
)

// BuildRequest creates an http.Request from a config.Request definition.
func BuildRequest(req *config.Request, defaults *config.DefaultRequest) (*http.Request, error) {
	method := req.Method
	if method == "" && defaults != nil {
		method = defaults.Method
	}
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	bodyStr := req.Body
	if bodyStr == "" && defaults != nil {
		bodyStr = defaults.Body
	}
	if bodyStr != "" {
		body = strings.NewReader(bodyStr)
	}

	url := req.URL
	if url == "" && defaults != nil {
		url = defaults.URL
	}
	if url == "" {
		return nil, fmt.Errorf("no URL specified")
	}

	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	// Set headers
	headers := req.Headers
	if defaults != nil {
		for k, v := range defaults.Headers {
			if _, ok := headers[k]; !ok {
				headers[k] = v
			}
		}
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// Set auth token
	authToken := req.AuthToken
	if authToken == "" && defaults != nil {
		authToken = defaults.AuthToken
	}
	if authToken != "" {
		httpReq.Header.Set("Authorization", authToken)
	}

	// Set query params
	queryParams := req.QueryParams
	if defaults != nil {
		for k, v := range defaults.QueryParams {
			if _, ok := queryParams[k]; !ok {
				queryParams[k] = v
			}
		}
	}
	q := httpReq.URL.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	httpReq.URL.RawQuery = q.Encode()

	// Set timeout
	timeout := req.Timeout
	if timeout == nil && defaults != nil {
		timeout = defaults.Timeout
	}
	if timeout != nil {
		httpReq.Header.Set("X-Timeout", timeout.Duration.String())
	}

	return httpReq, nil
}

// BuildRequestWithTimeout is like BuildRequest but applies a global timeout override.
func BuildRequestWithTimeout(req *config.Request, defaults *config.DefaultRequest, timeout time.Duration) (*http.Request, error) {
	httpReq, err := BuildRequest(req, defaults)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-Global-Timeout", timeout.String())
	return httpReq, nil
}
