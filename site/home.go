package site

import (
	"html/template"
	"net/http"

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
	if sess != nil {
		u = new(outlived.User)
		err = s.dsClient.Get(ctx, sess.UserKey, u)
		if err != nil {
			httpErr(w, 0, "getting user: %s", err)
			return
		}

		today := outlived.Today()
		alive = today.Since(u.Born)
		figures, err = outlived.FiguresAliveForAtMost(ctx, s.dsClient, alive-1, 20)
		if err != nil {
			httpErr(w, 0, "getting figures: %s", err)
			return
		}
	}
	tmpl, err := template.ParseFiles("content/home.html.tmpl") // xxx
	if err != nil {
		httpErr(w, 0, "parsing HTML template: %s", err)
		return
	}
	dict := map[string]interface{}{
		"user":    u,
		"figures": figures,
		"alive":   alive,
	}
	err = tmpl.Execute(w, dict)
	if err != nil {
		httpErr(w, 0, "executing HTML template: %s", err)
		return
	}
}
