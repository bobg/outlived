package site

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"google.golang.org/api/iterator"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/pkg/errors"

	"github.com/bobg/outlived"
)

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

	// xxx range-check m and d

	ctx := req.Context()
	return outlived.ScrapeDay(ctx, time.Month(m), d, func(ctx context.Context, href, title, desc string) error {
		u, _ := url.Parse("/scrapeperson")

		v := url.Values{}
		v.Set("href", href)
		v.Set("title", title)
		v.Set("desc", desc)
		u.RawQuery = v.Encode()

		_, err := s.ctClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
			Parent: xxx,
			Task: &taskspb.Task{
				Name: xxx,
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
		}
		return outlived.ReplaceFigures(ctx, s.dsClient, []*outlived.Figure{fig})
	})
}

// Runs as a goroutine.
// Once a day, it checks to see if the "scrape" cloudtasks queue is empty.
// If it is, it kicks off a new scrape.
func (s *Server) scrape(ctx context.Context) {
	if s.ctClient == nil {
		return
	}

	defer log.Print("exiting scrape goroutine")

	ticker := time.NewTicker(24 * time.Hour)
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			req := &taskspb.ListTasksRequest{
				Parent: xxx,
			}
			iter := s.ctClient.ListTasks(ctx, req)
			_, err := iter.Next()
			if err != nil && err != iterator.Done {
				log.Printf("scrape goroutine: error listing queue tasks: %s", err)
				continue
			}
			if err == nil {
				continue
			}
			// err == iterator.Done (i.e., the queue is empty)
			for m := time.January; m <= time.December; m++ {
				for d := 1; d <= daysInMonth[m]; d++ {
					_, err = s.ctClient.CreateTask(ctx, &taskspb.CreateTaskRequest{
						Parent: xxx,
						Task: &taskspb.Task{
							Name: xxx,
							MessageType: &taskspb.Task_AppEngineHttpRequest{
								AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
									HttpMethod:  taskspb.HttpMethod_GET,
									RelativeUri: fmt.Sprintf("/scrapeday?m=%d&d=%d", m, d),
								},
							},
						},
					})
					if err != nil {
						log.Printf("error queueing /scrapeday task for m=%d, d=%d: %s", m, d, err)
					}
				}
			}
		}
	}
}
