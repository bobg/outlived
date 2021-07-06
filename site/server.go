package site

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/bobg/mid"
	"github.com/pkg/errors"
	"google.golang.org/appengine"
)

func NewServer(ctx context.Context, contentDir, projectID, locationID string, dsClient *datastore.Client, ctClient *cloudtasks.Client) (*Server, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	s := &Server{
		addr:       addr,
		contentDir: contentDir,
		projectID:  projectID,
		locationID: locationID,
		dsClient:   dsClient,
	}

	if appengine.IsAppEngine() {
		s.tasks = (*gCloudTasks)(ctClient)

		domain, err := aesite.GetSetting(ctx, dsClient, "mailgun_domain")
		if err != nil {
			return nil, errors.Wrap(err, "getting setting for mailgun_domain")
		}
		apiKey, err := aesite.GetSetting(ctx, dsClient, "mailgun_api_key")
		if err != nil {
			return nil, errors.Wrap(err, "getting setting for mailgun_api_key")
		}
		s.sender = newMailgunSender(string(domain), string(apiKey))
	} else {
		s.tasks = newLocalTasks(ctx, addr)
		s.sender = new(testSender)
	}

	return s, nil
}

type Server struct {
	addr       string
	contentDir string
	projectID  string
	locationID string
	dsClient   *datastore.Client
	tasks      taskService
	sender     sender
}

func (s *Server) Serve(ctx context.Context) {
	mux := http.NewServeMux()

	// This is for testing. In production, / is routed by app.yaml.
	mux.Handle("/", mid.Err(s.handleStatic))

	mux.Handle("/s/data", s.sessHandler(mid.JSON(s.handleData)))
	mux.Handle("/s/forgot", mid.Err(s.handleForgot))
	mux.Handle("/s/load", mid.Err(s.handleLoad))
	mux.Handle("/s/login", mid.JSON(s.handleLogin))
	mux.Handle("/s/logout", mid.Err(s.handleLogout))
	mux.Handle("/s/resetpw", mid.Err(s.handleResetPW))
	mux.Handle("/s/reverify", s.sessHandler(mid.JSON(s.handleReverify)))
	mux.Handle("/s/setactive", s.sessHandler(mid.JSON(s.handleSetActive)))
	mux.Handle("/s/setbirthdate", s.sessHandler(mid.JSON(s.handleSetBirthdate)))
	mux.Handle("/s/signup", mid.JSON(s.handleSignup))
	mux.Handle("/s/verify", mid.Err(s.handleVerify))

	mux.Handle("/unsubscribe", http.RedirectHandler("/", http.StatusMovedPermanently))

	mux.Handle("/r", mid.Err(s.handleRedirect))

	// cron-initiated
	mux.Handle("/t/scrape", mid.Err(s.handleScrape))
	mux.Handle("/t/expire", mid.Err(s.handleExpire))
	mux.Handle("/t/send", mid.Err(s.handleSend))

	// task-queue-initiated
	mux.Handle("/t/scrapeday", mid.Err(s.handleScrapeday))
	mux.Handle("/t/scrapeperson", mid.Err(s.handleScrapeperson))

	log.Printf("listening for requests on %s", s.addr)

	srv := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	if appengine.IsAppEngine() {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		go srv.ListenAndServe()
		<-ctx.Done()
		srv.Shutdown(ctx)
	}
}

// Function respWriter wraps an http.ResponseWriter,
// delegating calls to it.
// It tracks whether Write or WriteHeader is ever called.
type respWriter struct {
	w           http.ResponseWriter
	writeCalled bool
}

func (w *respWriter) Header() http.Header {
	return w.w.Header()
}

func (w *respWriter) Write(b []byte) (int, error) {
	w.writeCalled = true
	return w.w.Write(b)
}

func (w *respWriter) WriteHeader(code int) {
	w.writeCalled = true
	w.w.WriteHeader(code)
}

// See
// https://cloud.google.com/appengine/docs/standard/go112/scheduling-jobs-with-cron-yaml#validating_cron_requests.
func (s *Server) checkCron(req *http.Request) error {
	if !appengine.IsAppEngine() {
		return nil
	}

	h := strings.TrimSpace(req.Header.Get("X-Appengine-Cron"))
	if h != "true" {
		return mid.CodeErr{C: http.StatusUnauthorized}
	}
	return nil
}

// See
// https://cloud.google.com/tasks/docs/creating-appengine-handlers#reading_request_headers.
func (s *Server) checkTaskQueue(req *http.Request, queue string) error {
	if !appengine.IsAppEngine() {
		return nil
	}

	ctx := req.Context()
	masterKey, err := aesite.GetSetting(ctx, s.dsClient, "master-key")
	if err == nil && strings.TrimSpace(req.Header.Get("X-Outlived-Key")) == string(masterKey) {
		return nil
	}

	h := strings.TrimSpace(req.Header.Get("X-AppEngine-QueueName"))
	if h != queue {
		return mid.CodeErr{C: http.StatusUnauthorized}
	}
	return nil
}

var homeURL *url.URL

func init() {
	if appengine.IsAppEngine() {
		homeURL = &url.URL{
			Scheme: "https",
			Host:   "outlived.net",
			Path:   "/",
		}
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		homeURL = &url.URL{
			Scheme: "http",
			Host:   "localhost:" + port,
			Path:   "/",
		}
	}
}
