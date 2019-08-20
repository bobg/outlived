package site

import (
	"context"
	"log"
	"time"

	"github.com/bobg/outlived"
)

// Function expire runs as a goroutine
// and periodically deletes figures that have not been updated lately.
func (s *Server) expire(ctx context.Context) {
	if s.dsClient == nil {
		return
	}

	defer log.Print("exiting expire goroutine")

	ticker := time.NewTicker(24 * time.Hour)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			err := outlived.ExpireFigures(ctx, s.dsClient)
			if err != nil {
				log.Printf("expiring figures: %s", err)
			}
		}
	}
}
