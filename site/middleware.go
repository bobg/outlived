package site

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/bobg/mid"
)

func (s *Server) sessHandler(next http.Handler) http.Handler {
	return sessHandlerType{next: next, dsClient: s.dsClient}
}

type sessHandlerType struct {
	dsClient *datastore.Client
	next     http.Handler
}

func (s sessHandlerType) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err == http.ErrNoCookie {
		log.Print("no session cookie in HTTP request")
	} else if err == datastore.ErrNoSuchEntity {
		log.Print("session cookie not found in datastore, skipping")
	} else if err == aesite.ErrInactive {
		log.Print("found inactive session, skipping")
	} else if err != nil {
		mid.Errf(w, 0, "%s", err)
		return
	}
	ctx = context.WithValue(ctx, sessKey{}, sess)
	req = req.WithContext(ctx)
	s.next.ServeHTTP(w, req)
}

type sessKey struct{}

func getSess(ctx context.Context) *aesite.Session {
	val := ctx.Value(sessKey{})
	if val != nil {
		return val.(*aesite.Session)
	}
	return nil
}
