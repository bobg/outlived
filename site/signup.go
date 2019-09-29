package site

import (
	"context"
	"log"
	"net/http"

	"github.com/bobg/aesite"
	"github.com/bobg/hj"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleSignup(
	ctx context.Context,
	req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		BornStr  string `json:"born"`
		TZName   string `json:"tzname"`
	},
) (*userData, error) {
	born, err := outlived.ParseDate(req.BornStr)
	if err != nil {
		return nil, codeErr(err, http.StatusBadRequest, "parsing birthdate")
	}

	var (
		now   = tzNow(req.TZName)
		today = outlived.TimeDate(now)
		loc   = now.Location()
	)

	_, tzoffset := now.Zone()

	u := &outlived.User{
		Born:     born,
		Active:   true,
		TZName:   loc.String(),
		TZSector: outlived.TZSector(tzoffset),
	}
	err = aesite.NewUser(ctx, s.dsClient, req.Email, req.Password, u)
	if err != nil {
		return nil, errors.Wrap(err, "creating new user")
	}

	err = s.sendVerificationMail(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "sending verification mail")
	}

	log.Printf("signed up new user %s", u.Email)

	sess, err := aesite.NewSession(ctx, s.dsClient, u.Key())
	if err != nil {
		return nil, errors.Wrapf(err, "creating session for user %s", req.Email)
	}
	_, d, err := s.getUserData2(ctx, sess, u, today)
	if err != nil {
		return nil, errors.Wrap(err, "getting user data")
	}

	w := hj.Response(ctx)
	sess.SetCookie(w)

	return d, nil
}

const postSignupTmpl = `
<html>
  <head>
    <title>
      Outlived
    </title>
  </head>
  <body>
    <h1>Check e-mail</h1>

    <p>
      Activate your Outlived account by following the verification link
      in the e-mail we just sent you. The link expires in one hour.
    </p>

  </body>
</html>
`
