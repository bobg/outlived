package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/time/rate"
)

var (
	monthName = []string{
		"",
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}

	dateRegex1 = regexp.MustCompile(`(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d+),\s+(\d+)(\s+BC)?`)
	dateRegex2 = regexp.MustCompile(`(\d+)\s+(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d+)(\s+BC)?`)

	daysInMonth = []int{
		0,
		31,
		29,
		31,
		30,
		31,
		30,
		31,
		31,
		30,
		31,
		30,
		31,
	}

	bRegex = regexp.MustCompile(`\(b\.\s*\d+\)$`)
)

func scrape(ctx context.Context, startM, startD int, dur time.Duration) error {
	first := true
	lim := rate.NewLimiter(rate.Every(dur))
	for m := startM; m <= 12; m++ {
		for d := 1; d <= daysInMonth[m]; d++ {
			if first {
				d = startD
				first = false
			}

			err = scrapeDay(ctx, m, d, lim)
			if err != nil {
				return errors.Wrapf(err, "getting date %d/%d", m, d)
			}
		}
	}
}

func scrapeDay(ctx context.Context, m, d int, lim *rate.Limiter) error {
	url := fmt.Sprintf("https://en.wikipedia.org/wiki/%s_%d", monthName[m], d)
	err := lim.Wait(ctx)
	if err != nil {
		return errors.Wrapf(err, "waiting to get %s", url)
	}
	log.Printf("getting %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "getting %s", url)
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "parsing %s", url)
	}

	deaths := findDeathsUL(tree)
	if deaths == nil {
		return fmt.Errorf("no Deaths node in %s", url)
	}

	for li := deaths.FirstChild; li != nil; li = li.NextSibling {
		// Find the <a> node past a plain text node containing "–" (that's a dash, not a hyphen).
		var aNode *html.Node
		for node := li.FirstChild; node != nil; node = node.NextSibling {
			if node.Type != html.TextNode {
				continue
			}
			if !strings.Contains(node.Data, "–") {
				continue
			}
			if next := node.NextSibling; next != nil && next.Type == html.ElementNode && next.DataAtom == atom.A {
				aNode = next
				break
			}
		}
		if aNode == nil {
			continue // xxx log a warning?
		}

		href := elAttr(aNode, "href")
		title := elAttr(aNode, "title")

		b := new(bytes.Buffer)

		for node := aNode.NextSibling; node != nil; node = node.NextSibling {
			if node.Type == html.ElementNode && node.DataAtom == atom.Sup {
				break
			}
			toPlainText(b, node)
		}

		desc := b.String()
		desc = strings.TrimPrefix(desc, ",")
		if found := bRegex.FindStringIndex(desc); found != nil {
			desc = desc[:found[0]]
		}
		desc = strings.TrimSpace(desc)

		// log.Printf("parsed %s [%s] %s", title, href, desc)

		err = scrapePerson(ctx, href, title, desc, lim)
		if err != nil {
			log.Printf("scraping person %s: %s", title, err)
		}
	}

	return nil
}

func scrapePerson(ctx context.Context, href, title, desc string, lim *rate.Limiter) error {
	url := "https://en.wikipedia.org" + href
	err := lim.Wait(ctx)
	if err != nil {
		return errors.Wrapf(err, "waiting to scrape %s", url)
	}
	log.Printf("getting %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "getting %s", url)
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "parsing %s", url)
	}

	infobox := findInfoBox(tree)
	if infobox == nil {
		return fmt.Errorf("no infobox in %s", url)
	}

	bornY, bornM, bornD, err := findDateRow(infobox, "Born")
	if err != nil {
		return errors.Wrap(err, "finding Born row")
	}
	diedY, diedM, diedD, err := findDateRow(infobox, "Died")
	if err != nil {
		return errors.Wrap(err, "finding Died row")
	}

	by := bornY
	if by < 0 {
		by++
	}
	dy := diedY
	if dy < 0 {
		dy++
	}
	born := time.Date(by, time.Month(bornM), bornD, 0, 0, 0, 0, time.UTC)
	died := time.Date(dy, time.Month(diedM), diedD, 0, 0, 0, 0, time.UTC)
	aliveDur := died.Sub(born)
	aliveDays := int(aliveDur / (24 * time.Hour))

	bornStr := fmt.Sprintf("%d-%02d-%02d", bornY, bornM, bornD)
	diedStr := fmt.Sprintf("%d-%02d-%02d", diedY, diedM, diedD)

	// xxx
	// out.Write([]string{title, desc, bornStr, diedStr, strconv.Itoa(aliveDays), href})
	// log.Printf("got [%s], [%s], [%d-%02d-%02d]-[%d-%02d-%02d] [%d days] from [%s]", title, desc, bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, href)

	return nil
}

