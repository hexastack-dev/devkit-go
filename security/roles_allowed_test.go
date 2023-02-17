package security_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexastack-dev/devkit-go/security"
	"github.com/hexastack-dev/devkit-go/security/principal"
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
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("response should be Unauthorized: %d", rr.Code)
	}

	u := &principal.User{
		Id: "123abc",
	}
	u.Roles().Add(principal.Role{Name: "accounting"})

	req = req.WithContext(principal.ContextWithUser(req.Context(), u))

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res = rr.Result()
	if res.StatusCode != http.StatusForbidden {
		t.Errorf("response should be Forbidden: %d", rr.Code)
	}

	u.Roles().Add(principal.Role{Name: "editor"})
	req = req.WithContext(principal.ContextWithUser(req.Context(), u))

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	res = rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("response should be OK: %d", rr.Code)
	}
}
