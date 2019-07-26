package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"google.golang.org/api/option"

	"github.com/bobg/outlived/site"
)

func cliServe(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		addr      = flagset.String("addr", ":80", "web server listen address")
		smtpAddr  = flagset.String("smtp", "localhost:587", "smtp submission address")
		creds     = flagset.String("creds", "", "credentials file")
		projectID = flagset.String("project", "outlived-163105", "project ID")
		seed      = flagset.Int64("seed", time.Now().Unix(), "RNG seed")
		test      = flagset.Bool("test", false, "run in test mode")
	)

	err := flagset.Parse(args)
	if err != nil {
		return err
	}

	rand.Seed(*seed)

	if *test {
		if *creds != "" {
			log.Fatal("cannot supply both -test and -creds")
		}

		err := aesite.DSTest(ctx, *projectID)
		if err != nil {
			log.Fatal(err)
		}
	}

	var options []option.ClientOption
	if *creds != "" {
		options = append(options, option.WithCredentialsFile(*creds))
	}
	dsClient, err := datastore.NewClient(ctx, *projectID, options...)
	if err != nil {
		return nil, errors.Wrap(err, "creating datastore client")
	}

	s := site.NewServer(*addr, *smtpAddr, dsClient)
	s.Serve(ctx)

	return nil
}
