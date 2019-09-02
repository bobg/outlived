package site

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleHome(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}

	var dict map[string]interface{}

	if sess != nil {
		u := new(outlived.User)
		err = s.dsClient.Get(ctx, sess.UserKey, u)
		if err != nil {
			return errors.Wrap(err, "getting user")
		}

		csrf, err := sess.CSRFToken()
		if err != nil {
			return errors.Wrap(err, "setting CSRF token")
		}
		dict = map[string]interface{}{
			"user": u,
			"csrf": csrf,
		}
	}

	tmpl, err := template.ParseFiles(filepath.Join(s.contentDir, "html/home.html.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}

	err = tmpl.Execute(w, dict)
	return errors.Wrap(err, "executing HTML template")
}
