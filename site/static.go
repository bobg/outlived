package site

import (
	"net/http"
	"path/filepath"
)

func (s *Server) handleStatic(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, filepath.Join(s.contentDir, req.URL.Path))
}
