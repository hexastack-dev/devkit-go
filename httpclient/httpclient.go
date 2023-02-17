package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Config struct {
	// InsecureSkipVerify determine wheter to use tls.Config InsecureSkipVerify or not.
	// Setting this to true is generally not recommended. Defaults to false.
	InsecureSkipVerify bool
	// CACerts list of custom CA certificates that we need to trust. This certificates
	// will be appended to host root CA if possible.
	CACerts []string
	// Timeout specifies http.Client timeout. If value is zero, will use default value
	// of 30 seconds.
	Timeout time.Duration
}

func New(config Config) (*http.Client, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	var (
		rootCAs *x509.CertPool
		err     error
	)
	if len(config.CACerts) > 0 {
		if rootCAs, err = loadCACerts(config.CACerts); err != nil {
			return nil, err
		}
	}

	tlsConfig, err := newTLSConfig(rootCAs, config.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	transport := newTransport(tlsConfig)
	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: otelhttp.NewTransport(transport),
	}

	return client, nil
}

func newTLSConfig(rootCAs *x509.CertPool, insecureSkipVerify bool) (*tls.Config, error) {
	tlsCofig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		RootCAs:            rootCAs,
	}
	return tlsCofig, nil
}

func loadCACerts(cacerts []string) (*x509.CertPool, error) {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		rootCAs = x509.NewCertPool()
	}

	for _, cacert := range cacerts {
		if err := appendCert(rootCAs, cacert); err != nil {
			return nil, fmt.Errorf("httpclient: failed to append certificate %s into rootCAs: %w", cacert, err)
		}
	}
	return rootCAs, nil
}

func appendCert(certPool *x509.CertPool, certfile string) error {
	absCertfile, err := filepath.Abs(certfile)
	if err != nil {
		return fmt.Errorf("httpclient: failed to locate file: %w", err)
	}
	cert, err := os.ReadFile(absCertfile)
	if err != nil {
		return fmt.Errorf("httpclient: failed to read file: %w", err)
	}

	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		return errors.New("httpclient: failed to append cert")
	}
	return nil
}

func newTransport(tlsConfig *tls.Config) *http.Transport {
	transport := &http.Transport{
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
	}
	return transport
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
