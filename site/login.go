package site

import (
	"bytes"
	"encoding/json"
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

	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Forgot   bool   `json:"forgot"`
		TZName   string `json:"tzname"`
	}
	err := json.NewDecoder(req.Body).Decode(&loginReq)
	if err != nil {
		return errors.Wrap(err, "decoding request")
	}

	var (
		now   = tzNow(loginReq.TZName)
		today = outlived.TimeDate(now)
	)

	var u outlived.User

	ctx := req.Context()
	err = aesite.LookupUser(ctx, s.dsClient, loginReq.Email, &u)
	if err != nil {
		// TODO: distinguish "not found" errors from others
		return errors.Wrapf(err, "looking up user %s", loginReq.Email)
	}

	if loginReq.Forgot {
		expSecs, nonce, vtoken, err := aesite.VerificationToken(&u)
		if err != nil {
			return errors.Wrap(err, "generating verification token")
		}

		link, err := url.Parse(fmt.Sprintf("/s/forgot?e=%d&n=%s&t=%s&u=%s", expSecs, nonce, vtoken, u.Key().Encode()))
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

	ok, err := u.CheckPW(loginReq.Password)
	if err != nil {
		return errors.Wrapf(err, "checking password for user %s", loginReq.Email)
	}
	if !ok {
		return codeErr(errors.New("email/password invalid"), http.StatusUnauthorized)
	}

	log.Printf("logging in user %s", loginReq.Email)

	sess, err := aesite.NewSession(ctx, s.dsClient, u.Key())
	if err != nil {
		return errors.Wrapf(err, "creating session for user %s", loginReq.Email)
	}
	_, d, err := s.getUserData2(ctx, sess, &u, today)
	if err != nil {
		return errors.Wrap(err, "getting user data")
	}

	sess.SetCookie(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(d)

	return errors.Wrap(err, "encoding response")
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
