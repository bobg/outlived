package site

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"golang.org/x/text/message"

	"github.com/bobg/outlived"
)

func (s *Server) handleHome(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}

	today := outlived.Today()

	p := message.NewPrinter(message.MatchLanguage("en"))
	numprinter := func(n int) string {
		return p.Sprintf("%v", n)
	}

	dict := map[string]interface{}{
		"todaystr":   time.Now().Format("Monday, 2 January 2006"),
		"numprinter": numprinter,
	}

	if sess != nil {
		u := new(outlived.User)
		err = s.dsClient.Get(ctx, sess.UserKey, u)
		if err != nil {
			return errors.Wrap(err, "getting user")
		}

		alive := today.Since(u.Born)
		figures, err := outlived.FiguresAliveForAtMost(ctx, s.dsClient, alive-1, 20)
		if err != nil {
			return errors.Wrap(err, "getting figures")
		}

		dict["user"] = u
		dict["figures"] = figures
		dict["alive"] = alive
	} else {
		figures, err := outlived.FiguresDiedOn(ctx, s.dsClient, today.M, today.D, 20)
		if err != nil {
			return errors.Wrap(err, "getting figures")
		}

		dict["figures"] = figures
	}

	tmpl, err := template.ParseFiles(filepath.Join(s.contentDir, "html/home.html.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}

	if sess != nil {
		dict["csrf"], err = sess.CSRFToken()
		if err != nil {
			return errors.Wrap(err, "setting CSRF token")
		}
	}

	err = tmpl.Execute(w, dict)
	return errors.Wrap(err, "executing HTML template")
}
