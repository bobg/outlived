package site

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleForgot(w http.ResponseWriter, req *http.Request) error {
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

	err = aesite.CheckVerificationToken(&user, expSecs, nonce, vtoken)
	if err != nil {
		return errors.Wrap(err, "checking verification token")
	}

	// Make a secure idempotency token out of the vtoken
	// to make sure it can be used only once for password reset.
	// It works only for the intended user.
	// It is registered/checked when the user submits the /reset form.
	idem, err := user.SecureToken(strings.NewReader(vtoken))
	if err != nil {
		return errors.Wrap(err, "generating idempotency key")
	}

	// TODO: check for/cancel existing session?

	dict := map[string]interface{}{
		"u":    userKeyStr,
		"t":    vtoken,
		"idem": idem,
	}

	tmpl, err := template.New("").Parse(forgotTmpl)
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}

	err = tmpl.Execute(w, dict)
	return errors.Wrap(err, "executing HTML template")
}

func (s *Server) handleResetPW(w http.ResponseWriter, req *http.Request) error {
	if req.Method != "POST" {
		return fmt.Errorf("method %s not allowed", req.Method)
	}

	ctx := req.Context()

	var (
		userKeyStr = req.FormValue("u")
		vtoken     = req.FormValue("t")
		idem       = req.FormValue("idem")
		newPW      = req.FormValue("p")
	)

	userKey, err := datastore.DecodeKey(userKeyStr)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "decoding user key")
	}

	var user outlived.User
	err = s.dsClient.Get(ctx, userKey, &user)
	if err != nil {
		return errors.Wrap(err, "getting user record")
	}

	// Check that idem is both valid for this user and not yet used.

	err = user.CheckToken(strings.NewReader(vtoken), idem)
	if err != nil {
		return errors.Wrap(err, "checking idempotency key")
	}

	err = aesite.Idempotent(ctx, s.dsClient, idem)
	if err != nil {
		return errors.Wrap(err, "checking for token reuse")
	}

	err = aesite.UpdatePW(ctx, s.dsClient, &user, newPW)
	if err != nil {
		return errors.Wrap(err, "storing updated password")
	}

	log.Printf("updated password for %s", user.Email)

	sess, err := aesite.NewSession(ctx, s.dsClient, user.Key())
	if err != nil {
		return errors.Wrapf(err, "creating session for user %s", user.Email)
	}
	sess.SetCookie(w)
	http.Redirect(w, req, "/", http.StatusSeeOther)

	return nil
}

const forgotTmpl = `
<html>
  <head>
    <title>
      Outlived
    </title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
  </head>
  <body>
    <h1>Outlived</h1>

    <form method="POST" action="/s/resetpw">
      <input type="hidden" name="u" value="{{ .u }}"></input>
      <input type="hidden" name="t" value="{{ .t }}"></input>
      <input type="hidden" name="idem" value="{{ .idem }}"></input>
      <label for="newpw">New password</label>
      <input type="password" name="p"></input>
      <button type="submit">Submit</button>
    </form>

  </body>
</html>
`
