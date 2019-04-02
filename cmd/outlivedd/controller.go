package main

import (
	"context"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2beta3"
	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type controller struct {
	dsClient   *datastore.Client
	ctClient   *cloudtasks.Client
	projectID  string
	locationID string
}

func newController(ctx context.Context, projectID, locationID string) (*controller, error) {
	options := []option.ClientOption{}

	dsClient, err := datastore.NewClient(ctx, projectID, options...)
	if err != nil {
		return nil, errors.Wrap(err, "creating datastore client")
	}
	ctClient, err := cloudtasks.NewClient(ctx, options...)
	if err != nil {
		return nil, errors.Wrap(err, "creating cloudtasks client")
	}
	return &controller{
		dsClient:   dsClient,
		ctClient:   ctClient,
		projectID:  projectID,
		locationID: locationID,
	}, nil
}
