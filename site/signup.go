package site

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bobg/aesite"
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
) error {
	born, err := outlived.ParseDate(req.BornStr)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "parsing birthdate")
	}

	loc, err := time.LoadLocation(req.TZName)
	if err != nil {
		log.Printf("error loading timezone %s, falling back to UTC: %s", req.TZName, err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	_, tzoffset := now.Zone()

	u := &outlived.User{
		Born:     born,
		Active:   true,
		TZName:   loc.String(),
		TZSector: outlived.TZSector(tzoffset),
	}
	err = aesite.NewUser(ctx, s.dsClient, req.Email, req.Password, u)
	if err != nil {
		return errors.Wrap(err, "creating new user")
	}

	err = s.sendVerificationMail(ctx, u)
	if err != nil {
		return errors.Wrap(err, "sending verification mail")
	}

	log.Printf("signed up new user %s", u.Email)

	// xxx

	return nil
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
