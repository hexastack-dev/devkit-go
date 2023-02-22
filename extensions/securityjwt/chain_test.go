package securityjwt

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/hexastack-dev/devkit-go/security/principal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	tokenString     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzcwNTM2NDEsImlhdCI6MTY3NzA1MzM0MSwiYXV0aF90aW1lIjoxNjc3MDUwNDMxLCJqdGkiOiI0Y2IyNGIxMS1jNTkzLTQxNzItYmJhMi05YmI4YzkyOTc0N2QiLCJpc3MiOiJodHRwczovL3Nzby5kZXYuaGV4YXN0YWNrLmxvY2FsOjg0NDMvcmVhbG1zL3NkayIsImF1ZCI6WyJvaWRjLWV4dGVuc2lvbiIsImFjY291bnQiXSwic3ViIjoiNGEzZGNkYzQtMmUxNC00ZDdlLTljM2YtZGI4MWEyNGRmOTI0IiwidHlwIjoiQmVhcmVyIiwiYXpwIjoiZnJvbnRlbmQtc2VydmljZSIsInNlc3Npb25fc3RhdGUiOiIzMDhmNGJiOS0zOGUwLTQ0YmYtOWNjZi03ZWI0OGY0NWU0NTEiLCJhY3IiOiIwIiwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbImVkaXRvciIsInZpZXdlciIsImRlZmF1bHQtcm9sZXMtc2RrIiwib2ZmbGluZV9hY2Nlc3MiLCJ1bWFfYXV0aG9yaXphdGlvbiJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im9pZGMtZXh0ZW5zaW9uIjp7InJvbGVzIjpbImNsaWVudC1vaWRjLXJvbGUiXX0sImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInNjb3BlIjoib3BlbmlkIGVtYWlsIHByb2ZpbGUiLCJzaWQiOiIzMDhmNGJiOS0zOGUwLTQ0YmYtOWNjZi03ZWI0OGY0NWU0NTEiLCJlbWFpbF92ZXJpZmllZCI6ZmFsc2UsIm5hbWUiOiJBbGljZSIsInByZWZlcnJlZF91c2VybmFtZSI6ImFsaWNlIiwiZ2l2ZW5fbmFtZSI6IkFsaWNlIiwiZmFtaWx5X25hbWUiOiIiLCJlbWFpbCI6ImFsaWNlQGV4YW1wbGUuY29tIn0.o2P0l-pp33n1ZRDhrN1XOOdnkwwwkxANo9peGV0nOiE"
	expectedPayload = "4a3dcdc4-2e14-4d7e-9c3f-db81a24df924\tAlice\t"
	secretKey       = []byte("1rip9UTE1hO6ay9zac#&5Riphl3iJESp")

	wrongAudTokenString = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NzcwNTM2NDEsImlhdCI6MTY3NzA1MzM0MSwiYXV0aF90aW1lIjoxNjc3MDUwNDMxLCJqdGkiOiI0Y2IyNGIxMS1jNTkzLTQxNzItYmJhMi05YmI4YzkyOTc0N2QiLCJpc3MiOiJodHRwczovL3Nzby5kZXYuaGV4YXN0YWNrLmxvY2FsOjg0NDMvcmVhbG1zL3NkayIsImF1ZCI6WyJvdGhlcnMiLCJhY2NvdW50Il0sInN1YiI6IjRhM2RjZGM0LTJlMTQtNGQ3ZS05YzNmLWRiODFhMjRkZjkyNCIsInR5cCI6IkJlYXJlciIsImF6cCI6ImZyb250ZW5kLXNlcnZpY2UiLCJzZXNzaW9uX3N0YXRlIjoiMzA4ZjRiYjktMzhlMC00NGJmLTljY2YtN2ViNDhmNDVlNDUxIiwiYWNyIjoiMCIsInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJlZGl0b3IiLCJ2aWV3ZXIiLCJkZWZhdWx0LXJvbGVzLXNkayIsIm9mZmxpbmVfYWNjZXNzIiwidW1hX2F1dGhvcml6YXRpb24iXX0sInJlc291cmNlX2FjY2VzcyI6eyJvaWRjLWV4dGVuc2lvbiI6eyJyb2xlcyI6WyJjbGllbnQtb2lkYy1yb2xlIl19LCJhY2NvdW50Ijp7InJvbGVzIjpbIm1hbmFnZS1hY2NvdW50IiwibWFuYWdlLWFjY291bnQtbGlua3MiLCJ2aWV3LXByb2ZpbGUiXX19LCJzY29wZSI6Im9wZW5pZCBlbWFpbCBwcm9maWxlIiwic2lkIjoiMzA4ZjRiYjktMzhlMC00NGJmLTljY2YtN2ViNDhmNDVlNDUxIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJuYW1lIjoiQWxpY2UiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJhbGljZSIsImdpdmVuX25hbWUiOiJBbGljZSIsImZhbWlseV9uYW1lIjoiIiwiZW1haWwiOiJhbGljZUBleGFtcGxlLmNvbSJ9.4G6T9giADywrqi38gI6l7r-FxXi0qFykN8IC6sBpTg0"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	if u, ok := principal.UserFromContext(r.Context()); ok {
		var roles []string
		for _, role := range u.Roles().Values() {
			roles = append(roles, role.Name)
		}
		fmt.Fprintf(w, "%s\t%s\t%s", u.Id, u.Name, strings.Join(roles, ","))
		return
	}
}

