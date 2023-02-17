package security

import (
	"net/http"

	"github.com/hexastack-dev/devkit-go/security/principal"
)

// RolesAllowed is http middleware which filter incoming request based on speficied roles
// againts user roles.
// Notes: This handler doesn't works alone, you need to use other security component
// to handle authentication and save the user info into the sdk principal package,
// such as: security-oidc.
func RolesAllowed(r0 string, rn ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := []string{r0}
			if len(rn) > 0 {
				roles = append(roles, rn...)
			}
			if u, ok := principal.UserFromContext(r.Context()); ok {
				var allowed bool
				uroles := u.Roles()
				for _, r := range roles {
					if _, ok := uroles[r]; ok {
						allowed = true
						break
					}
				}

				if allowed {
					next.ServeHTTP(w, r)
				} else {
					// user is authenticated but do not have sufficient roles
					w.WriteHeader(http.StatusForbidden)
				}
			} else {
				// user is not authenticated thus unauthorized
				w.WriteHeader(http.StatusUnauthorized)
			}
		})
	}
}
