package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"strings"
	"time"
)

// errInvalidHubCA is returned when KFLEET_HUB_CA holds no usable certificate.
var errInvalidHubCA = errors.New("KFLEET_HUB_CA must contain at least one PEM-encoded certificate")

// parseHubCA turns the PEM contents of KFLEET_HUB_CA into a root certificate
// pool. An empty value means the default system trust store, which is the
// current behavior when no custom CA is configured.
func parseHubCA(pemData string) (*x509.CertPool, error) {
	if strings.TrimSpace(pemData) == "" {
		return nil, nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(pemData)) {
		return nil, errInvalidHubCA
	}
	return pool, nil
}

// HubTLSConfig returns the TLS configuration used to connect to the hub, or
// nil when no custom CA bundle is configured. A nil result means "use system
// roots", which is also the zero-value TLS behavior.
func (c *Config) HubTLSConfig() *tls.Config {
	if c == nil || c.HubCAPool == nil {
		return nil
	}
	return &tls.Config{RootCAs: c.HubCAPool}
}

// HubHTTPClient returns an HTTP client for hub connections whose TLS root
// pool trusts the configured hub CA bundle, falling back to the default
// transport when no bundle is set. timeout bounds whole requests; pass 0 for
// long-lived connections such as the log channel WebSocket, where a request
// timeout would close the channel mid-stream.
func (c *Config) HubHTTPClient(timeout time.Duration) *http.Client {
	tlsConfig := c.HubTLSConfig()
	if tlsConfig == nil {
		return &http.Client{Timeout: timeout}
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig.RootCAs = tlsConfig.RootCAs
	return &http.Client{Timeout: timeout, Transport: transport}
}
