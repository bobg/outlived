package site

import (
	"log"
	"net/http"
	"strconv"

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
