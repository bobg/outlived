package outlived

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Date struct {
	Y int
	M time.Month
	D int
}

var dateRegex = regexp.MustCompile(`(\d+)-(\d+)-(\d+)`)

var errDateParse = errors.New("bad dates")

// Parses dates of the form "yyyy-mm-dd."
func ParseDate(s string) (Date, error) {
	m := dateRegex.FindStringSubmatch(s)
	if len(m) != 4 {
		return Date{}, errDateParse
	}
	y, err := strconv.Atoi(m[1])
	if err != nil {
		return Date{}, errDateParse
	}
	mon, err := strconv.Atoi(m[2])
	if err != nil {
		return Date{}, errDateParse
	}
	d, err := strconv.Atoi(m[3])
	if err != nil {
		return Date{}, errDateParse
	}
	if y <= 0 {
		return Date{}, errDateParse
	}
	if mon < 1 || mon > 12 {
		return Date{}, errDateParse
	}
	if d < 1 || d > daysInMonth(y, time.Month(mon)) {
		return Date{}, errDateParse
	}
	return Date{Y: y, M: time.Month(mon), D: d}, nil
}

func Today(loc *time.Location) Date {
	now := time.Now().In(loc)
	return TimeDate(now)
}

func TimeDate(t time.Time) Date {
	return Date{Y: t.Year(), M: t.Month(), D: t.Day()}
}

func (d Date) Since(other Date) int {
	t1 := time.Date(d.Y, d.M, d.D, 0, 0, 0, 0, time.Local)
	t2 := time.Date(other.Y, other.M, other.D, 0, 0, 0, 0, time.Local)
	return int(t1.Sub(t2) / (24 * time.Hour))
}

// Returns years and days from other to d.
// Note: other must be on or before d.
func (d Date) YDSince(other Date) (years, days int) {
	years = d.Y - other.Y
	if other.M > d.M || (other.M == d.M && other.D > d.D) {
		years--
		days = 1 + Date{other.Y, 12, 31}.Since(other) + d.Since(Date{d.Y, 1, 1})
	} else {
		days = d.Since(Date{d.Y, other.M, other.D})
	}
	return years, days
}

func (d Date) YDSinceStr(other Date) string {
	years, days := d.YDSince(other)
	if years == 0 && days == 0 {
		return "0 days"
	}
	if years == 0 && days == 1 {
		return "1 day"
	}
	if years == 0 && days > 1 {
		return fmt.Sprintf("%d days", days)
	}
	if years == 1 && days == 0 {
		return "1 year"
	}
	if years == 1 && days == 1 {
		return "1 year, 1 day"
	}
	if years == 1 && days > 1 {
		return fmt.Sprintf("1 year, %d days", days)
	}
	if years > 1 && days == 0 {
		return fmt.Sprintf("%d years", years)
	}
	if years > 1 && days == 1 {
		return fmt.Sprintf("%d years, 1 day", years)
	}
	return fmt.Sprintf("%d years, %d days", years, days)
}

func (d Date) String() string {
	m := d.M.String()
	return fmt.Sprintf("%d %s %d", d.D, m[:3], d.Y)
}

func (d Date) YYYYMMDD() string {
	return fmt.Sprintf("%d-%02d-%02d", d.Y, d.M, d.D)
}

func daysInMonth(y int, m time.Month) int {
	switch m {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	}
	if isLeapYear(y) {
		return 29
	}
	return 28
}

func isLeapYear(y int) bool {
	if y%400 == 0 {
		return true
	}
	if y%100 == 0 {
		return false
	}
	return y%4 == 0
}
