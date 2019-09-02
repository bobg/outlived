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

		dict["user"] = u
		dict["alive"] = alive

		dict["csrf"], err = sess.CSRFToken()
		if err != nil {
			return errors.Wrap(err, "setting CSRF token")
		}
	}

	tmpl, err := template.ParseFiles(filepath.Join(s.contentDir, "html/home.html.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}

	err = tmpl.Execute(w, dict)
	return errors.Wrap(err, "executing HTML template")
}
