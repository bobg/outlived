package site

import (
	htemplate "html/template"
	"log"
	"net/http"
	"time"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleSignup(w http.ResponseWriter, req *http.Request) error {
	var (
		ctx      = req.Context()
		email    = req.FormValue("email")
		password = req.FormValue("password")
		bornStr  = req.FormValue("born")
		tzname   = req.FormValue("tzname")
	)
	born, err := outlived.ParseDate(bornStr)
	if err != nil {
		return codeErr(err, http.StatusBadRequest, "parsing birthdate")
	}

	loc, err := time.LoadLocation(tzname)
	if err != nil {
		log.Printf("error loading timezone %s, falling back to UTC: %s", tzname, err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	_, tzoffset := now.Zone()

	u := &outlived.User{
		Born:     born,
		Active:   true,
		TZName:   tzname,
		TZSector: outlived.TZSector(tzoffset),
	}
	err = aesite.NewUser(ctx, s.dsClient, email, password, u)
	if err != nil {
		return errors.Wrap(err, "creating new user")
	}

	err = s.sendVerificationMail(ctx, u, req)
	if err != nil {
		return errors.Wrap(err, "sending verification mail")
	}

	log.Printf("signed up new user %s", u.Email)

	htmpl, err := htemplate.New("").Parse(postSignupTmpl)
	if err != nil {
		return errors.Wrap(err, "parsing post-signup page template")
	}
	err = htmpl.Execute(w, nil)
	return errors.Wrap(err, "rendering post-signup page")
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
      in the e-mail we just sent you.
    </p>

  </body>
</html>
`
