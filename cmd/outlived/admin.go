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

var adminCommands = map[string]func(context.Context, *flag.FlagSet, []string) error{
	"list": cliAdminList,
	"get":  cliAdminGet,
	"set":  cliAdminSet,
}

func cliAdmin(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	err := flagset.Parse(args)
	if err != nil {
		return err
	}

	if flagset.NArg() == 0 {
		return errors.New("usage: outlived admin <subcommand> [args]")
	}

	cmd := flagset.Arg(0)
	fn, ok := adminCommands[cmd]
	if !ok {
		return fmt.Errorf("unknown admin subcommand %s", cmd)
	}

	args = flagset.Args()
	return fn(ctx, flag.NewFlagSet("", flag.ContinueOnError), args[1:])
}

func cliAdminList(ctx context.Context, flagset *flag.FlagSet, args []string) error {
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
		_, err := it.Next(&u)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "iterating over users")
		}

		fmt.Printf("%+v\n", u)
	}
}

func cliAdminGet(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		creds     = flagset.String("creds", "", "credentials file")
		projectID = flagset.String("project", "outlived-163105", "project ID")
		test      = flagset.Bool("test", false, "run in test mode")
	)

	err := flagset.Parse(args)
	if err != nil {
		return err
	}

	if flagset.NArg() != 1 {
		return errors.New("usage: outlived admin get VAR")
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

	val, err := aesite.GetSetting(ctx, dsClient, flagset.Arg(0))
	if err != nil {
		return err
	}

	fmt.Println(string(val))
	return nil
}

func cliAdminSet(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		creds     = flagset.String("creds", "", "credentials file")
		projectID = flagset.String("project", "outlived-163105", "project ID")
		test      = flagset.Bool("test", false, "run in test mode")
	)

	err := flagset.Parse(args)
	if err != nil {
		return err
	}

	if flagset.NArg() != 2 {
		return errors.New("usage: outlived admin set VAR VALUE")
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

	return aesite.SetSetting(ctx, dsClient, flagset.Arg(0), []byte(flagset.Arg(1)))
}
