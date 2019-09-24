package site

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/bobg/hj"
	"github.com/pkg/errors"
	"golang.org/x/text/message"
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
		p:          message.NewPrinter(message.MatchLanguage("en")),
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
	home       *url.URL
	p          *message.Printer
}

func (s *Server) Serve(ctx context.Context) {
	mux := http.NewServeMux()

	// This is for testing. In production, / is routed by app.yaml.
	handleErrFunc(mux, "/", s.handleStatic)

	mux.Handle("/s/data", s.sessHandler(hj.Handler(s.handleData, onErr)))

	handleErrFunc(mux, "/s/forgot", s.handleForgot)
	handleErrFunc(mux, "/s/load", s.handleLoad)
	mux.Handle("/s/login", hj.Handler(s.handleLogin, onErr))
	handleErrFunc(mux, "/s/logout", s.handleLogout)
	mux.Handle("/s/resetpw", hj.Handler(s.handleResetPW, onErr))
	mux.Handle("/s/reverify", s.sessHandler(hj.Handler(s.handleReverify, onErr)))
	mux.Handle("/s/setactive", s.sessHandler(hj.Handler(s.handleSetActive, onErr)))
	mux.Handle("/s/signup", hj.Handler(s.handleSignup, onErr))
	handleErrFunc(mux, "/s/verify", s.handleVerify)

	mux.Handle("/s/unsubscribe", http.RedirectHandler("/", http.StatusMovedPermanently))

	handleErrFunc(mux, "/r", s.handleRedirect)

	// cron-initiated
	handleErrFunc(mux, "/t/scrape", s.handleScrape)
	handleErrFunc(mux, "/t/expire", s.handleExpire)
	handleErrFunc(mux, "/t/send", s.handleSend)

	// task-queue-initiated
	handleErrFunc(mux, "/t/scrapeday", s.handleScrapeday)
	handleErrFunc(mux, "/t/scrapeperson", s.handleScrapeperson)

	log.Printf("listening for requests on %s", s.addr)

	srv := &http.Server{
		Addr: s.addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			log.Printf("%s %s", req.Method, req.URL)
			mux.ServeHTTP(w, req)
		}),
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

func onErr(_ context.Context, err error) {
	log.Print(err.Error())
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

func handle(pattern string, f interface{}) {
	http.Handle(pattern, logHandler{next: hj.Handler(f, logErr)})
}

type logHandler struct {
	next http.Handler
}

func (lh logHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s", req.Method, req.URL)
	lh.next.ServeHTTP(w, req)
}

func logErr(_ context.Context, err error) {
	log.Print(err.Error())
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
func (s *Server) checkCron(req *http.Request) error {
	if !appengine.IsAppEngine() {
		return nil
	}

	h := strings.TrimSpace(req.Header.Get("X-Appengine-Cron"))
	if h != "true" {
		return codeErrType{code: http.StatusUnauthorized}
	}
	return nil
}

// See
// https://cloud.google.com/tasks/docs/creating-appengine-handlers#reading_request_headers.
func (s *Server) checkTaskQueue(req *http.Request, queue string) error {
	if !appengine.IsAppEngine() {
		return nil
	}

	h := strings.TrimSpace(req.Header.Get("X-AppEngine-QueueName"))
	if h != queue {
		return codeErrType{code: http.StatusUnauthorized}
	}
	return nil
}

func (s *Server) numPrinter(n int) string {
	return s.p.Sprintf("%v", n)
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
