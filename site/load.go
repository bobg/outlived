package site

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

func (s *Server) handleLoad(w http.ResponseWriter, req *http.Request) error {
	// xxx authorized callers only

	ctx := req.Context()
	csvr := csv.NewReader(req.Body)

	var figures []*outlived.Figure

	for {
		fields, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if len(fields) != 7 {
			return fmt.Errorf("cannot parse: %v", fields)
		}
		if err != nil {
			return errors.Wrap(err, "reading request")
		}
		born, err := outlived.ParseDate(fields[2])
		if err != nil {
			return codeErr(err, http.StatusBadRequest, "parsing %s", fields[2])
		}
		died, err := outlived.ParseDate(fields[3])
		if err != nil {
			return codeErr(err, http.StatusBadRequest, "parsing %s: %s", fields[3])
		}
		daysAlive := died.Since(born)
		pageViews, err := strconv.Atoi(fields[6])
		if err != nil {
			return codeErr(err, http.StatusBadRequest, "parsing pageviews count %s: %s", fields[6])
		}
		f := &outlived.Figure{
			Name:      fields[0],
			Desc:      fields[1],
			Link:      fields[5],
			Born:      born,
			Died:      died,
			DaysAlive: daysAlive,
			Pageviews: pageViews,
		}
		figures = append(figures, f)
	}
	err := outlived.ReplaceFigures(ctx, s.dsClient, figures)
	if err != nil {
		return errors.Wrap(err, "writing to datastore")
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}
