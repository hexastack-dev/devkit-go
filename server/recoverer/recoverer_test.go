package recoverer_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexastack-dev/devkit-go/server/recoverer"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	panic(errors.New("Ooopsie"))
}

func TestRecoverer_DefaultHandler(t *testing.T) {
	h := recoverer.New(nil)(http.HandlerFunc(handleHello))
	req, _ := http.NewRequest("GET", "/hello", nil)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("response should be Internal Server Error: %d", rr.Code)
	}
}

func handleError(w http.ResponseWriter, r *http.Request) {
	var res struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
	res.Status = http.StatusBadGateway
	if err, ok := recoverer.GetErrFromContext(r.Context()); ok {
		res.Message = fmt.Sprintf("got panic from: %v", err)
	}

	if b, err := json.Marshal(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "this should not happened: %v", err)
	} else {
		w.WriteHeader(http.StatusBadGateway)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(b)
	}
}

func TestRecoverer_CustomHandler(t *testing.T) {
	h := recoverer.New(http.HandlerFunc(handleError))(http.HandlerFunc(handleHello))
	req, _ := http.NewRequest("GET", "/hello", nil)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusBadGateway {
		t.Errorf("response should be Bad Gateway: %d", rr.Code)
	}
	expected := `{"status":502,"message":"got panic from: Ooopsie"}`
	if b, err := io.ReadAll(res.Body); err != nil {
		t.Errorf("failed to read response body: %v", err)
	} else {
		body := string(b)
		if body != expected {
			t.Errorf("response body should equals expected: %v", body)
		}
	}
	res.Body.Close()
}
