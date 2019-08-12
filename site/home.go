package site

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bobg/aesite"

	"github.com/bobg/outlived"
)

func (s *Server) handleHome(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		httpErr(w, 0, "getting session: %s", err)
		return
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
			httpErr(w, 0, "getting user: %s", err)
			return
		}

		alive = today.Since(u.Born)
		figures, err = outlived.FiguresAliveForAtMost(ctx, s.dsClient, alive-1, 20)
		if err != nil {
			httpErr(w, 0, "getting figures: %s", err)
			return
		}
	} else {
		figures, err = outlived.FiguresDiedOn(ctx, s.dsClient, today.M, today.D, 20)
		if err != nil {
			httpErr(w, 0, "getting figures: %s", err)
			return
		}
	}
	tmpl, err := template.ParseFiles(filepath.Join(s.contentDir, "html/home.html.tmpl"))
	if err != nil {
		httpErr(w, 0, "parsing HTML template: %s", err)
		return
	}
	dict := map[string]interface{}{
		"user":     u,
		"figures":  figures,
		"alive":    alive,
		"todaystr": time.Now().Format("Monday, 2 January 2006"),
	}
	err = tmpl.Execute(w, dict)
	if err != nil {
		httpErr(w, 0, "executing HTML template: %s", err)
		return
	}
}
