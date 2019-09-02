package site

import (
	"bytes"
	"fmt"
	htemplate "html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	ttemplate "text/template"
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

	// xxx
	ttmpl, err := ttemplate.ParseFiles(filepath.Join(s.contentDir, "html/verify.mail.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing verification-mail template")
	}

	expSecs, nonce, vtoken, err := aesite.VerificationToken(u)
	if err != nil {
		return errors.Wrap(err, "generating verification token")
	}

	link, err := url.Parse(fmt.Sprintf("/verify?e=%d&n=%s&t=%s&u=%s", expSecs, nonce, vtoken, u.Key().Encode()))
	if err != nil {
		return errors.Wrap(err, "constructing verification link")
	}
	link = requrl(req, link)

	buf := new(bytes.Buffer)
	err = ttmpl.Execute(buf, map[string]interface{}{"link": link})
	if err != nil {
		return errors.Wrap(err, "executing verification-mail template")
	}

	textPart := buf.String()

	const subject = "Verify your Outlived e-mail address"

	err = s.sender.send(ctx, from, []string{u.Email}, subject, strings.NewReader(textPart), nil)
	if err != nil {
		return errors.Wrap(err, "sending verification mail")
	}

	log.Printf("signed up new user %s", u.Email)

	htmpl, err := htemplate.ParseFiles(filepath.Join(s.contentDir, "html/postsignup.html.tmpl"))
	if err != nil {
		return errors.Wrap(err, "parsing post-signup page template")
	}
	err = htmpl.Execute(w, nil)
	return errors.Wrap(err, "rendering post-signup page")
}
