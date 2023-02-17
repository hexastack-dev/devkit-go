package errors

import "net/http"

// ErrorWithHandler implements both error and http.Handler, this handler
// will call ErrorHandlerFunc to handle ServeHTTP, if nil this handler by default
// will only send resonse code without body.
type ErrorWithHandler struct {
	Err              error
	StatusCode       int
	ErrorHandlerFunc func(err error, statusCode int) func(http.ResponseWriter, *http.Request)
}

func (h *ErrorWithHandler) Error() string {
	return h.Err.Error()
}

func (h *ErrorWithHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.ErrorHandlerFunc != nil {
		h.ErrorHandlerFunc(h.Err, h.StatusCode)(w, r)
		return
	}
	w.WriteHeader(h.StatusCode)
}
