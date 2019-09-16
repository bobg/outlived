package site

import (
	"net/http"
	"path/filepath"
)

func (s *Server) handleStatic(w http.ResponseWriter, req *http.Request) error {
	path := req.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	http.ServeFile(w, req, filepath.Join(s.contentDir, path))
	return nil
}
