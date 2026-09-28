package network

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// DefaultUserAgent used across all HTTP queries.
const DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// ClientConfig holds customization parameters for constructing an HTTP client.
type ClientConfig struct {
	Timeout             time.Duration
	DialTimeout         time.Duration
	KeepAlive           time.Duration
	TLSHandshakeTimeout time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	InsecureSkipVerify  bool
	UserAgent           string
}

// DefaultClientConfig returns robust, production-ready connection pool defaults.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Timeout:             60 * time.Second,
		DialTimeout:         10 * time.Second,
		KeepAlive:           30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		InsecureSkipVerify:  true,
		UserAgent:           DefaultUserAgent,
	}
}

// NewClient constructs an http.Client with connection pooling, timeouts, and redirect handling.
func NewClient(cfg ...ClientConfig) *http.Client {
	c := DefaultClientConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   c.DialTimeout,
			KeepAlive: c.KeepAlive,
		}).DialContext,
		MaxIdleConns:          c.MaxIdleConns,
		MaxIdleConnsPerHost:   c.MaxIdleConnsPerHost,
		IdleConnTimeout:       c.IdleConnTimeout,
		TLSHandshakeTimeout:   c.TLSHandshakeTimeout,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify},
		ForceAttemptHTTP2:     false,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   c.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			ua := c.UserAgent
			if ua == "" {
				ua = DefaultUserAgent
			}
			req.Header.Set("User-Agent", ua)
			if len(via) > 0 {
				req.Header.Set("Referer", via[len(via)-1].URL.String())
				if rangeHdr := via[0].Header.Get("Range"); rangeHdr != "" {
					req.Header.Set("Range", rangeHdr)
				}
			}
			return nil
		},
	}
}
