package security

import (
	"net/http"

	"github.com/hexastack-dev/devkit-go/security/principal"
)

// Authenticated is http middleware which filter out unauthenticated incoming request.
// Notes: This handler doesn't works alone, you need to use other security component
// to handle authentication and save the user info into the sdk principal package,
// such as: security-oidc.
func Authenticated() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := principal.UserFromContext(r.Context())
			if !ok {
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
