package rwlog

import (
	"bufio"
	"net"
	"net/http"

	"github.com/hexastack-dev/devkit-go/errors"
	"github.com/hexastack-dev/devkit-go/log"
)

var (
	ErrWriterNotPusher   error = errors.New("underlying ResponseWriter is not a Pusher")
	ErrWriterNotHijacker error = errors.New("underlying ResponseWriter is not a Hijacker")
)

// New returns a middleware that wraps ResponseWriter which calls next.ServeHTTP
// any error returned from ResponseWritter.Write trigger onErr.
func New(onError func(error)) func(next http.Handler) http.Handler {
	if onError == nil {
		onError = defaultOnError
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w2 := &loggedResponseWriter{ResponseWriter: w, onErr: onError}
			next.ServeHTTP(w2, r)
		})
	}
}

func defaultOnError(err error) {
	log.Error("Error when writing response", err)
}

type loggedResponseWriter struct {
	http.ResponseWriter
	onErr func(error)
}

func (w *loggedResponseWriter) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	if err != nil {
		w.onErr(errors.Errorf("error when writing response: %w", err))
	}
	return
}

func (w *loggedResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	} else {
		log.Warn("Underlying ResponseWriter is not a Flusher but Flush() is called")
	}
}

func (w *loggedResponseWriter) Hijack() (_ net.Conn, _ *bufio.ReadWriter, err error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.Tag(ErrWriterNotHijacker, 2)
}

func (w *loggedResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return errors.Tag(ErrWriterNotPusher, 2)
}
