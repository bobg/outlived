package site

import (
	"context"
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

func sessHandler(next http.Handler) http.Handler {
	return sessHandlerType{next: next}
}

type sessHandlerType struct {
	dsClient *datastore.Client
	next     http.Handler
}

func (s sessHandlerType) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	sess, err := aesite.GetSession(ctx, s.dsClient, req)
	if err != nil {
		errRespond(w, err)
		return
	}
	if sess != nil {
		ctx = context.WithValue(ctx, sessKey{}, sess)
		req = req.WithContext(ctx)
	}
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
	if c, ok := err.(hj.CodeErr); ok {
		c.Respond(w)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
