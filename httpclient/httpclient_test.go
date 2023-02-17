package httpclient_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexastack-dev/devkit-go/httpclient"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Hello World")
}

func BenchmarkNew(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(helloHandler))
	defer srv.Close()

	client, err := httpclient.New(httpclient.Config{})
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		client.Get(srv.URL)
	}
}
