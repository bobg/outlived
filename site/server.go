package site

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
)

func NewServer(addr, smtpAddr string, dsClient *datastore.Client) *Server {
	return &Server{
		addr:     addr,
		smtpAddr: smtpAddr,
		dsClient: dsClient,
	}
}

type Server struct {
	addr     string
	smtpAddr string
	dsClient *datastore.Client
}

func (s *Server) Serve(ctx context.Context) {
	http.HandleFunc("/", s.handleHome)
	http.HandleFunc("/load", s.handleLoad)
	http.HandleFunc("/signup", s.handleSignup)
	http.HandleFunc("/verify", s.handleVerify)

	log.Printf("listening for requests on %s", s.addr)

	srv := &http.Server{Addr: *addr}
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
