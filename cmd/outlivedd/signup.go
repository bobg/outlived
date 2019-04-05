package main

import (
	"net/http"

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
		httpErr(w, 0, http.StatusBadRequest, "parsing birthdate: %s", err)
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
	// xxx send verification mail
}
