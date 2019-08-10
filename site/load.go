package site

import (
	"encoding/csv"
	"io"
	"net/http"
	"strconv"

	"github.com/bobg/outlived"
)

func (s *Server) handleLoad(w http.ResponseWriter, req *http.Request) {
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
			httpErr(w, 0, "cannot parse: %v", fields)
			return
		}
		if err != nil {
			httpErr(w, 0, "reading request: %s", err)
			return
		}
		born, err := outlived.ParseDate(fields[2])
		if err != nil {
			httpErr(w, http.StatusBadRequest, "parsing %s: %s", fields[2], err)
			return
		}
		died, err := outlived.ParseDate(fields[3])
		if err != nil {
			httpErr(w, http.StatusBadRequest, "parsing %s: %s", fields[3], err)
			return
		}
		daysAlive := died.Since(born)
		pageViews, err := strconv.Atoi(fields[6])
		if err != nil {
			httpErr(w, http.StatusBadRequest, "parsing pageviews count %s: %s", fields[6], err)
			return
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
		httpErr(w, 0, "writing to datastore: %s", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
