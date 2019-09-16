package site

import (
	"bytes"
	"context"
	"fmt"
	htemplate "html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	ttemplate "text/template"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleVerify(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	var (
		expSecsStr = req.FormValue("e")
		nonce      = req.FormValue("n")
		vtoken     = req.FormValue("t")
		userKeyStr = req.FormValue("u")
	)

	expSecs, err := strconv.ParseInt(expSecsStr, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "parsing expSecs parameter %s", expSecsStr)
	}

	userKey, err := datastore.DecodeKey(userKeyStr)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "decoding user key")
	}

	var user outlived.User
	err = s.dsClient.Get(ctx, userKey, &user)
	if err != nil {
		return errors.Wrap(err, "getting user record")
	}

	err = aesite.VerifyUser(ctx, s.dsClient, &user, expSecs, nonce, vtoken)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "verifying token")
	}

	log.Printf("verified user %s", user.Email)

	sess, err := aesite.NewSession(ctx, s.dsClient, user.Key())
	if err != nil {
		return errors.Wrapf(err, "creating session for user %s", user.Email)
	}
	sess.SetCookie(w)
	http.Redirect(w, req, "/", http.StatusSeeOther)

	return nil
}

func (s *Server) handleReverify(w http.ResponseWriter, req *http.Request) error {
	var (
		ctx  = req.Context()
		csrf = req.FormValue("csrf")
	)
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}
	if sess == nil {
		return codeErrType{code: http.StatusUnauthorized}
	}
	err = sess.CSRFCheck(csrf)
	if err != nil {
		return errors.Wrap(err, "checking CSRF token")
	}
	var u outlived.User
	err = sess.GetUser(ctx, s.dsClient, &u)
	if err != nil {
		return errors.Wrapf(err, "getting user for session %d", sess.ID)
	}
	err = s.sendVerificationMail(ctx, &u, req)
	return errors.Wrap(err, "sending verification mail")
}

func (s *Server) sendVerificationMail(ctx context.Context, u *outlived.User, req *http.Request) error {
	expSecs, nonce, vtoken, err := aesite.VerificationToken(u)
	if err != nil {
		return errors.Wrap(err, "generating verification token")
	}

	link, err := url.Parse(fmt.Sprintf("/s/verify?e=%d&n=%s&t=%s&u=%s", expSecs, nonce, vtoken, u.Key().Encode()))
	if err != nil {
		return errors.Wrap(err, "constructing verification link")
	}
	link = homeURL.ResolveReference(link)

	dict := map[string]interface{}{"link": link}

	ttmpl, err := ttemplate.New("").Parse(vmailText)
	if err != nil {
		return errors.Wrap(err, "parsing plain-text template")
	}
	textBuf := new(bytes.Buffer)
	err = ttmpl.Execute(textBuf, dict)
	if err != nil {
		return errors.Wrap(err, "executing plain-text template")
	}

	htmpl, err := htemplate.New("").Parse(vmailHTML)
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}
	htmlBuf := new(bytes.Buffer)
	err = htmpl.Execute(htmlBuf, dict)
	if err != nil {
		return errors.Wrap(err, "executing HTML template")
	}

	const subject = "Verify your Outlived e-mail address"
	err = s.sender.send(ctx, from, []string{u.Email}, subject, textBuf, htmlBuf)
	return errors.Wrap(err, "sending verification mail")
}

const vmailText = `Follow this link to verify your Outlived account:

  {{ .link }}

This link expires in one hour.
`

const vmailHTML = `
<p>Follow <a href="{{ .link }}">this link</a> to verify your Outlived account:</p>
<p><a href="{{ .link }}">{{ .link }}</a></p>
<p>This link expires in one hour.</p>
`
