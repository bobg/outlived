package site

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"golang.org/x/text/message"

	"github.com/bobg/outlived"
)

type (
	figuresResp struct {
		Today   string
		Email   string
		Born    string
		Alive   string
		Figures []respFigure
	}
	respFigure struct {
		Link      string
		ImgSrc    string
		ImgAlt    string
		Name      string
		Desc      string
		Born      string
		Died      string
		DaysAlive string
		YDAlive   string
	}
)

func (s *Server) handleFigures(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	var (
		tzname    = req.FormValue("tzname")
		csrfToken = req.FormValue("csrf")
	)

	loc, err := time.LoadLocation(tzname)
	if err != nil {
		log.Printf("error loading timezone %s, falling back to UTC: %s", tzname, err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := outlived.TimeDate(now)

	resp := figuresResp{Today: now.Format("Monday, 2 January 2006")}

	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}

	p := message.NewPrinter(message.MatchLanguage("en"))
	numprinter := func(n int) string {
		return p.Sprintf("%v", n)
	}

	var figures []*outlived.Figure
	if sess != nil {
		err = sess.CSRFCheck(csrfToken)
		if err != nil {
			return errors.Wrap(err, "checking CSRF token")
		}

		u := new(outlived.User)
		err = s.dsClient.Get(ctx, sess.UserKey, u)
		if err != nil {
			return errors.Wrap(err, "getting user")
		}

		alive := today.Since(u.Born)

		resp.Email = u.Email
		resp.Born = u.Born.String()
		resp.Alive = numprinter(alive)

		figures, err = outlived.FiguresAliveForAtMost(ctx, s.dsClient, alive-1, 20)
		if err != nil {
			return errors.Wrap(err, "getting figures")
		}
	} else {
		figures, err = outlived.FiguresDiedOn(ctx, s.dsClient, today.M, today.D, 20)
		if err != nil {
			return errors.Wrap(err, "getting figures")
		}
	}

	for _, figure := range figures {
		y, d := figure.Died.YDSince(figure.Born)
		var ydAlive string

		switch {
		case y == 0 && d == 0:
			ydAlive = "0 days"

		case y == 0 && d == 1:
			ydAlive = "1 day"

		case y == 0 && d > 1:
			ydAlive = fmt.Sprintf("%d days", d)

		case y == 1 && d == 0:
			ydAlive = "1 year"

		case y == 1 && d == 1:
			ydAlive = "1 year, 1 day"

		case y == 1 && d > 1:
			ydAlive = fmt.Sprintf("1 year, %d days", d)

		case y > 1 && d == 0:
			ydAlive = fmt.Sprintf("%d years", y)

		case y > 1 && d == 1:
			ydAlive = fmt.Sprintf("%d years, 1 day", y)

		default:
			ydAlive = fmt.Sprintf("%d years, %d days", y, d)
		}

		resp.Figures = append(resp.Figures, respFigure{
			Link:      figure.Link,
			ImgSrc:    figure.ImgSrc,
			ImgAlt:    figure.ImgAlt,
			Name:      figure.Name,
			Desc:      figure.Desc,
			Born:      figure.Born.String(),
			Died:      figure.Died.String(),
			DaysAlive: numprinter(figure.DaysAlive),
			YDAlive:   ydAlive,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}
