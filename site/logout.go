package site

import (
	"fmt"
	"net/http"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"
)

func (s *Server) handleLogout(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return fmt.Errorf("method %s not allowed", req.Method)
	}

	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}
	if sess != nil {
		csrfToken := req.FormValue("csrf")
		err = sess.CSRFCheck(csrfToken)
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
