package site

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/bobg/hj"
)

func handleErrFunc(mux *http.ServeMux, pattern string, f func(http.ResponseWriter, *http.Request) error) {
	mux.Handle(pattern, errFuncHandler{f: f})
}

type errFuncHandler struct {
	f func(http.ResponseWriter, *http.Request) error
}

func (e errFuncHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := e.f(w, req)
	if err != nil {
		errRespond(w, err)
	}
}

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
		errRespond(w, err)
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

func errRespond(w http.ResponseWriter, err error) {
	if r, ok := err.(hj.Responder); ok {
		r.Respond(w)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
