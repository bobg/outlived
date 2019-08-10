package site

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
)

func NewServer(addr, smtpAddr, contentDir string, dsClient *datastore.Client) *Server {
	return &Server{
		addr:       addr,
		smtpAddr:   smtpAddr,
		contentDir: contentDir,
		dsClient:   dsClient,
	}
}

type Server struct {
	addr       string
	smtpAddr   string
	contentDir string
	dsClient   *datastore.Client
	sender     sender
}

type sender interface {
	send(ctx context.Context, from string, to []string, subject string, body io.Reader) error
}

func (s *Server) Serve(ctx context.Context) {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/load", s.handleLoad)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/verify", s.handleVerify)

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
