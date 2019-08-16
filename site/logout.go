package site

import (
	"net/http"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"
)

func (s *Server) handleLogout(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}
	if sess != nil {
		err = sess.Cancel(ctx, s.dsClient)
		if err != nil {
			return errors.Wrap(err, "canceling session")
		}
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
	return nil
}
