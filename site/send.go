package site

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	ttemplate "text/template"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"golang.org/x/text/message"
	"google.golang.org/api/iterator"

	"github.com/bobg/outlived"
)

const subject = "You have outlived!"

func (s *Server) handleSend(w http.ResponseWriter, req *http.Request) error {
	err := s.checkCron(req)
	if err != nil {
		return err
	}

	ctx := req.Context()

	var (
		tzname      = req.FormValue("tzname")
		tzoffsetStr = req.FormValue("tzoffset")
	)

	tzoffset, err := strconv.Atoi(tzoffsetStr)
	if err != nil {
		return errors.Wrapf(err, "parsing tzoffset %s", tzoffsetStr)
	}

	loc := time.FixedZone(tzname, tzoffset)
	now := time.Now().In(loc)
	today := outlived.Date{Y: now.Year(), M: now.Month(), D: now.Day()}

	idemKey := fmt.Sprintf("send-%s-%s", today, tzoffsetStr)
	err = aesite.Idempotent(ctx, s.dsClient, idemKey)
	if err != nil {
		return errors.Wrap(err, "checking idempotency")
	}

	q := datastore.NewQuery("User")
	q = q.Filter("Verified =", true).Filter("Active =", true)
	q = q.Filter("TZSector =", outlived.TZSector(tzoffset))
	q = q.Order("Born.Y").Order("Born.M").Order("Born.D")
	it := s.dsClient.Run(ctx, q)

	var (
		users    []*outlived.User
		lastBorn outlived.Date
	)

	p := message.NewPrinter(message.MatchLanguage("en"))
	numprinter := func(n int) string {
		return p.Sprintf("%v", n)
	}

	redir := func(inp string) string {
		r, _ := rlink(req, inp)
		return r.String()
	}

	wrap := func() error {
		if len(users) == 0 {
			return nil
		}
		defer func() {
			users = nil
		}()

		born := users[0].Born
		since := today.Since(born)
		figures, err := outlived.FiguresAliveFor(ctx, s.dsClient, since-1, 24)
		if err != nil {
			return errors.Wrapf(err, "looking up figures alive for %d days", since-1)
		}
		if len(figures) == 0 {
			log.Printf("%d users west of %s born %d days ago, but no figures alive for %d days", len(users), loc, since, since-1)
			return nil
		}

		dict := map[string]interface{}{
			"born":       born,
			"alivedays":  since,
			"figures":    figures,
			"numprinter": numprinter,
			"redir":      redir,
		}

		ttmpl, err := ttemplate.New("").Parse(mailTextTemplate)
		if err != nil {
			return errors.Wrap(err, "parsing mail text template")
		}
		tbuf := new(bytes.Buffer)
		err = ttmpl.Execute(tbuf, dict)
		if err != nil {
			return errors.Wrap(err, "executing mail text template")
		}
		textPart := tbuf.String()

		htmpl, err := htemplate.New("").Parse(mailHTMLTemplate)
		if err != nil {
			return errors.Wrap(err, "parsing mail HTML template")
		}
		hbuf := new(bytes.Buffer)
		err = htmpl.Execute(hbuf, dict)
		if err != nil {
			return errors.Wrap(err, "executing mail HTML template")
		}
		htmlPart := hbuf.String()

		var to []string
		for _, u := range users {
			to = append(to, u.Email)
		}

		err = s.sender.send(ctx, from, to, subject, strings.NewReader(textPart), strings.NewReader(htmlPart))
		if err != nil {
			return errors.Wrap(err, "sending message")
		}

		log.Printf("sent message to %d users west of %s born %d days ago about %d figure(s) alive for %d days", len(users), loc, since, len(figures), since-1)

		return nil
	}

	for {
		var u outlived.User
		_, err = it.Next(&u)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Wrap(err, "iterating over users")
		}
		if u.Born != lastBorn {
			err = wrap()
			if err != nil {
				return err
			}
		}
		users = append(users, &u)
		lastBorn = u.Born
	}
	return wrap()
}

const mailTextTemplate = `
This is an update from Outlived <https://outlived.net>!

You were born on {{ .born }}, which was {{ call .numprinter .alivedays }} days ago.

You have now outlived:

{{ $redir := .redir }}
{{ range .figures }}
- {{ .Name }}, {{ if .Desc }}{{ .Desc }}, {{ end }}{{ .Born }}—{{ .Died }}. {{ call $redir .Link }}
{{ end }}

To stop receiving these updates, visit https://outlived.net/unsubscribe.
`

const mailHTMLTemplate = `
<p>This is an update from <a href="https://outlived.net/">Outlived</a>!</p>

<p>You were born on {{ .born }}, which was {{ call .numprinter .alivedays }} days ago.</p>

<p>You have now outlived:</p>

<div style="text-align: center;">
  {{ $redir := .redir }}
  {{ range .figures }}
    <div style="display: inline-block; vertical-align: top; margin: 1em 2em; width: 16em;">
      <a href="{{ call $redir .Link }}" target="_blank">
        {{ if .ImgSrc }}
          <img style="max-width: 64px; height: auto;" src="{{ call $redir .ImgSrc }}" alt="{{ .ImgAlt }}"><br>
        {{ end }}
        {{ .Name }}<br>
      </a>
      {{ if .Desc }}
        {{ .Desc }}<br>
      {{ end }}
      {{ .Born }}&mdash;{{ .Died }}
    </div>
  {{ end }}
</div>

<p style="font-size: smaller;">To stop receiving these updates, visit <a href="https://outlived.net/unsubscribe">Outlived</a>.</p>
`