func keyFunc(_ *jwt.Token) (interface{}, error) {
	return secretKey, nil
}

func TestNew(t *testing.T) {
	ignoreExpiration = true
	h := New(keyFunc, func() jwt.Claims { return &KeycloakClaims{} }, KeycloakUserMapper)(http.HandlerFunc(handleHello))

	t.Run("Test unauthenticated", testUnauthenticated(h))
	t.Run("Test empty authorization", testEmptyAuthorization(h))
	t.Run("Test wrong auth scheme", testAuthScheme(h))
	t.Run("Test audience", testAudience(h))
	t.Run("Test authenticated", testAuthenticated(h))
}

func TestNewKeycloak(t *testing.T) {
	ignoreExpiration = true
	h := NewKeycloak(keyFunc)(http.HandlerFunc(handleHello))

	t.Run("Test unauthenticated", testUnauthenticated(h))
	t.Run("Test empty authorization", testEmptyAuthorization(h))
	t.Run("Test wrong auth scheme", testAuthScheme(h))
	t.Run("Test audience", testAudience(h))
	t.Run("Test authenticated", testAuthenticated(h))
}

func TestIntegrationKeycloak(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jwks, err := keyfunc.Get(os.Getenv("JWKS_URL"), DefaultJWKSOptions(ctx))
	require.NoError(t, err)
	h := NewKeycloak(jwks.Keyfunc)(http.HandlerFunc(handleHello))

	tokenString = os.Getenv("TEST_TOKEN")

	t.Run("Test unauthenticated", testUnauthenticated(h))
	t.Run("Test empty authorization", testEmptyAuthorization(h))
	t.Run("Test wrong auth scheme", testAuthScheme(h))
	t.Run("Test authenticated", testAuthenticated(h))
}

func testUnauthenticated(h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		h.ServeHTTP(rr, req)
		res := rr.Result()
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode, "status should be OK")
		assert.Empty(t, b)
	}
}

func testEmptyAuthorization(h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("authorization", "")
		h.ServeHTTP(rr, req)
		res := rr.Result()
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "status should be Unauthorized")
		assert.Empty(t, b)
	}
}

func testAuthScheme(h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("authorization", fmt.Sprintf("Basic %s", tokenString))
		h.ServeHTTP(rr, req)
		res := rr.Result()
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "status should be Unauthorized")
		assert.Empty(t, b)
	}
}

func testAudience(h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", wrongAudTokenString))
		h.ServeHTTP(rr, req)
		res := rr.Result()
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, res.StatusCode, "status should be Forbidden")
		assert.Empty(t, b)
	}
}

func testAuthenticated(h http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", tokenString))
		h.ServeHTTP(rr, req)
		res := rr.Result()
		b, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode, "status should be OK")
		swPayload := string(b)[:len(expectedPayload)]
		assert.Equalf(t, expectedPayload, swPayload, "unexpected payload: %s", swPayload)

		rolesString := string(b)[43:]
		assert.Truef(t, strings.Contains(rolesString, "editor"), "should contains editor: %s", rolesString)
		assert.Truef(t, strings.Contains(rolesString, "viewer"), "should contains viewer: %s", rolesString)
		assert.Truef(t, strings.Contains(rolesString, "client-oidc-role"), "should contains client-oidc-role: %s", rolesString)
		assert.Truef(t, strings.Contains(rolesString, "view-profile"), "view-profile: %s", rolesString)
	}
}
