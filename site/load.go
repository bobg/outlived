package site

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bobg/aesite"
	"github.com/bobg/mid"
	"github.com/pkg/errors"

	"outlived"
)

func (s *Server) handleLoad(w http.ResponseWriter, req *http.Request) error {
	ctx := req.Context()

	// Authorized callers only.
	masterKey, err := aesite.GetSetting(ctx, s.dsClient, "master-key")
	if err != nil {
		return errors.Wrap(err, "getting master key")
	}
	if strings.TrimSpace(req.Header.Get("X-Outlived-Key")) != string(masterKey) {
		return mid.CodeErr{C: http.StatusUnauthorized}
	}

	csvr := csv.NewReader(req.Body)
	now := time.Now()

	var figures []*outlived.Figure

	for {
		fields, err := csvr.Read()
		if err == io.EOF {
			break
		}
		if len(fields) != 9 {
			return fmt.Errorf("cannot parse: %v", fields)
		}
		if err != nil {
			return errors.Wrap(err, "reading request")
		}
		born, err := outlived.ParseDate(fields[2])
		if err != nil {
			return errors.Wrapf(mid.CodeErr{C: http.StatusBadRequest, Err: err}, "parsing %s", fields[2])
		}
		died, err := outlived.ParseDate(fields[3])
		if err != nil {
			return errors.Wrapf(mid.CodeErr{C: http.StatusBadRequest, Err: err}, "parsing %s", fields[3])
		}
		daysAlive := died.Since(born)
		pageViews, err := strconv.Atoi(fields[8])
		if err != nil {
			return errors.Wrapf(mid.CodeErr{C: http.StatusBadRequest, Err: err}, "parsing pageviews count %s", fields[8])
		}
		f := &outlived.Figure{
			Name:      fields[0],
			Desc:      fields[1],
			Link:      fields[5],
			ImgSrc:    fields[6],
			ImgAlt:    fields[7],
			Born:      born,
			Died:      died,
			DaysAlive: daysAlive,
			Pageviews: pageViews,
			Updated:   now,
		}
		figures = append(figures, f)
	}
	err = outlived.ReplaceFigures(ctx, s.dsClient, figures)
	return errors.Wrap(err, "writing to datastore")
}
