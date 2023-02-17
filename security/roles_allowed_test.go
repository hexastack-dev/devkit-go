package security_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexastack-dev/devkit-go/security"
	"github.com/hexastack-dev/devkit-go/security/principal"
	"github.com/stretchr/testify/assert"
)

func handleHello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	io.WriteString(w, "Hello World")
}

func TestRolesAllowed(t *testing.T) {
	h := security.RolesAllowed("editor", "viewer")(http.HandlerFunc(handleHello))
	req, _ := http.NewRequest("GET", "/hello", nil)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res := rr.Result()
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "response should be Unauthorized")

	u := &principal.User{
		Id: "123abc",
	}
	u.Roles().Add(principal.Role{Name: "accounting"})

	req = req.WithContext(principal.ContextWithUser(req.Context(), u))

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res = rr.Result()
	assert.Equal(t, http.StatusForbidden, res.StatusCode, "response should be Forbidden")

	u.Roles().Add(principal.Role{Name: "editor"})
	req = req.WithContext(principal.ContextWithUser(req.Context(), u))

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res = rr.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode, "response should be OK")
}
