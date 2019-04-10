package main

import (
	"bytes"
	htemplate "html/template"
	"net/http"
	"net/mail"
	ttemplate "text/template"

	"github.com/bobg/aesite"

	"github.com/bobg/outlived"
)

func (c *controller) handleSignup(w http.ResponseWriter, req *http.Request) {
	var (
		email    = req.FormValue("email")
		password = req.FormValue("password")
		bornStr  = req.FormValue("born")
	)
	born, err := outlived.ParseDate(bornStr)
	if err != nil {
		httpErr(w, http.StatusBadRequest, "parsing birthdate: %s", err)
		return
	}
	u := &outlived.User{
		Born: born,
	}
	err = aesite.NewUser(req.Context(), c.dsClient, email, password, u)
	if err != nil {
		httpErr(w, 0, "creating new user: %s", err)
		return
	}

	ttmpl, err := ttemplate.ParseFiles("content/verify.mail.tmpl")
	if err != nil {
		httpErr(w, 0, "parsing verification-mail template: %s", err)
		return
	}
	buf := new(bytes.Buffer)
	err = ttmpl.Execute(buf, map[string]interface{}{"link": link})
	if err != nil {
		httpErr(w, 0, "executing verification-mail template: %s", err)
		return
	}
	err = c.sender.send(req.Context(), []mail.Address{addr}, "Verify your Outlived account", bytes.NewReader(buf.Bytes()))
	if err != nil {
		httpErr(w, 0, "sending verification mail: %s", err)
		return
	}

	htmpl, err := htemplate.ParseFiles("content/postsignup.html.tmpl")
	if err != nil {
		httpErr(w, 0, "parsing post-signup page template: %s", err)
		return
	}
	err = htmpl.Execute(w, nil)
	if err != nil {
		httpErr(w, 0, "rendering post-signup page: %s", err)
		return
	}
}
