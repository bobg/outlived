package site

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
)

func NewServer(ctx context.Context, addr, smtpAddr, contentDir, projectID, locationID string, dsClient *datastore.Client, ctClient *cloudtasks.Client) (*Server, error) {
	s := &Server{
		addr:       addr,
		smtpAddr:   smtpAddr,
		contentDir: contentDir,
		projectID:  projectID,
		locationID: locationID,
		dsClient:   dsClient,
	}

	if ctClient == nil { // test mode
		s.tasks = newLocalTasks(ctx, addr)
		s.sender = new(testSender)
	} else {
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
	}
	return s, nil
}

type Server struct {
	addr       string
	smtpAddr   string
	contentDir string
	projectID  string
	locationID string
	dsClient   *datastore.Client
	tasks      taskService
	sender     sender
}

func (s *Server) Serve(ctx context.Context) {
	handle("/", s.handleHome)
	handle("/load", s.handleLoad)
	handle("/login", s.handleLogin)
	handle("/logout", s.handleLogout)
	handle("/setactive", s.handleSetActive)
	handle("/signup", s.handleSignup)
	handle("/verify", s.handleVerify)

	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(s.contentDir))))

	// cron-initiated
	handle("/task/scrape", s.handleScrape)
	handle("/task/expire", s.handleExpire)
	handle("/task/send", s.handleSend)

	// task-queue-initiated
	handle("/task/scrapeday", s.handleScrapeday)
	handle("/task/scrapeperson", s.handleScrapeperson)

	log.Printf("listening for requests on %s", s.addr)

	srv := &http.Server{Addr: s.addr}
	go srv.ListenAndServe()

	<-ctx.Done()
	srv.Shutdown(ctx)
}

func httpErr(w http.ResponseWriter, code int, format string, args ...interface{}) {
	if code == 0 {
		code = http.StatusInternalServerError
	}

	if format == "" && len(args) == 0 {
		format = "%s"
		args = []interface{}{http.StatusText(code)}
	}

	log.Printf(format, args...)
	http.Error(w, fmt.Sprintf(format, args...), code)
}

type handlerFunc func(http.ResponseWriter, *http.Request) error

func handlerCaller(f handlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("%s %s", req.Method, req.URL.String())

		ww := &respWriter{w: w}
		err := f(ww, req)
		if err != nil {
			code := http.StatusInternalServerError
			if err, ok := err.(codeErrType); ok {
				code = err.code
			}
			log.Printf("%s", err)
			http.Error(w, err.Error(), code)
			return
		}
		if !ww.writeCalled {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func handle(pattern string, f handlerFunc) {
	http.HandleFunc(pattern, handlerCaller(f))
}

type codeErrType struct {
	err  error // can be nil
	code int
}

func (e codeErrType) Error() string {
	s := http.StatusText(e.code)
	if s == "" {
		s = fmt.Sprintf("HTTP status %d", e.code)
	} else {
		s = fmt.Sprintf("%s (HTTP status %d)", s, e.code)
	}
	if e.err != nil {
		return fmt.Sprintf("%s: %s", e.err, s)
	}
	return s
}

func codeErr(err error, code int, args ...interface{}) error {
	if len(args) > 0 {
		f := args[0].(string)
		err = errors.Wrapf(err, f, args[1:]...)
	}
	return codeErrType{err: err, code: code}
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
func checkCron(req *http.Request) error {
	h := strings.TrimSpace(req.Header.Get("X-Appengine-Cron"))
	if h != "true" {
		return codeErrType{code: http.StatusUnauthorized}
	}
	return nil
}

// See
// https://cloud.google.com/tasks/docs/creating-appengine-handlers#reading_request_headers.
func checkTaskQueue(req *http.Request, queue string) error {
	h := strings.TrimSpace(req.Header.Get("X-AppEngine-QueueName"))
	if h != queue {
		return codeErrType{code: http.StatusUnauthorized}
	}
	return nil
}
