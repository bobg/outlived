package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/bobg/outlived"
)

func cliAdmin(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		creds     = flagset.String("creds", "", "credentials file")
		projectID = flagset.String("project", "outlived-163105", "project ID")
		test      = flagset.Bool("test", false, "run in test mode")
	)

	err := flagset.Parse(args)
	if err != nil {
		return err
	}

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
		return errors.Wrap(err, "creating datastore client")
	}

	q := datastore.NewQuery("User")
	it := dsClient.Run(ctx, q)
	for {
		var u outlived.User
		_, err = it.Next(&u)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "iterating over users")
		}
		fmt.Printf("%+v\n", u)
	}
}
