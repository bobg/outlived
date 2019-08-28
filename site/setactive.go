package site

import (
	"net/http"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleSetActive(w http.ResponseWriter, req *http.Request) error {
	var (
		ctx       = req.Context()
		active    = req.FormValue("active") == "true"
		csrfToken = req.FormValue("csrf")
	)
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}
	if sess == nil {
		return codeErrType{code: http.StatusUnauthorized}
	}
	err = sess.CSRFCheck(csrfToken)
	if err != nil {
		return errors.Wrap(err, "checking CSRF token")
	}
	var u outlived.User
	err = sess.GetUser(ctx, s.dsClient, &u)
	if err != nil {
		return errors.Wrapf(err, "getting user for session %d", sess.ID)
	}
	u.Active = active
	_, err = s.dsClient.Put(ctx, u.Key(), &u)
	return errors.Wrapf(err, "updating user %s", u.Email)
}
