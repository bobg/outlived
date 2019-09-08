package site

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"log"
	"net/http"
	"net/url"
	ttemplate "text/template"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleLogin(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return fmt.Errorf("method %s not allowed", req.Method)
	}

	var (
		u        outlived.User
		email    = req.FormValue("email")
		password = req.FormValue("password")
		forgot   = req.FormValue("forgot")
	)
	ctx := req.Context()
	err := aesite.LookupUser(ctx, s.dsClient, email, &u)
	if err != nil {
		// TODO: distinguish "not found" errors from others
		return errors.Wrapf(err, "looking up user %s", email)
	}

	if forgot != "" {
		expSecs, nonce, vtoken, err := aesite.VerificationToken(&u)
		if err != nil {
			return errors.Wrap(err, "generating verification token")
		}

		link, err := url.Parse(fmt.Sprintf("/forgot?e=%d&n=%s&t=%s&u=%s", expSecs, nonce, vtoken, u.Key().Encode()))
		if err != nil {
			return errors.Wrap(err, "constructing forgot-password link")
		}
		link = homeURL.ResolveReference(link)

		dict := map[string]interface{}{"link": link}

		ttmpl, err := ttemplate.New("").Parse(fmailText)
		if err != nil {
			return errors.Wrap(err, "parsing plain-text template")
		}
		textBuf := new(bytes.Buffer)
		err = ttmpl.Execute(textBuf, dict)
		if err != nil {
			return errors.Wrap(err, "executing plain-text template")
		}

		htmpl, err := htemplate.New("").Parse(fmailHTML)
		if err != nil {
			return errors.Wrap(err, "parsing HTML template")
		}
		htmlBuf := new(bytes.Buffer)
		err = htmpl.Execute(htmlBuf, dict)
		if err != nil {
			return errors.Wrap(err, "executing HTML template")
		}

		const subject = "Reset your Outlived password"
		err = s.sender.send(ctx, from, []string{u.Email}, subject, textBuf, htmlBuf)
		if err != nil {
			return errors.Wrap(err, "sending forgot-password mail")
		}

		htmpl, err = htemplate.New("").Parse(postForgotTmpl)
		if err != nil {
			return errors.Wrap(err, "parsing post-forgot page template")
		}
		err = htmpl.Execute(w, nil)
		return errors.Wrap(err, "rendering post-forgot page")
	}

	ok, err := u.CheckPW(password)
	if err != nil {
		return errors.Wrapf(err, "checking password for user %s", email)
	}
	if !ok {
		return codeErr(errors.New("email/password invalid"), http.StatusUnauthorized)
	}

	log.Printf("logging in user %s", email)

	sess, err := aesite.NewSession(ctx, s.dsClient, u.Key())
	if err != nil {
		return errors.Wrapf(err, "creating session for user %s", email)
	}
	sess.SetCookie(w)
	http.Redirect(w, req, "/", http.StatusSeeOther)
	return nil
}

const fmailText = `Follow this link to reset your Outlived password:

  {{ .link }}

This link expires in one hour.
`

const fmailHTML = `
<p>Follow <a href="{{ .link }}">this link</a> to reset your Outlived password:</p>
<p><a href="{{ .link }}">{{ .link }}</a></p>
<p>This link expires in one hour.</p>
`

const postForgotTmpl = `
<html>
  <head>
    <title>
      Outlived
    </title>
  </head>
  <body>
    <h1>Check e-mail</h1>

    <p>
      Reset your Outlived password by following the link
      in the e-mail we just sent you. The link expires in one hour.
    </p>

  </body>
</html>
`
