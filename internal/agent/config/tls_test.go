package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"
)

// selfSignedPEM mints a throwaway self-signed certificate and returns it as
// PEM, the way a private CA bundle or self-signed hub certificate looks.
func selfSignedPEM(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "kfleet-hub-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestParseHubCA(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantErr bool
	}{
		{name: "empty uses system roots", input: "", wantNil: true},
		{name: "whitespace only uses system roots", input: "  \n\t ", wantNil: true},
		{name: "valid certificate", input: selfSignedPEM(t)},
		{name: "valid certificate with surrounding whitespace", input: "\n" + selfSignedPEM(t) + "\n"},
		{name: "garbage is rejected", input: "not a certificate", wantErr: true},
		{name: "private key block is not a certificate", input: "-----BEGIN PRIVATE KEY-----\nAAAA\n-----END PRIVATE KEY-----\n", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pool, err := parseHubCA(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("parseHubCA(%q) error = nil, want invalid CA error", test.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHubCA(%q) error = %v", test.input, err)
			}
			if test.wantNil && pool != nil {
				t.Fatalf("parseHubCA(%q) pool = %v, want nil", test.input, pool)
			}
			if !test.wantNil && pool == nil {
				t.Fatalf("parseHubCA(%q) pool = nil, want populated pool", test.input)
			}
		})
	}
}

func TestLoadReadsHubCA(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	t.Setenv("KFLEET_HUB_CA", selfSignedPEM(t))
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HubCAPool == nil {
		t.Fatal("HubCAPool = nil, want populated pool")
	}
	if cfg.HubTLSConfig() == nil || cfg.HubTLSConfig().RootCAs != cfg.HubCAPool {
		t.Fatal("HubTLSConfig() = nil, want RootCAs pointing at the parsed pool")
	}
}

func TestLoadFailsFastOnInvalidHubCA(t *testing.T) {
	t.Setenv("KFLEET_HUB_URL", "https://hub.example.test")
	t.Setenv("KFLEET_CLUSTER_NAME", "production")
	t.Setenv("KFLEET_HUB_TOKEN", "secret")
	t.Setenv("KFLEET_HUB_CA", "-----BEGIN CERTIFICATE-----\nZm9v\n-----END CERTIFICATE-----\n")
	_, err := Load()
	if err == nil {
		t.Fatal("Load() with an invalid KFLEET_HUB_CA returned nil error")
	}
	if got := err.Error(); got != errInvalidHubCA.Error() {
		t.Fatalf("Load() error = %q, want %q", got, errInvalidHubCA.Error())
	}
}

func TestHubHTTPClientWithoutCABundleMatchesDefaults(t *testing.T) {
	cfg := &Config{}
	client := cfg.HubHTTPClient(10 * time.Second)
	if client.Timeout != 10*time.Second {
		t.Fatalf("Timeout = %v, want 10s", client.Timeout)
	}
	if client.Transport != nil {
		t.Fatalf("Transport = %T, want default (nil)", client.Transport)
	}
	if cfg.HubTLSConfig() != nil {
		t.Fatal("HubTLSConfig() != nil without a CA bundle")
	}
}

func TestHubHTTPClientWithCABundleTrustsPool(t *testing.T) {
	pool := x509.NewCertPool()
	cfg := &Config{HubCAPool: pool}
	client := cfg.HubHTTPClient(0)
	if client.Timeout != 0 {
		t.Fatalf("Timeout = %v, want 0 for long-lived connections", client.Timeout)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("Transport = %T, want *http.Transport", client.Transport)
	}
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.RootCAs != pool {
		t.Fatal("transport RootCAs does not trust the configured hub CA pool")
	}
	tlsConfig := cfg.HubTLSConfig()
	if tlsConfig == nil || tlsConfig.RootCAs != pool {
		t.Fatal("HubTLSConfig() does not trust the configured hub CA pool")
	}
}
