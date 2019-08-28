package site

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleLogin(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return fmt.Errorf("method %s not allowed", req.Method)
	}

	var (
		u        outlived.User
		email    = req.FormValue("email")
		password = req.FormValue("password")
	)
	ctx := req.Context()
	err := aesite.LookupUser(ctx, s.dsClient, email, &u)
	if err != nil {
		// TODO: distinguish "not found" errors from others
		return errors.Wrapf(err, "looking up user %s", email)
	}
	ok, err := u.CheckPW(password)
	if err != nil {
		return errors.Wrapf(err, "checking password for user %s", email)
	}
	if !ok {
		return codeErr(errors.New("email/password invalid"), http.StatusUnauthorized)
	}

	log.Printf("logging in user %s", email)

	sess, err := aesite.NewSession(ctx, s.dsClient, u.Key())
	if err != nil {
		return errors.Wrapf(err, "creating session for user %s", email)
	}
	sess.SetCookie(w)
	http.Redirect(w, req, "/", http.StatusSeeOther)
	return nil
}
