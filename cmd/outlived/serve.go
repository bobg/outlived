package main

import (
	"context"

	"github.com/pkg/errors"

	"outlived/site"
)

func (c *maincmd) serve(ctx context.Context, contentDir string, args []string) error {
	s, err := site.NewServer(ctx, contentDir, c.projectID, c.locationID, c.dsClient, c.ctClient)
	if err != nil {
		return errors.Wrap(err, "creating server")
	}
	s.Serve(ctx)

	return nil
}
