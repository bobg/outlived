package outlived

import (
	"fmt"
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	cases := []struct {
		y1, m1, d1 int
		y2, m2, d2 int
		want       int
	}{
		{2000, 1, 1, 2000, 1, 1, 0},
		{2000, 1, 1, 2000, 1, 2, 1},
		{2000, 1, 2, 2000, 1, 1, -1},
		{2000, 1, 1, 2000, 2, 1, 31},
		{2000, 1, 1, 2000, 3, 1, 60},
		{1999, 12, 31, 2000, 1, 1, 1},
		{2003, 1, 1, 2004, 1, 1, 365},
		{2003, 1, 1, 2005, 1, 1, 731},
		{1900, 2, 1, 1900, 3, 1, 28},
		{1904, 2, 1, 1904, 3, 1, 29},
		{2000, 2, 1, 2000, 3, 1, 29},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%d", i+1), func(t *testing.T) {
			d1 := Date{Y: c.y1, M: time.Month(c.m1), D: c.d1}
			d2 := Date{Y: c.y2, M: time.Month(c.m2), D: c.d2}
			delta := d2.Since(d1)
			if delta != c.want {
				t.Errorf("got %d, want %d", delta, c.want)
			}
		})
	}
}
