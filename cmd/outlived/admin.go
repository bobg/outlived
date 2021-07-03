package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"outlived"
)

var adminCommands = map[string]func(context.Context, *flag.FlagSet, []string) error{
	"list-figures": cliAdminListFigures,
	"list-users":   cliAdminListUsers,
	"get":          cliAdminGet,
	"set":          cliAdminSet,
	"scrape":       cliAdminScrape,
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

func cliAdminListFigures(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		creds     = flagset.String("creds", "", "credentials file")
		projectID = flagset.String("project", "outlived-163105", "project ID")
		test      = flagset.Bool("test", false, "run in test mode")
		diedStr   = flagset.String("died", "", "died-on date, like Jan-2")
		limit     = flagset.Int("limit", 100, "limit on figures to return")
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
			return err
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

	died, err := time.Parse("Jan-2", *diedStr)
	if err != nil {
		return err
	}

	figs, err := outlived.FiguresDiedOn(ctx, dsClient, died.Month(), died.Day(), *limit)
	if err != nil {
		return err
	}
	for _, fig := range figs {
		fmt.Printf("%s (%s), born %s, died %s, alive %d days, %d pageviews\n", fig.Name, fig.Link, fig.Born, fig.Died, fig.DaysAlive, fig.Pageviews)
	}
	return nil
}

func cliAdminListUsers(ctx context.Context, flagset *flag.FlagSet, args []string) error {
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
			return fmt.Errorf("cannot supply both -test and -creds")
		}

		done, err := aesite.DSTestWithDoneChan(ctx, *projectID)
		if err != nil {
			return err
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
			return fmt.Errorf("cannot supply both -test and -creds")
		}

		done, err := aesite.DSTestWithDoneChan(ctx, *projectID)
		if err != nil {
			return err
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
			return fmt.Errorf("cannot supply both -test and -creds")
		}

		done, err := aesite.DSTestWithDoneChan(ctx, *projectID)
		if err != nil {
			return err
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

	return aesite.SetSetting(ctx, dsClient, flagset.Arg(0), []byte(flagset.Arg(1)))
}

var daysInMonth = []int{
	0,
	31,
	29,
	31,
	30,
	31,
	30,
	31,
	31,
	30,
	31,
	30,
	31,
}

func cliAdminScrape(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		creds     = flagset.String("creds", "", "credentials file")
		projectID = flagset.String("project", "outlived-163105", "project ID")
		test      = flagset.Bool("test", false, "run in test mode")
		monthStr  = flagset.String("month", "", "3-letter month")
		onlyDay   = flagset.Int("day", 0, "day of month")
		limit     = flagset.Duration("limit", time.Second, "rate limit")
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
			return err
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

	var (
		startMonth = time.January
		endMonth   = time.December
	)
	if *monthStr == "" && *onlyDay != 0 {
		return fmt.Errorf("must specify -month with -day")
	}
	client := &http.Client{
		Transport: &rlroundtripper{
			limiter: rate.NewLimiter(rate.Every(*limit), 1),
			rt:      http.DefaultTransport,
		},
	}
	if *monthStr != "" {
		d, err := time.Parse("Jan", *monthStr)
		if err != nil {
			return err
		}
		if *onlyDay != 0 {
			return scrapeMonthDay(ctx, client, dsClient, d.Month(), *onlyDay)
		}
		startMonth = d.Month()
		endMonth = d.Month()
	}
	for m := startMonth; m <= endMonth; m++ {
		for d := 1; d <= daysInMonth[m]; d++ {
			err = scrapeMonthDay(ctx, client, dsClient, m, d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func scrapeMonthDay(ctx context.Context, client *http.Client, dsClient *datastore.Client, m time.Month, d int) error {
	return outlived.ScrapeDay(ctx, client, m, d, func(ctx context.Context, href, title, desc string) error {
		log.Printf("scraping %s-%d", m, d)
		err := outlived.ScrapePerson(ctx, client, href, title, desc,
			func(
				ctx context.Context,
				title, desc, href, imgSrc, imgAlt string,
				bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, pageviews int,
			) error {
				log.Printf("updating person %s (href %s), %d pageviews", title, href, pageviews)
				fig := &outlived.Figure{
					Name:      title,
					Desc:      desc,
					Link:      href,
					ImgSrc:    imgSrc,
					ImgAlt:    imgAlt,
					Born:      outlived.Date{Y: bornY, M: time.Month(bornM), D: bornD},
					Died:      outlived.Date{Y: diedY, M: time.Month(diedM), D: diedD},
					DaysAlive: aliveDays,
					Pageviews: pageviews,
					Updated:   time.Now(),
				}
				return outlived.ReplaceFigures(ctx, dsClient, []*outlived.Figure{fig})
			})
		if err != nil {
			log.Printf("ERROR: %s", err)
		}
		return nil
	})
}

type rlroundtripper struct {
	limiter *rate.Limiter
	rt      http.RoundTripper
}

func (rt *rlroundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	err := rt.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	return rt.rt.RoundTrip(req)
}
