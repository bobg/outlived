package site

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"golang.org/x/text/message"

	"github.com/bobg/outlived"
)

type (
	dataResp struct {
		// Figures is a list of figures that died on this date.
		Figures []figureData `json:"figures"`

		// Today is today's date.
		Today string `json:"today"`

		// User is info about the signed-in user.
		User *userData `json:"user"`
	}

	figureData struct {
		Name string `json:"name"`
		Desc string `json:"desc"`

		Born string `json:"born"`
		Died string `json:"died"`

		Days      string `json:"days"`
		YearsDays string `json:"yearsDays"`

		Href   string `json:"href"`
		ImgAlt string `json:"imgAlt"`
		ImgSrc string `json:"imgSrc"`
	}

	userData struct {
		CSRF      string       `json:"csrf"`
		Born      string       `json:"born"`
		Days      string       `json:"days"`
		YearsDays string       `json:"yearsDays"`
		Email     string       `json:"email"`
		Figures   []figureData `json:"figures"`
	}
)

func (s *Server) handleData(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	tzname := req.FormValue("tzname")
	loc, err := time.LoadLocation(tzname)
	if err != nil {
		log.Printf("error loading timezone %s, falling back to UTC: %s", tzname, err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := outlived.TimeDate(now)
	resp := dataResp{Today: now.Format("Monday, 2 Jan 2006")}

	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}

	p := message.NewPrinter(message.MatchLanguage("en"))
	numprinter := func(n int) string {
		return p.Sprintf("%v", n)
	}

	toFigureData := func(figure *outlived.Figure) figureData {
		return figureData{
			Name:      figure.Name,
			Desc:      figure.Desc,
			Born:      figure.Born.String(),
			Died:      figure.Died.String(),
			Days:      numprinter(figure.DaysAlive),
			YearsDays: figure.Died.YDSinceStr(figure.Born),
			Href:      "https://en.wikipedia.org/wiki/" + figure.Link,
			ImgAlt:    figure.ImgAlt,
			ImgSrc:    figure.ImgSrc,
		}
	}

	figures, err := outlived.FiguresDiedOn(ctx, s.dsClient, today.M, today.D, 24)
	if err != nil {
		return errors.Wrapf(err, "getting figures that died on %d %s", today.D, today.M)
	}
	for _, figure := range figures {
		f := toFigureData(figure)
		resp.Figures = append(resp.Figures, f)
	}

	if sess != nil {
		var u outlived.User
		err = sess.GetUser(ctx, s.dsClient, &u)
		if err != nil {
			return errors.Wrap(err, "getting user from session")
		}

		csrf, err := sess.CSRFToken()
		if err != nil {
			return errors.Wrap(err, "generating CSRF token")
		}

		alive := today.Since(u.Born)

		d := userData{
			CSRF:      csrf,
			Born:      u.Born.String(),
			Days:      numprinter(alive),
			YearsDays: today.YDSinceStr(u.Born),
			Email:     u.Email,
		}

		figures, err := outlived.FiguresAliveForAtMost(ctx, s.dsClient, alive-1, 24)
		if err != nil {
			return errors.Wrapf(err, "getting figures that died %d days ago", alive-1)
		}
		for _, figure := range figures {
			f := toFigureData(figure)
			d.Figures = append(d.Figures, f)
		}

		resp.User = &d
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(resp)
	return errors.Wrap(err, "encoding response")
}
