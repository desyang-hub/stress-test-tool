# Stress Test Tool

High-performance API stress testing tool for Go.

## Features

- HTTP method support: GET, POST, PUT, DELETE
- Custom headers, body, and query parameters
- Concurrent gradient testing with configurable stages
- RPS rate limiting via token bucket algorithm
- SSL/TLS with configurable minimum version
- Environment variable references (`${VAR}`, `$VAR`)
- Assertion evaluation: status code, latency, body, headers, JSON
- Multi-format reports: console, HTML, JSON, CSV
- Graceful shutdown on SIGINT/SIGTERM

## Installation

### Depends
- Go v1.24
- golangci-lint


### From source

```bash
go build -o stresstest ./cmd/stresstest/
```

### Using Docker

```bash
docker build -t stresstest .
docker run stresstest run --config examples/basic.yml
```

## Quick Start

Create a config file:

```yaml
name: my-api-test
default_request:
  method: GET
  url: http://localhost:8080/api/health
stages:
  - concurrency: 100
    duration: 30s
output:
  directory: ./reports
  formats: [console, html, json]
```

Run the test:

```bash
./stresstest run --config my-test.yml
```

Validate a config without running:

```bash
./stresstest validate --config my-test.yml
```

## Configuration

### Top-level fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Test name |
| `description` | string | Test description |
| `default_request` | object | Default HTTP request settings |
| `stages` | array | Test phases (sequential) |
| `timeout` | duration | Global request timeout |
| `tls` | object | TLS configuration |
| `cookie_jar` | bool | Enable cookie persistence |
| `requests` | array | Additional request definitions |
| `output` | object | Report configuration |

### Default request

| Field | Type | Description |
|-------|------|-------------|
| `method` | string | HTTP method |
| `url` | string | Target URL |
| `headers` | map | Request headers |
| `body` | string | Request body |
| `auth_token` | string | Bearer token |
| `timeout` | duration | Per-request timeout |
| `follow_redirects` | bool | Follow HTTP redirects |
| `retries` | int | Number of retry attempts |
| `retry_delay` | duration | Delay between retries |
| `query_params` | map | URL query parameters |

### Stage

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Stage name |
| `concurrency` | int | Number of concurrent goroutines |
| `duration` | duration | How long to run |
| `rate_limit` | int | Maximum RPS |

## Report Formats

- **console**: Human-readable terminal output with latency table, status codes, and histogram
- **html**: Styled HTML report with CSS-only charts (no JavaScript)
- **json**: Machine-readable JSON for integration with dashboards
- **csv**: Per-request results (per-request data export)

## Examples

See the `examples/` directory for configuration templates:

- `basic.yml` - Simple single-stage test
- `gradient.yml` - Multi-stage gradient load test with rate limiting
- `enterprise.yml` - Full-featured test with assertions, TLS, and long-running stages

## Build

```bash
make build      # Build the binary
make test       # Run tests
make bench      # Run benchmarks
make lint       # Run linter
```

## Cross-Compilation

```bash
GOOS=linux GOARCH=amd64 go build -o stresstest-linux ./cmd/stresstest/
GOOS=windows GOARCH=amd64 go build -o stresstest-windows.exe ./cmd/stresstest/
GOOS=darwin GOARCH=arm64 go build -o stresstest-darwin ./cmd/stresstest/
```
