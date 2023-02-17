package errors_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexastack-dev/devkit-go/errors"
	"github.com/stretchr/testify/assert"
)

var (
	errServerErr = &errors.ErrorWithHandler{
		Err:              errors.New("something went wrong"),
		StatusCode:       http.StatusInternalServerError,
		ErrorHandlerFunc: errHandlerFunc,
	}
	errNotFound = &errors.ErrorWithHandler{
		Err:        errors.New("not found"),
		StatusCode: http.StatusNotFound,
	}
)

func errHandlerFunc(err error, statusCode int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, `{"status":%d,"message":"%s"}`, statusCode, err.Error())
	}
}

func handleHTTP(err error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := err.(*errors.ErrorWithHandler); ok {
			h.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func TestErrorHandler_compability(t *testing.T) {
	err := fmt.Errorf("oopsie: %w", errServerErr)

	if !errors.Is(err, errServerErr) {
		t.Errorf("err should equivalent errServerErr: %v", err)
	}
}

func TestErrorWithHandler(t *testing.T) {
	h := http.HandlerFunc(handleHTTP(errServerErr))
	req, _ := http.NewRequest("GET", "/server-err", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res := rr.Result()
	b, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, `{"status":500,"message":"something went wrong"}`, string(b))

	h = http.HandlerFunc(handleHTTP(errNotFound))
	req, _ = http.NewRequest("GET", "/not-found", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res = rr.Result()
	b, err = io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Len(t, b, 0)
}
