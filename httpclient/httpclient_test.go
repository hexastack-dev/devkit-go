package httpclient_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexastack-dev/devkit-go/httpclient"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/trace"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Hello World")
}

func TestHttpClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(helloHandler))
	defer srv.Close()

	client, err := httpclient.New(httpclient.Config{})
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", srv.URL, nil)
	res, err := client.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello World"), b)
}

func TestHttpClient_WithTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(helloHandler))
	defer srv.Close()

	rootCAs, err := loadCerts(srv.TLS.Certificates)
	if err != nil {
		t.Fatal(err)
	}
	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	client, err := httpclient.New(httpclient.Config{})
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", srv.URL, nil)
	_, err = client.Do(req)
	var certErr *tls.CertificateVerificationError
	assert.True(t, errors.As(err, &certErr))

	client = httpclient.NewWithTransport(httpclient.NewTransport(tlsConfig), 30*time.Second)

	req, _ = http.NewRequest("GET", srv.URL, nil)
	res, err := client.Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, []byte("Hello World"), b)
}

func loadCerts(certs []tls.Certificate) (*x509.CertPool, error) {
	rootCAs := x509.NewCertPool()

	for _, c := range certs {
		roots, err := x509.ParseCertificates(c.Certificate[len(c.Certificate)-1])
		if err != nil {
			return nil, err
		}
		for _, root := range roots {
			rootCAs.AddCert(root)
		}
	}

	return rootCAs, nil
}

func BenchmarkHttpClient(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(helloHandler))
	defer srv.Close()

	client, err := httpclient.New(httpclient.Config{})
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", srv.URL, nil)
		client.Do(req)
	}
}

func BenchmarkHttpClient_WithContext(b *testing.B) {
	ctx := context.Background()
	tp := trace.NewTracerProvider()
	ctx, span := tp.Tracer("").Start(ctx, "testWithContext")
	defer span.End()

	srv := httptest.NewServer(http.HandlerFunc(helloHandler))
	defer srv.Close()

	client, err := httpclient.New(httpclient.Config{})
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", srv.URL, nil)
		client.Do(req)
	}
}
