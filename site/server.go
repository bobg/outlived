package site

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
)

func NewServer(addr string, dsClient *datastore.Client, smtpAddr string) *Server {
	return &Server{
		addr:     addr,
		dsClient: dsClient,
		smtpAddr: smtpAddr,
	}
}

type Server struct {
	addr     string
	dsClient *datastore.Client
	smtpAddr string
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
