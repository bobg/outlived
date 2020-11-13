package site

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bobg/mid"
)

func (s *Server) handleRedirect(w http.ResponseWriter, req *http.Request) error {
	redirect := func(f string, args ...interface{}) {
		http.Redirect(w, req, fmt.Sprintf(f, args...), http.StatusMovedPermanently)
	}

	if ct := req.FormValue("ct"); ct != "" {
		first := ct[0]
		redirect("https://upload.wikimedia.org/wikipedia/commons/thumb/%c/%s", first, ct)
		return nil
	}

	if c := req.FormValue("c"); c != "" {
		first := c[0]
		redirect("https://upload.wikimedia.org/wikipedia/commons/%c/%s", first, c)
		return nil
	}

	if et := req.FormValue("et"); et != "" {
		first := et[0]
		redirect("https://upload.wikimedia.org/wikipedia/en/thumb/%c/%s", first, et)
		return nil
	}

	if e := req.FormValue("e"); e != "" {
		first := e[0]
		redirect("https://upload.wikimedia.org/wikipedia/en/%c/%s", first, e)
		return nil
	}

	if w := req.FormValue("w"); w != "" {
		redirect("https://en.wikipedia.org/wiki/%s", w)
		return nil
	}

	return mid.CodeErr{C: http.StatusBadRequest}
}

func rlink(target string) (*url.URL, error) {
	r := &url.URL{
		Path: "/r",
	}

	v := make(url.Values)

	try := func(prefix string) string {
		rest := strings.TrimPrefix(target, prefix)
		if rest == target {
			return ""
		}

		// expect rest to match ^(.)/\1./.+
		if len(rest) < 6 {
			return ""
		}
		if rest[0] != rest[2] {
			return ""
		}
		if rest[1] != '/' || rest[4] != '/' {
			return ""
		}

		return rest[2:]
	}

	if strings.HasPrefix(target, "http:") || strings.HasPrefix(target, "https:") {
		return url.Parse(target)
	}

	if rest := try("//upload.wikimedia.org/wikipedia/commons/thumb/"); rest != "" {
		v.Set("ct", rest)
	} else if rest := try("//upload.wikimedia.org/wikipedia/commons/"); rest != "" {
		v.Set("c", rest)
	} else if rest := try("//upload.wikimedia.org/wikipedia/en/thumb/"); rest != "" {
		v.Set("et", rest)
	} else if rest := try("//upload.wikimedia.org/wikipedia/en/"); rest != "" {
		v.Set("e", rest)
	} else {
		v.Set("w", strings.TrimPrefix(target, "/wiki/"))
	}

	r.RawQuery = v.Encode()
	return homeURL.ResolveReference(r), nil
}
