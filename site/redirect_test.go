package site

import (
	"fmt"
	"testing"
)

func TestRlink(t *testing.T) {
	cases := []struct {
		inp, want string
	}{
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/thumb/x/xy/foo",
			want: "/r?ct=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/thumb/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/commons/thumb/x/yx/foo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/x/xy/foo",
			want: "/r?c=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/commons/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/commons/x/yx/foo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/thumb/x/xy/foo",
			want: "/r?et=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/thumb/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/en/thumb/x/yx/foo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/x/xy/foo",
			want: "/r?e=xy%2Ffoo",
		},
		{
			inp:  "//upload.wikimedia.org/wikipedia/en/x/yx/foo",
			want: "//upload.wikimedia.org/wikipedia/en/x/yx/foo",
		},
		{
			inp:  "/wiki/foo",
			want: "/r?w=foo",
		},
		{
			inp:  "foo",
			want: "foo",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%02d", i), func(t *testing.T) {
			u, _ := rlink(c.inp)
			u.Scheme = ""
			u.Host = ""
			got := u.String()
			if got != c.want {
				t.Errorf("got %s, want %s", got, c.want)
			}
		})
	}
}