var errNotFound = errors.New("not found")

func findDateRow(node *html.Node, label string) (year, mon, day int, err error) {
	if node.Type == html.ElementNode && node.DataAtom == atom.Th {
		b := new(bytes.Buffer)
		toPlainText(b, node)
		if b.String() != label {
			return 0, 0, 0, errNotFound
		}
		td := node.NextSibling
		if td == nil || td.Type != html.ElementNode || td.DataAtom != atom.Td {
			return 0, 0, 0, errNotFound
		}
		b.Reset()
		toPlainText(b, td)
		return parseDate(b.String())
	}
	if node.Type == html.TextNode {
		return 0, 0, 0, errNotFound
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if year, mon, day, err = findDateRow(child, label); err == nil {
			return
		}
	}
	return 0, 0, 0, errNotFound
}

func parseDate(s string) (year, mon, day int, err error) {
	if m := dateRegex1.FindStringSubmatch(s); m != nil {
		return parseDate2(m[3], m[1], m[2], m[4])
	}
	if m := dateRegex2.FindStringSubmatch(s); m != nil {
		return parseDate2(m[3], m[2], m[1], m[4])
	}
	return 0, 0, 0, errNotFound
}

func parseDate2(yearStr, monStr, dayStr, bcStr string) (year, mon, day int, err error) {
	for i := 1; i <= 12; i++ {
		if monStr == monthName[i] {
			mon = i
			break
		}
	}
	if mon == 0 {
		return 0, 0, 0, errors.Wrap(err, "parsing month")
	}
	day, err = strconv.Atoi(dayStr)
	if err != nil { // xxx also range check?
		return 0, 0, 0, errors.Wrap(err, "parsing day")
	}
	year, err = strconv.Atoi(yearStr) // xxx also range check?
	if err != nil {
		return 0, 0, 0, errors.Wrap(err, "parsing year")
	}
	if bcStr != "" {
		year = -year
	}
	return
}

func findDeathsUL(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.DataAtom == atom.H2 {
		span := node.FirstChild
		if span == nil {
			return nil
		}
		if span.Type != html.ElementNode {
			return nil
		}
		if span.DataAtom != atom.Span {
			return nil
		}
		if id := elAttr(span, "id"); id != "Deaths" {
			return nil
		}

		for sib := node.NextSibling; sib != nil; sib = sib.NextSibling {
			if sib.Type == html.ElementNode && sib.DataAtom == atom.Ul {
				return sib
			}
		}

		return nil
	}
	if node.Type == html.TextNode {
		return nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findDeathsUL(child); result != nil {
			return result
		}
	}
	return nil
}

func findInfoBox(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.DataAtom == atom.Table && elClassContains(node, "infobox") {
		return node
	}
	if node.Type == html.TextNode {
		return nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if result := findInfoBox(child); result != nil {
			return result
		}
	}
	return nil
}

func toPlainText(w io.Writer, node *html.Node) {
	switch node.Type {
	case html.TextNode:
		w.Write([]byte(node.Data))

	case html.ElementNode:
		if node.DataAtom == atom.Br {
			w.Write([]byte("\n"))
			return
		}
		for subnode := node.FirstChild; subnode != nil; subnode = subnode.NextSibling {
			toPlainText(w, subnode)
		}
	}
}

func elAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func elClassContains(node *html.Node, probe string) bool {
	classes := strings.Fields(elAttr(node, "class"))
	for _, c := range classes {
		if c == probe {
			return true
		}
	}
	return false
}
