package site

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleVerify(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	userKeyStr := req.FormValue("u")
	userKey, err := datastore.DecodeKey(userKeyStr)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "decoding user key")
	}

	var user outlived.User
	err = s.dsClient.Get(ctx, userKey, &user)
	if err != nil {
		return errors.Wrap(err, "getting user record")
	}

	vtoken := req.FormValue("v")
	err = aesite.VerifyUser(ctx, s.dsClient, &user, vtoken)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "verifying token")
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)

	return nil
}
