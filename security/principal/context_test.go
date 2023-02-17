package principal_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/hexastack-dev/devkit-go/security/principal"
)

func TestContext(t *testing.T) {
	ctx := context.TODO()
	if v, ok := principal.UserFromContext(ctx); ok {
		t.Errorf("user should not be found in the context: %v", v)
	}

	u := &principal.User{
		Id: "123abc",
	}
	u.Roles().Add(principal.Role{Name: "ROLE_ABC"})
	ctx = principal.ContextWithUser(ctx, u)

	if u, ok := principal.UserFromContext(ctx); !ok {
		t.Error("user should be found in the context")
	} else {
		if u.Roles() == nil {
			t.Errorf("roles should not be nil: %+v", u.Roles())
		} else {
			roles := u.Roles()
			if len(roles) != 1 {
				t.Errorf("roles length should be 1: %d", len(roles))
			}
			role, ok := roles.Get("ROLE_ABC")
			if !ok {
				t.Error("ROLE_ABC should be found")
			}
			if role.Name != "ROLE_ABC" {
				t.Errorf("role.Name should be equals ROLE_ABC: %s", role.Name)
			}
			rolev := u.Roles().Values()
			if len(rolev) != 1 {
				t.Errorf("roleValues length should be 1: %d", len(rolev))
			}
			if rolev[0].Name != "ROLE_ABC" {
				t.Errorf("role[0].Name should equals ROLE_ABC: %s", rolev[0].Name)
			}
		}
	}

	u, ok := principal.UserFromContext(ctx)
	if !ok {
		t.Fatal("user should be found in the context")
	}
	u.Roles().Add(principal.Role{Name: "ROLE_XYZ"})

	if u, ok := principal.UserFromContext(ctx); !ok {
		t.Error("user should be found in the context")
	} else {
		if u.Roles() == nil {
			t.Errorf("roles should not be nil: %+v", u.Roles())
		} else {
			roles := u.Roles()
			if len(roles) != 2 {
				t.Errorf("roles length should be 2: %d", len(roles))
			}

			expected := make(principal.Roles)
			expected.Add(principal.Role{Name: "ROLE_ABC"})
			expected.Add(principal.Role{Name: "ROLE_XYZ"})
			if !reflect.DeepEqual(expected, roles) {
				t.Errorf("roles should be deep equals expected: %v", roles)
			}

			role, ok := roles.Get("ROLE_ABC")
			if !ok {
				t.Error("ROLE_ABC should be found")
			}
			if role.Name != "ROLE_ABC" {
				t.Errorf("role.Name should be equals ROLE_ABC: %s", role.Name)
			}
			role, ok = roles.Get("ROLE_XYZ")
			if !ok {
				t.Error("ROLE_XYZ should be found")
			}
			if role.Name != "ROLE_XYZ" {
				t.Errorf("role.Name should be equals ROLE_XYZ: %s", role.Name)
			}

			rolev := u.Roles().Values()

			if len(rolev) != 2 {
				t.Errorf("roleValues length should be 2: %d", len(rolev))
			}

			// .Values() return unordered slice, we need to make it static for test
			expected1 := make([]string, 2)
			for _, v := range rolev {
				if v.Name == "ROLE_ABC" {
					expected1[0] = v.Name
				} else if v.Name == "ROLE_XYZ" {
					expected1[1] = v.Name
				}
			}
			if expected1[0] != "ROLE_ABC" {
				t.Errorf("role[0].Name should equals ROLE_ABC: %+v", expected1[0])
			}
			if expected1[1] != "ROLE_XYZ" {
				t.Errorf("role[1].Name should equals ROLE_XYZ: %+v", expected1[1])
			}

			role, ok = roles.Get("NoOp")
			if ok {
				t.Error("noop role should not be found")
			}

			noop := principal.Role{}
			if noop != role {
				t.Errorf("non-exitance role should be equals empty: %+v", role)
			}
		}
	}
}
