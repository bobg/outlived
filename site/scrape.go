package site

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bobg/basexx"
	"github.com/pkg/errors"

	"outlived"
)

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

func (s *Server) scrapeQueue() string {
	return fmt.Sprintf("projects/%s/locations/%s/queues/scrape", s.projectID, s.locationID)
}

func (s *Server) taskName(inp string) string {
	hasher := sha256.New()
	hasher.Write([]byte{1}) // version of this hash
	hasher.Write([]byte(inp))
	h := hasher.Sum(nil)
	src := basexx.NewBuffer(h, basexx.Binary)
	buf := make([]byte, basexx.Length(256, 50, len(h)))
	dest := basexx.NewBuffer(buf[:], basexx.Base50)
	_, err := basexx.Convert(dest, src)
	if err != nil {
		panic(err)
	}
	converted := dest.Written()
	return fmt.Sprintf("%s/tasks/%s", s.scrapeQueue(), string(converted))
}

// Function handleScrape launches a new scrape: one task for each day of the year.
// (Each handled by handleScrapeday.)
// A task is queued only if the scrape queue is empty.
func (s *Server) handleScrape(w http.ResponseWriter, req *http.Request) error {
	err := s.checkCron(req)
	if err != nil {
		return err
	}

	ctx := req.Context()
	empty, err := s.tasks.queueEmpty(ctx, s.scrapeQueue())
	if err != nil {
		return errors.Wrap(err, "checking scrape queue for emptiness")
	}
	if !empty {
		log.Print("scrape queue is not empty")
		return nil
	}

	log.Print("starting new scrape")

	// err == iterator.Done (i.e., the queue is empty)
	for m := time.January; m <= time.December; m++ {
		for d := 1; d <= daysInMonth[m]; d++ {
			err = s.tasks.enqueueTask(
				ctx,
				s.scrapeQueue(),
				s.taskName(fmt.Sprintf("%d/%d", m, d)),
				fmt.Sprintf("/t/scrapeday?m=%d&d=%d", m, d),
			)
			if err != nil {
				return errors.Wrapf(err, "queueing /scrapeday task for m=%d, d=%d", m, d)
			}
		}
	}

	return nil
}

func (s *Server) handleScrapeday(w http.ResponseWriter, req *http.Request) error {
	err := s.checkTaskQueue(req, "scrape")
	if err != nil {
		return err
	}

	mstr := req.FormValue("m")
	m, err := strconv.Atoi(mstr)
	if err != nil {
		return errors.Wrapf(err, "parsing value for m: %s", mstr)
	}

	dstr := req.FormValue("d")
	d, err := strconv.Atoi(dstr)
	if err != nil {
		return errors.Wrapf(err, "parsing value for d: %s", dstr)
	}

	if m < 1 || m > 12 || d < 1 || d > daysInMonth[m] {
		return errors.Wrapf(err, "month %d, day %d is out of range", m, d)
	}

	log.Printf("scraping day %s %d", time.Month(m), d)

	ctx := req.Context()
	return outlived.ScrapeDay(ctx, new(http.Client), time.Month(m), d, func(ctx context.Context, href, title, desc string) error {
		u, _ := url.Parse("/t/scrapeperson")

		v := url.Values{}
		v.Set("href", href)
		v.Set("title", title)
		v.Set("desc", desc)
		u.RawQuery = v.Encode()

		err := s.tasks.enqueueTask(
			ctx,
			s.scrapeQueue(),
			s.taskName(href),
			u.String(),
		)
		if err != nil {
			log.Printf("enqueueing scrapeperson task for %s (%s): %s", title, href, err)
			// otherwise ignore error
		}
		return nil
	})
}

func (s *Server) handleScrapeperson(w http.ResponseWriter, req *http.Request) error {
	err := s.checkTaskQueue(req, "scrape")
	if err != nil {
		return err
	}

	var (
		href  = req.FormValue("href")
		title = req.FormValue("title")
		desc  = req.FormValue("desc")
	)

	ctx := req.Context()
	err = outlived.ScrapePerson(ctx, new(http.Client), href, title, desc, func(ctx context.Context, title, desc, href, imgSrc, alt string, bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, pageviews int) error {
		fig := &outlived.Figure{
			Name:      title,
			Desc:      desc,
			Link:      href,
			ImgSrc:    imgSrc,
			ImgAlt:    alt,
			Born:      outlived.Date{Y: bornY, M: time.Month(bornM), D: bornD},
			Died:      outlived.Date{Y: diedY, M: time.Month(diedM), D: diedD},
			DaysAlive: aliveDays,
			Pageviews: pageviews,
			Updated:   time.Now(),
		}
		return outlived.ReplaceFigures(ctx, s.dsClient, []*outlived.Figure{fig})
	})
	if err != nil {
		log.Printf("scraping person %s: %s", title, err)
		// Otherwise ignore this error. We'll get this person next time round.
		// (Or the error will persist and the person will expire out of the datastore.)
	}

	return nil
}
