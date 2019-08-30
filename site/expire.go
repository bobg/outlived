package site

import (
	"log"
	"net/http"

	"github.com/bobg/outlived"
)

// Function handleExpire expires stale figures.
func (s *Server) handleExpire(w http.ResponseWriter, req *http.Request) error {
	err := s.checkCron(req)
	if err != nil {
		return err
	}

	log.Print("expiring stale figures")
	return outlived.ExpireFigures(req.Context(), s.dsClient)
}
