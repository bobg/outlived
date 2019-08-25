package site

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	ttemplate "text/template"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

const from = "Outlived <bobg+outlived@emphatic.com>" // xxx

func (s *Server) handleSignup(w http.ResponseWriter, req *http.Request) error {
	var (
		ctx      = req.Context()
		email    = req.FormValue("email")
		password = req.FormValue("password")
		bornStr  = req.FormValue("born")
	)
	born, err := outlived.ParseDate(bornStr)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "parsing birthdate")
	}
	u := &outlived.User{
		Born:   born,
		Active: true,
	}
	err = aesite.NewUser(ctx, s.dsClient, email, password, u)
	if err != nil {
		return errors.Wrap(err, "creating new user")
	}

	// xxx
	ttmpl, err := ttemplate.ParseFiles(filepath.Join(s.contentDir, "html/verify.mail.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing verification-mail template")
	}

	link, err := url.Parse(fmt.Sprintf("/verify?u=%s&v=%s", u.Key().Encode(), u.VToken))
	if err != nil {
		return errors.Wrap(err, "constructing verification link")
	}
	link = req.URL.ResolveReference(link)

	buf := new(bytes.Buffer)
	err = ttmpl.Execute(buf, map[string]interface{}{"link": link})
	if err != nil {
		return errors.Wrap(err, "executing verification-mail template")
	}
	err = s.sender.send(ctx, from, []string{u.Email}, "Verify your Outlived account", bytes.NewReader(buf.Bytes()), nil)
	if err != nil {
		return errors.Wrap(err, "sending verification mail")
	}

	log.Printf("signed up new user %s", u.Email)

	htmpl, err := htemplate.ParseFiles("content/postsignup.html.tmpl")
	if err != nil {
		return errors.Wrap(err, "parsing post-signup page template")
	}
	err = htmpl.Execute(w, nil)
	return errors.Wrap(err, "rendering post-signup page")
}
