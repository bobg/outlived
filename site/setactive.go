package site

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleSetActive(
	ctx context.Context,
	req struct {
		CSRF   string
		Active bool
	},
) error {
	sess := getSess(ctx)
	if sess == nil {
		return codeErrType{code: http.StatusUnauthorized}
	}
	err := sess.CSRFCheck(req.CSRF)
	if err != nil {
		return errors.Wrap(err, "checking CSRF token")
	}
	var u outlived.User
	err = sess.GetUser(ctx, s.dsClient, &u)
	if err != nil {
		return errors.Wrapf(err, "getting user for session %d", sess.ID)
	}
	if u.Active == req.Active {
		return nil
	}
	u.Active = req.Active
	_, err = s.dsClient.Put(ctx, u.Key(), &u)
	return errors.Wrapf(err, "updating user %s", u.Email)
}
