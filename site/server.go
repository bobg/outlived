package site

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
)

func NewServer(ctx context.Context, addr, smtpAddr, contentDir string, dsClient *datastore.Client, ctClient *cloudtasks.Client) *Server {
	s := &Server{
		addr:       addr,
		smtpAddr:   smtpAddr,
		contentDir: contentDir,
		dsClient:   dsClient,
		ctClient:   ctClient,
	}
	go s.scrape(ctx)
	return s
}

type Server struct {
	addr       string
	smtpAddr   string
	contentDir string
	dsClient   *datastore.Client
	ctClient   *cloudtasks.Client
	sender     sender
}

type sender interface {
	send(ctx context.Context, from string, to []string, subject string, body io.Reader) error
}

func (s *Server) Serve(ctx context.Context) {
	handle("/", s.handleHome)
	handle("/load", s.handleLoad)
	handle("/login", s.handleLogin)
	handle("/logout", s.handleLogout)
	handle("/signup", s.handleSignup)
	handle("/verify", s.handleVerify)
	handle("/scrapeday", s.handleScrapeday)
	handle("/scrapeperson", s.handleScrapeperson)
	http.HandleFunc("/js/", s.handleStatic)
	http.HandleFunc("/css/", s.handleStatic)

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
		err := f(w, req)
		if err != nil {
			code := http.StatusInternalServerError
			if err, ok := err.(codeErrType); ok {
				code = err.code
			}
			log.Printf("%s", err)
			http.Error(w, err.Error(), code)
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
