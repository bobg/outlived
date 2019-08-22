package outlived

import (
	"bytes"
	"context"
	htemplate "html/template"
	ttemplate "text/template"

	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
)

const maxRecipients = 20

func SendMail(ctx context.Context, client *datastore.Client) error {
	var (
		today    = Today()
		lastBorn Date
		users    []*User
	)
	wrap := func() error {
		if len(users) == 0 {
			return nil
		}

		defer func() {
			users = nil
		}()

		birthdate := users[0].Born
		since := today.Since(birthdate)
		figures, err := FiguresAliveFor(ctx, client, since-1, 20)
		if err != nil {
			return errors.Wrapf(err, "looking up figures alive for %d days", since-1)
		}
		if len(figures) == 0 {
			return nil
		}
		for len(users) > 0 {
			var nextUsers []*User
			if len(users) > maxRecipients {
				users, nextUsers = users[:maxRecipients], nextUsers[maxRecipients:]
			}

			dict := map[string]interface{}{
				"birthdate": birthdate,
				"alivedays": since,
				"figures":   figures,
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

			mdict := map[string]interface{}{
				"textpart": textPart,
				"htmlpart": htmlPart,
			}

			mtmpl, err := ttemplate.New("").Parse(mimeTemplate)
			if err != nil {
				return errors.Wrap(err, "parsing MIME template")
			}
			mbuf := new(bytes.Buffer)
			err = mtmpl.Execute(mbuf, mdict)
			if err != nil {
				return errors.Wrap(err, "executing MIME template")
			}

			// xxx construct message and send
			// (need Mime-Version: 1.0 field and Content-Type: multipart/alternative; boundary="x")

			users = nextUsers
		}
		return nil
	}
	err := ForUserByAge(ctx, client, func(ctx context.Context, user *User) error {
		if user.Born != lastBorn {
			err := wrap()
			if err != nil {
				return errors.Wrapf(err, "processing users born on %s", lastBorn)
			}
		}
		users = append(users, user)
		lastBorn = user.Born
		return nil
	})
	if err != nil {
		return err
	}
	return wrap()
}

const mailTextTemplate = `
You were born on {{ .birthdate }}, which was {{ .alivedays }} days ago.

You have now outlived:

{{ range .figures }}
- {{ .Name }}, {{ if .Desc }}{{ .Desc }}, {{ end }}{{ .Born }} - {{ .Died }}. https://en.wikipedia.org{{ .Link }}
{{ end }}
`

const mailHTMLTemplate = `
<p>You were born on {{ .birthdate }}, which was {{ .alivedays }} days ago.</p>

<p>You have now outlived:</p>

<ul>
  {{ range .figures }}
    <li>
      <a href="https://en.wikipedia.org{{ .Link }}" target="_blank">{{ .Name }}</a>,
      {{ if .Desc }}
        {{ .Desc }},
      {{ end }}
      {{ .Born }}&mdash;{{ .Died }}
    </li>
  {{ end }}
</ul>
`

const mimeTemplate = `

--x
Content-Type: text/plain; charset=utf-8

{{ .textpart }}

--x
Content-Type: text/html; charset=utf-8

{{ .htmlpart }}

--x--
`
