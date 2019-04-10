package main

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"

	"github.com/bobg/outlived"
)

func (c *controller) handleVerify(w http.ResponseWriter, req *http.Request) {
	userKeyStr := req.FormValue("u")
	userKey, err := datastore.DecodeKey(userKeyStr)
	if err != nil {
		httpErr(w, http.StatusBadRequest, "decoding user key: %s", err)
		return
	}

	var user outlived.User
	err = c.dsClient.Get(req.Context(), userKey, &user)
	if err != nil {
		httpErr(w, 0, "getting user record: %s", err)
		return
	}

	vtoken := req.FormValue("v")
	err = aesite.VerifyUser(req.Context(), c.dsClient, &user, vtoken)
	if err != nil {
		httpErr(w, http.StatusBadRequest, "verifying token: %s", err)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
