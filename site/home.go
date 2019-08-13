package site

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

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
	var (
		u       *outlived.User
		figures []*outlived.Figure
		alive   int
	)
	today := outlived.Today()
	if sess != nil {
		u = new(outlived.User)
		err = s.dsClient.Get(ctx, sess.UserKey, u)
		if err != nil {
			return errors.Wrap(err, "getting user")
		}

		alive = today.Since(u.Born)
		figures, err = outlived.FiguresAliveForAtMost(ctx, s.dsClient, alive-1, 20)
		if err != nil {
			return errors.Wrap(err, "getting figures")
		}
	} else {
		figures, err = outlived.FiguresDiedOn(ctx, s.dsClient, today.M, today.D, 20)
		if err != nil {
			return errors.Wrap(err, "getting figures")
		}
	}
	tmpl, err := template.ParseFiles(filepath.Join(s.contentDir, "html/home.html.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}
	dict := map[string]interface{}{
		"user":     u,
		"figures":  figures,
		"alive":    alive,
		"todaystr": time.Now().Format("Monday, 2 January 2006"),
	}
	err = tmpl.Execute(w, dict)
	return errors.Wrap(err, "executing HTML template")
}
