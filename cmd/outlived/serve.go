package main

import (
	"context"
	"flag"
	"fmt"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"google.golang.org/api/option"

	"outlived/site"
)

func cliServe(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		creds      = flagset.String("creds", "", "credentials file")
		contentDir = flagset.String("dir", "web/build", "content dir")
		projectID  = flagset.String("project", "outlived-163105", "project ID")
		locationID = flagset.String("location", "us-central1", "location ID")
		test       = flagset.Bool("test", false, "run in test mode")
	)

	err := flagset.Parse(args)
	if err != nil {
		return err
	}

	if *test {
		if *creds != "" {
			return fmt.Errorf("cannot supply both -test and -creds")
		}

		done, err := aesite.DSTestWithDoneChan(ctx, *projectID)
		if err != nil {
			return errors.Wrap(err, "starting test datastore service")
		}
		defer func() { <-done }()
	}

	var options []option.ClientOption
	if *creds != "" {
		options = append(options, option.WithCredentialsFile(*creds))
	}
	dsClient, err := datastore.NewClient(ctx, *projectID, options...)
	if err != nil {
		return errors.Wrap(err, "creating datastore client")
	}

	var ctClient *cloudtasks.Client
	if !*test {
		ctClient, err = cloudtasks.NewClient(ctx, options...)
		if err != nil {
			return errors.Wrap(err, "creating cloudtasks client")
		}
	}

	s, err := site.NewServer(ctx, *contentDir, *projectID, *locationID, dsClient, ctClient)
	if err != nil {
		return errors.Wrap(err, "creating server")
	}
	s.Serve(ctx)

	return nil
}
