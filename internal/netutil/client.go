package netutil

import (
	"context"
	"net"
	"net/http"
	"time"
)

// GetHttpClient returns a pre-configured http.Client that is optimized for CLI usage
// and includes a more resilient DNS resolver to handle issues common in environments
// like Android (Termux) where the default Go resolver might fail on [::1]:53.
func GetHttpClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Custom resolver that tries the system DNS first, but falls back to public DNS
	// if the system resolver is unreachable or fails (common in Termux).
	dialer.Resolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second * 5,
			}
			
			// Try the system-provided DNS address first
			conn, err := d.DialContext(ctx, network, address)
			if err == nil {
				return conn, nil
			}

			// If that fails (e.g. connection refused on [::1]:53), 
			// fall back to a reliable public DNS.
			// Cloudflare (1.1.1.1) and Google (8.8.8.8) are good candidates.
			return d.DialContext(ctx, network, "1.1.1.1:53")
		},
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
