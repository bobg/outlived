package site

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

func (s *Server) handleLogout(
	ctx context.Context,
	req struct {
		CSRF string
	},
) error {
	sess := getSess(ctx)
	if sess != nil {
		err := sess.CSRFCheck(req.CSRF)
		if err != nil {
			return errors.Wrap(err, "checking CSRF token")
		}
		err = sess.Cancel(ctx, s.dsClient)
		if err != nil {
			return errors.Wrap(err, "canceling session")
		}
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
	return nil
}
