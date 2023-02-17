package principal

// User represent current user information, usually information contained in
// OIDC ID Token, see https://auth0.com/docs/secure/tokens/id-tokens/id-token-structure
// for available claims. However, since we are using keycloak as our default IdP we should
// adapt to keycloak about which fields are optional. As of this writing, only
// Id (sub) and username (preferred_username) is mandatory in keycloak.
type User struct {
	Id       string `json:"id"`
	Username string `json:"preferred_username"`

	Email         string `json:"email,omitempty"`
	EmailVerified string `json:"email_verified,omitempty"`
	Name          string `json:"name,omitempty"`

	roles Roles
}

// Roles return list of roles (as map) asociated with user, if roles nil then return empty map.
func (u *User) Roles() Roles {
	if u.roles == nil {
		u.roles = make(Roles)
	}
	return u.roles
}
