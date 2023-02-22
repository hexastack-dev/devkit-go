package securityjwt

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hexastack-dev/devkit-go/security/principal"
)

type ResourceRoles struct {
	Roles []string `json:"roles"`
}
type ResourceAccess map[string]ResourceRoles

type KeycloakClaims struct {
	jwt.RegisteredClaims
	Username       string         `json:"preferred_username"`
	Name           string         `json:"given_name"`
	LastName       string         `json:"family_name"`
	Email          string         `json:"email"`
	EmailVerified  bool           `json:"email_verified"`
	Scope          string         `json:"scope"`
	RealmAccess    ResourceRoles  `json:"realm_access"`
	ResourceAccess ResourceAccess `json:"resource_access"`
}

func KeycloakClaimsFactory() jwt.Claims {
	return &KeycloakClaims{}
}

func KeycloakUserMapper(claims *KeycloakClaims) principal.User {
	u := principal.User{
		Id:   claims.Subject,
		Name: claims.Name,
	}
	addUserRoles(&u, claims.RealmAccess, claims.ResourceAccess)
	return u
}

func NewKeycloak(keyfunc jwt.Keyfunc) func(http.Handler) http.Handler {
	return New(keyfunc, KeycloakClaimsFactory, KeycloakUserMapper)
}
