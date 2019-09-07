package site

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestRlink(t *testing.T) {
	cases := []struct {
		inp, want string
	}{
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/thumb/x/xy/foo",
			want: "https://localhost/r?ct=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/thumb/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/commons/thumb/x/yx/foo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/x/xy/foo",
			want: "https://localhost/r?c=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/commons/x/yx/foo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/thumb/x/xy/foo",
			want: "https://localhost/r?et=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/thumb/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/en/thumb/x/yx/foo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/x/xy/foo",
			want: "https://localhost/r?e=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/en/x/yx/foo",
		},
		{
			inp:  "/wiki/foo",
			want: "https://localhost/r?w=foo",
		},
		{
			inp:  "foo",
			want: "foo",
		},
	}

	req := &http.Request{Host: "localhost", URL: new(url.URL)}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%02d", i), func(t *testing.T) {
			got, _ := rlink(req, c.inp)
			if got.String() != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}
