package site

import (
	"log"
	"net/http"

	"outlived"
)

// Function handleExpire expires stale figures.
func (s *Server) handleExpire(w http.ResponseWriter, req *http.Request) error {
	err := s.checkCron(req)
	if err != nil {
		return err
	}

	count, err := outlived.ExpireFigures(req.Context(), s.dsClient)
	log.Printf("expired %d stale figure(s)", count)
	return err
}
