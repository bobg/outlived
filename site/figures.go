package site

import (
	"encoding/json"
	"net/http"

	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"golang.org/x/text/message"

	"github.com/bobg/outlived"
)

type (
	figuresResp struct {
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
	}
)

func (s *Server) handleFigures(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		return errors.Wrap(err, "getting session")
	}

	today := outlived.Today()

	p := message.NewPrinter(message.MatchLanguage("en"))
	numprinter := func(n int) string {
		return p.Sprintf("%v", n)
	}

	var figures []*outlived.Figure
	if sess != nil {
		u := new(outlived.User)
		err = s.dsClient.Get(ctx, sess.UserKey, u)
		if err != nil {
			return errors.Wrap(err, "getting user")
		}

		alive := today.Since(u.Born)
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

	var resp figuresResp
	for _, figure := range figures {
		resp.Figures = append(resp.Figures, respFigure{
			Link:      figure.Link,
			ImgSrc:    figure.ImgSrc,
			ImgAlt:    figure.ImgAlt,
			Name:      figure.Name,
			Desc:      figure.Desc,
			Born:      figure.Born.String(),
			Died:      figure.Died.String(),
			DaysAlive: numprinter(figure.DaysAlive),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(resp)
}
