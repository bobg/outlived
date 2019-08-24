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

	"google.golang.org/api/iterator"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/bobg/basexx"
	"github.com/pkg/errors"

	"github.com/bobg/outlived"
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
	h := sha256.Sum256([]byte(inp))
	src := basexx.NewBuffer(h[:], basexx.Binary)
	buf := make([]byte, basexx.Length(256, 50, len(h)))
	dest := basexx.NewBuffer(buf[:], basexx.Base50)
	_, err := basexx.Convert(dest, src) // discard error
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
	// xxx auth

	ctx := req.Context()

	ltreq := &taskspb.ListTasksRequest{
		Parent: s.scrapeQueue(),
	}
	iter := s.ctClient.ListTasks(ctx, ltreq)
	_, err := iter.Next()
	if err != nil && err != iterator.Done {
		return errors.Wrap(err, "checking scrape queue for emptiness")
	}
	if err == nil {
		log.Print("scrape queue is not empty")
		return nil
	}

	log.Print("starting new scrape")

	// err == iterator.Done (i.e., the queue is empty)
	for m := time.January; m <= time.December; m++ {
		for d := 1; d <= daysInMonth[m]; d++ {
			_, err = s.ctClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
				Parent: s.scrapeQueue(),
				Task: &taskspb.Task{
					Name: s.taskName(fmt.Sprintf("%d/%d", m, d)),
					MessageType: &taskspb.Task_AppEngineHttpRequest{
						AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
							HttpMethod:  taskspb.HttpMethod_GET,
							RelativeUri: fmt.Sprintf("/scrapeday?m=%d&d=%d", m, d),
						},
					},
				},
			})
			if err != nil {
				return errors.Wrapf(err, "queueing /scrapeday task for m=%d, d=%d", m, d)
			}
		}
	}
	log.Print("queued new scrapeday tasks")
	return nil
}

func (s *Server) handleScrapeday(w http.ResponseWriter, req *http.Request) error {
	// xxx auth

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
	return outlived.ScrapeDay(ctx, time.Month(m), d, func(ctx context.Context, href, title, desc string) error {
		u, _ := url.Parse("/scrapeperson")

		v := url.Values{}
		v.Set("href", href)
		v.Set("title", title)
		v.Set("desc", desc)
		u.RawQuery = v.Encode()

		_, err := s.ctClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
			Parent: s.scrapeQueue(),
			Task: &taskspb.Task{
				Name: s.taskName(href),
				MessageType: &taskspb.Task_AppEngineHttpRequest{
					AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
						HttpMethod:  taskspb.HttpMethod_GET,
						RelativeUri: u.String(),
					},
				},
			},
		})
		return err
	})
}

func (s *Server) handleScrapeperson(w http.ResponseWriter, req *http.Request) error {
	// xxx auth

	var (
		href  = req.FormValue("href")
		title = req.FormValue("title")
		desc  = req.FormValue("desc")
	)

	log.Printf("scraping person %s (%s)", title, href)

	ctx := req.Context()
	return outlived.ScrapePerson(ctx, href, title, desc, func(ctx context.Context, title, desc, href string, bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, pageviews int) error {
		fig := &outlived.Figure{
			Name:      title,
			Desc:      desc,
			Link:      href,
			Born:      outlived.Date{Y: bornY, M: time.Month(bornM), D: bornD},
			Died:      outlived.Date{Y: diedY, M: time.Month(diedM), D: diedD},
			DaysAlive: aliveDays,
			Pageviews: pageviews,
			Updated:   time.Now(),
		}
		return outlived.ReplaceFigures(ctx, s.dsClient, []*outlived.Figure{fig})
	})
}
