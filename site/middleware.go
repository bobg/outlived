package site

import "net/http"

func handleErrFunc(mux *http.ServeMux, pattern string, f func(http.ResponseWriter, *http.Request) error) {
	mux.Handle(pattern, errFuncHandler{f: f})
}

type errFuncHandler struct {
	f func(http.ResponseWriter, *http.Request)
}

func (e errFuncHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	err := e.f(w, req)
	if err != nil {
		// xxx
	}
}
