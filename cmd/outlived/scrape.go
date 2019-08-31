package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/bobg/outlived"
)

var (
	monthName = []string{
		"",
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}

	dateRegex1 = regexp.MustCompile(`(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d+),\s+(\d+)(\s+BC)?`)
	dateRegex2 = regexp.MustCompile(`(\d+)\s+(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d+)(\s+BC)?`)

	daysInMonth = []int{
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

	bRegex = regexp.MustCompile(`\(b\.\s*\d+\)$`)
)

func cliScrape(ctx context.Context, flagset *flag.FlagSet, args []string) error {
	var (
		startM = flagset.Int("startm", 1, "start month")
		startD = flagset.Int("startd", 1, "start day (of month)")
		limit  = flagset.Duration("limit", time.Second, "rate limit")
	)
	err := flagset.Parse(args)
	if err != nil {
		return err
	}

	w := csv.NewWriter(os.Stdout)

	limiter := rate.NewLimiter(rate.Every(*limit), 1)
	for m := time.Month(*startM); m <= time.December; m++ {
		for d := *startD; d <= daysInMonth[m]; d++ {
			err = limiter.Wait(ctx)
			if err != nil {
				return errors.Wrapf(err, "waiting to scrape day %d-%d", m, d)
			}
			err = outlived.ScrapeDay(ctx, m, d, func(ctx context.Context, href, title, desc string) error {
				err := limiter.Wait(ctx)
				if err != nil {
					return errors.Wrapf(err, "waiting to scrape person %s (%s)", title, href)
				}
				return outlived.ScrapePerson(ctx, href, title, desc, func(ctx context.Context, title, desc, href, imgSrc, imgAlt string, bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, pageviews int) error {
					bornStr := fmt.Sprintf("%d-%02d-%02d", bornY, bornM, bornD)
					diedStr := fmt.Sprintf("%d-%02d-%02d", diedY, diedM, diedD)

					return w.Write([]string{title, desc, bornStr, diedStr, strconv.Itoa(aliveDays), href, imgSrc, imgAlt, strconv.Itoa(pageviews)})
				})
			})
			if err != nil {
				return errors.Wrapf(err, "getting date %d/%d", m, d)
			}
		}
		w.Flush()
	}

	return nil
}
