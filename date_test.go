package outlived

import (
	"fmt"
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	type yd struct{ y, d int }

	cases := []struct {
		y1, m1, d1 int
		y2, m2, d2 int
		wantD      int
		wantYD     *yd
	}{
		{2000, 1, 1, 2000, 1, 1, 0, &yd{0, 0}},
		{2000, 1, 1, 2000, 1, 2, 1, &yd{0, 1}},
		{2000, 1, 2, 2000, 1, 1, -1, nil},
		{2000, 1, 1, 2000, 2, 1, 31, &yd{0, 31}},
		{2000, 1, 1, 2000, 3, 1, 60, &yd{0, 60}},
		{1999, 12, 31, 2000, 1, 1, 1, &yd{0, 1}},
		{2003, 1, 1, 2004, 1, 1, 365, &yd{1, 0}},
		{2003, 1, 1, 2005, 1, 1, 731, &yd{2, 0}},
		{1900, 2, 1, 1900, 3, 1, 28, &yd{0, 28}},
		{1904, 2, 1, 1904, 3, 1, 29, &yd{0, 29}},
		{2000, 2, 1, 2000, 3, 1, 29, &yd{0, 29}},
		{1966, 10, 22, 1969, 10, 28, 1102, &yd{3, 6}},
		{1966, 10, 22, 1977, 8, 5, 3940, &yd{10, 286}},
		{2004, 4, 10, 2015, 1, 23, 3940, &yd{10, 288}},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case_%d", i+1), func(t *testing.T) {
			d1 := Date{Y: c.y1, M: time.Month(c.m1), D: c.d1}
			d2 := Date{Y: c.y2, M: time.Month(c.m2), D: c.d2}
			delta := d2.Since(d1)
			if delta != c.wantD {
				t.Errorf("got %d, want %d", delta, c.wantD)
			}
			if c.wantYD == nil {
				return
			}
			y, d := d2.YDSince(d1)
			if y != c.wantYD.y || d != c.wantYD.d {
				t.Errorf("got yd %d,%d, want %d,%d", y, d, c.wantYD.y, c.wantYD.d)
			}
		})
	}
}
