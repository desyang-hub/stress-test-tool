package http

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/desyang-hub/stress-test-utils/internal/config"
)

// BuildTransport creates an http.Transport with the given TLS config.
func BuildTransport(tlsCfg *TLSConfig) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		ForceAttemptHTTP2:     true,
	}

	if tlsCfg != nil {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: tlsCfg.InsecureSkipVerify,
			MinVersion:         tlsVersion(tlsCfg.MinVersion),
		}

		if tlsCfg.CertFile != "" && tlsCfg.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
			if err == nil {
				transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
			}
		}

		if tlsCfg.CAFile != "" {
			caCert, err := os.ReadFile(tlsCfg.CAFile)
			if err == nil {
				caCertPool := x509.NewCertPool()
				caCertPool.AppendCertsFromPEM(caCert)
				transport.TLSClientConfig.RootCAs = caCertPool
			}
		}
	}

	return transport
}

// DefaultTransport returns a transport tuned for stress testing.
func DefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   500,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		ForceAttemptHTTP2:     true,
	}
}

// TLSConfig holds TLS settings for the HTTP client.
type TLSConfig struct {
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
	MinVersion         string
}

// FollowRedirects disables redirect following.
func FollowRedirects(c *http.Client) {
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func tlsVersion(s string) uint16 {
	switch s {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	default:
		return tls.VersionTLS12
	}
}

// BuildTransportWithConfig creates an http.Transport from config.TLSConfig.
func BuildTransportWithConfig(tlsCfg *config.TLSConfig) *http.Transport {
	return BuildTransport(&TLSConfig{
		InsecureSkipVerify: tlsCfg.InsecureSkipVerify,
		CertFile:           tlsCfg.CertFile,
		KeyFile:            tlsCfg.KeyFile,
		CAFile:             tlsCfg.CAFile,
		MinVersion:         tlsCfg.MinVersion,
	})
}
