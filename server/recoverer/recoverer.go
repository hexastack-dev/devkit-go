package recoverer

import (
	"context"
	"net/http"

	"github.com/hexastack-dev/devkit-go/errors"
	"github.com/hexastack-dev/devkit-go/log"
)

type contextKey string

var recovererContextKey = contextKey("recoverer")

func writeServerError(w http.ResponseWriter, r *http.Request) {
	if err, ok := GetErrFromContext(r.Context()); ok {
		log.Error("Panic occured", err)
	} else {
		log.Error("Panic occured", errors.New("unknown panic", errors.WithTag(1)))
	}
	w.WriteHeader(http.StatusInternalServerError)
}

// New returns a handler that will calls next.ServeHTTP and
// will recover from panic and calls errorHandler.ServeHTTP to handle the error.
// If error that causing panic is type of http.ErrAbortHandler, we will recover
// but will not inform any error (we ignore it). If panicHandler is nil, the
// default error handler will be used.
// Error from panic can be accessed using GetErrFromContext.
func New(panicHandler http.Handler) func(next http.Handler) http.Handler {
	if panicHandler == nil {
		panicHandler = http.HandlerFunc(writeServerError)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						// http.ErrAbortHandler should not be logged
						return
					}

					var err error
					switch e := rvr.(type) {
					case string:
						err = errors.New(e)
					case error:
						err = e
					default:
						err = errors.Errorf("unknown panic: %#v", rvr)
					}

					ctx := context.WithValue(r.Context(), recovererContextKey, err)
					r = r.WithContext(ctx)

					panicHandler.ServeHTTP(w, r)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// GetErrFromContext return captured error when panic occurs.
func GetErrFromContext(ctx context.Context) (error, bool) {
	err, ok := ctx.Value(recovererContextKey).(error)
	return err, ok
}
