package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"strings"

	"github.com/hexastack-dev/devkit-go/log"
	"github.com/hexastack-dev/devkit-go/security"
	"github.com/hexastack-dev/devkit-go/security/principal"
	"github.com/hexastack-dev/devkit-go/server"
)

func handlePublic(w http.ResponseWriter, r *http.Request) {
	msg := "hello %s, this path accessible by everyone"

	w.WriteHeader(http.StatusOK)
	if u, authenticated := principal.UserFromContext(r.Context()); authenticated {
		fmt.Fprintf(w, msg, u.Id)
	} else {
		fmt.Fprintf(w, msg, "anonymous")
	}

}

func handleAuditor(w http.ResponseWriter, r *http.Request) {
	msg := "hello %s, this path accessible by role auditor"

	w.WriteHeader(http.StatusOK)
	if u, authenticated := principal.UserFromContext(r.Context()); authenticated {
		fmt.Fprintf(w, msg, u.Id)
	} else {
		panic("should be authenticated")
	}
}

func handleEditor(w http.ResponseWriter, r *http.Request) {
	msg := "hello %s, this path accessible by role editor"

	w.WriteHeader(http.StatusOK)
	if u, authenticated := principal.UserFromContext(r.Context()); authenticated {
		fmt.Fprintf(w, msg, u.Id)
	} else {
		panic("should be authenticated")
	}
}

func handleAuthenticated(w http.ResponseWriter, r *http.Request) {
	msg := "hello %s, this path accessible by authenticated users"

	w.WriteHeader(http.StatusOK)
	if u, authenticated := principal.UserFromContext(r.Context()); authenticated {
		fmt.Fprintf(w, msg, u.Id)
	} else {
		panic("should be authenticated")
	}
}

func basicAuthGuard() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if authHdr, ok := r.Header["Authorization"]; ok {
				auth := authHdr[0]
				if len(auth) < 9 || auth[:6] != "Basic " {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("only support basic auth"))
					return
				}
				b, err := base64.StdEncoding.DecodeString(auth[6:])
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				creds := strings.Split(string(b), ":")
				if len(creds) != 2 {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("wrong credential format"))
					return
				}
				u, err := getUser(creds[0], creds[1])
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// pass context with authenticated user, so user is available
				// in forth handlers
				ctx := principal.ContextWithUser(r.Context(), u)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

var errWrongCredentials = errors.New("wrong credentials")

func getUser(username, password string) (*principal.User, error) {
	u := new(principal.User)
	switch username {
	case "alice":
		u.Id = "alice"
		u.Roles().Add(principal.Role{Name: "auditor"})
	case "bob":
		u.Id = "bob"
		u.Roles().Add(principal.Role{Name: "editor"})
	default:
		return nil, errWrongCredentials
	}

	if password != "password" {
		return nil, errWrongCredentials
	}

	return u, nil
}

func publicRoute() http.Handler {
	return basicAuthGuard()(http.HandlerFunc(handlePublic))
}

func auditorRoute() http.Handler {
	return basicAuthGuard()(security.RolesAllowed("auditor")(http.HandlerFunc(handleAuditor)))
}

func editorRoute() http.Handler {
	return basicAuthGuard()(security.RolesAllowed("editor")(http.HandlerFunc(handleAuditor)))
}

func authenticatedRoute() http.Handler {
	return basicAuthGuard()(security.Authenticated()(http.HandlerFunc(handleAuthenticated)))
}

func main() {
	logger := log.NewSimpleLogger(stdlog.Writer(), log.InfoLogLevel)
	log.SetLogger(logger)

	mux := http.NewServeMux()
	mux.Handle("/", publicRoute())
	mux.Handle("/auditor", auditorRoute())
	mux.Handle("/editor", editorRoute())
	mux.Handle("/authenticated", authenticatedRoute())
	srv := server.New(mux, nil)

	log.Info("Server started at port 8080")
	srv.ListenAndServe(":8080")
}
