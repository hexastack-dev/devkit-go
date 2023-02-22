package securityjwt

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hexastack-dev/devkit-go/errors"
	"github.com/hexastack-dev/devkit-go/log"
	"github.com/hexastack-dev/devkit-go/security/principal"
)

type AudienceVerifier interface {
	VerifyAudience(cmp string, req bool) bool
}

var ignoreExpiration = false // for test only

func New[T jwt.Claims](keyfunc jwt.Keyfunc, claimFactory func() jwt.Claims, mapper func(T) principal.User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hdr, ok := r.Header["Authorization"]
			if !ok {
				// nothing to authenticate, continue
				next.ServeHTTP(w, r)
				return
			}
			// since we're using JWT, it's won't less than 32
			if len(hdr[0]) < 32 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if strings.ToLower(hdr[0][:6]) != "bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenString := hdr[0][7:]
			token, err := jwt.ParseWithClaims(tokenString, claimFactory(), keyfunc)
			if err != nil {
				if !(errors.Is(err, jwt.ErrTokenExpired) && ignoreExpiration) {
					log.WithContext(r.Context()).Warn(fmt.Sprintf("Invalid token: %v", err))
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			claims, ok := token.Claims.(T)
			if !ok {
				log.WithContext(r.Context()).
					Error("Failed to casting Claims into keycloakClaims", errors.New("failed to casting Claims into keycloakClaims", errors.WithTag(1)))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var rawVerifier any = claims
			if verifier, ok := rawVerifier.(AudienceVerifier); ok {
				if ok := verifier.VerifyAudience("oidc-extension", true); !ok {
					err := errors.Tag(jwt.ErrTokenInvalidAudience, 1)
					log.Error("Invalid audience", err)
					w.WriteHeader(http.StatusForbidden)
					return
				}
			} else {
				log.Warn("Underlying Claims is not a AudienceVerifier")
			}

			// u := &principal.User{
			// 	Id:   claims.Subject,
			// 	Name: claims.Name,
			// }
			// addUserRoles(u, claims.RealmAccess, claims.ResourceAccess)

			u := mapper(claims)
			ctx := principal.ContextWithUser(r.Context(), &u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func addUserRoles(usr *principal.User, realmAccess ResourceRoles, resourceAccess ResourceAccess) {
	if len(realmAccess.Roles) > 0 {
		for _, name := range realmAccess.Roles {
			usr.Roles().Add(principal.Role{Name: name})
		}
	}
	if len(resourceAccess) > 0 {
		for _, v := range resourceAccess {
			for _, name := range v.Roles {
				usr.Roles().Add(principal.Role{Name: name})
			}
		}
	}
}

func DefaultJWKSOptions(ctx context.Context) keyfunc.Options {
	return keyfunc.Options{
		Ctx: ctx,
		RefreshErrorHandler: func(err error) {
			log.Error("Failed to refresh JWKS", err)
		},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute * 5,
		RefreshTimeout:    time.Second * 10,
		RefreshUnknownKID: true,
	}
}
