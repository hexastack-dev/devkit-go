package principal

import "context"

type contextKey string

var principalContextKey contextKey = "principalContextKey"

type contextValue struct {
	user *User
}

func fromContext(ctx context.Context) (*contextValue, bool) {
	v, ok := ctx.Value(principalContextKey).(*contextValue)
	return v, ok
}

// ContextWithUser store user information into context and return wrapped context.
func ContextWithUser(ctx context.Context, u *User) context.Context {
	if v, ok := fromContext(ctx); ok {
		v.user = u
		return ctx
	}
	v := &contextValue{user: u}
	return context.WithValue(ctx, principalContextKey, v)
}

// UserFromContext return stored user in the context, return user, true if found
// or return nil, false otherwise.
func UserFromContext(ctx context.Context) (*User, bool) {
	if v, ok := fromContext(ctx); ok {
		return v.user, true
	}
	return nil, false
}
