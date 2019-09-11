package outlived

import (
	"bytes"
	"context"
	"encoding/json"
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

	dateRegex1 = regexp.MustCompile(`(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d+),\s+(\d+)(.*BC)?`)
	dateRegex2 = regexp.MustCompile(`(\d+)\s+(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d+)(.*BC)?`)

	bRegex = regexp.MustCompile(`\(b\.\s*\d+\)$`)

	paren = regexp.MustCompile(`^(.*\S)\s*\([^()]*\)$`)
)

func ScrapeDay(ctx context.Context, client *http.Client, m time.Month, d int, onPerson func(ctx context.Context, href, title, desc string) error) error {
	pageName := fmt.Sprintf("%s_%d", monthName[m], d)
	resp, err := getWikiHTML(ctx, client, pageName)
	if err != nil {
		return errors.Wrapf(err, "getting %s", pageName)
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "parsing %s", pageName)
	}

	deaths := findDeathsUL(tree)
	if deaths == nil {
		return fmt.Errorf("no Deaths node in %s", pageName)
	}

	for li := deaths.FirstChild; li != nil; li = li.NextSibling {
		if err = ctx.Err(); err != nil {
			return err
		}

		if li.Type != html.ElementNode || li.DataAtom != atom.Li {
			continue
		}

		// Find the <span> containing just a "–" (that's a dash [0x2013], not a hyphen).
		dashNode := findNode(li, func(n *html.Node) bool {
			if n.Type != html.ElementNode || n.DataAtom != atom.Span {
				return false
			}
			b := new(bytes.Buffer)
			toPlainText(b, n)
			return b.String() == "–"
		})
		if dashNode == nil {
			continue
		}

		// Find the first sibling of dashNode that's an <a>.
		var aNode *html.Node
		for sib := dashNode.NextSibling; sib != nil; sib = sib.NextSibling {
			if sib.Type != html.ElementNode {
				continue
			}
			if sib.DataAtom != atom.A {
				continue
			}
			aNode = sib
			break
		}
		if aNode == nil {
			continue
		}

		href := elAttr(aNode, "href")
		href = strings.TrimPrefix(href, "./")

		title := elAttr(aNode, "title")
		title = paren.ReplaceAllString(title, "$1")

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

		err = onPerson(ctx, href, title, desc)
		if err != nil {
			return errors.Wrapf(err, "on person %s (%s)", title, href)
		}
	}

	return nil
}

func ScrapePerson(
	ctx context.Context,
	client *http.Client,
	href, title, desc string,
	onPerson func(
		ctx context.Context,
		title, desc, href, imgSrc, imgAlt string,
		bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, pageviews int,
	) error,
) error {
	resp, err := getWikiHTML(ctx, client, href)
	if err != nil {
		return errors.Wrapf(err, "getting %s", href)
	}
	defer resp.Body.Close()

	tree, err := html.Parse(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "parsing %s", href)
	}

	infobox := findInfoBox(tree)
	if infobox == nil {
		return fmt.Errorf("no infobox in %s", href)
	}

	if fullname := findFullName(infobox); fullname != "" {
		if fullname != title {
			log.Printf("updating %s -> %s", title, fullname)
		}
		title = fullname
	} else {
		log.Printf("no fullname for %s in %s", title, href)
	}

	if ind := strings.Index(title, "\n"); ind > 0 {
		title = title[:ind]
		log.Printf("truncating to %s", title)
	}

	title = strings.TrimSpace(title)

	imgSrc, imgAlt := findImg(infobox)

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

	pageviews, err := scrapePageviews(ctx, client, href)
	if err != nil {
		return errors.Wrap(err, "getting pageviews")
	}

	return onPerson(ctx, title, desc, href, imgSrc, imgAlt, bornY, bornM, bornD, diedY, diedM, diedD, aliveDays, pageviews)
}

var nameRegex = regexp.MustCompile(`[^/]+$`)

func scrapePageviews(ctx context.Context, client *http.Client, href string) (int, error) {
	m := nameRegex.FindString(href)
	if m == "" {
		return 0, fmt.Errorf("could not construct pageviews link from href %s", href)
	}

	var (
		now       = time.Now()
		yesterday = now.Add(-24 * time.Hour)
		start     = yesterday.Add(-90 * 24 * time.Hour) // 90 days before yesterday
	)

	// C.f. https://wikimedia.org/api/rest_v1/
	u := fmt.Sprintf("https://wikimedia.org/api/rest_v1/metrics/pageviews/per-article/en.wikipedia.org/all-access/user/%s/daily/%d%02d%02d00/%d%02d%02d00", m, start.Year(), start.Month(), start.Day(), yesterday.Year(), yesterday.Month(), yesterday.Day())
	resp, err := httpGetContext(ctx, client, u)
	if err != nil {
		return 0, errors.Wrapf(err, "fetching %s", u)
	}
	defer resp.Body.Close()

	type (
		respItemType struct {
			Views int
		}

		respType struct {
			Items []*respItemType
		}
	)

	var parsed respType
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&parsed)
	if err != nil {
		return 0, errors.Wrapf(err, "parsing response from %s", u)
	}

	var count int
	for _, item := range parsed.Items {
		count += item.Views
	}

	return count, nil
}

var errNotFound = errors.New("not found")

func findFullName(node *html.Node) string {
	fnNode := findNode(node, func(n *html.Node) bool {
		return elClassContains(n, "fn")
	})
	if fnNode == nil {
		return ""
	}
	buf := new(bytes.Buffer)
	toPlainText(buf, fnNode)
	return strings.TrimSpace(buf.String())
}

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
	if err != nil {
		return 0, 0, 0, errors.Wrap(err, "parsing day")
	}
	year, err = strconv.Atoi(yearStr) // TODO: range check?
	if err != nil {
		return 0, 0, 0, errors.Wrap(err, "parsing year")
	}
	if bcStr != "" {
		year = -year
	}
	if day < 1 || day > daysInMonth(year, time.Month(mon)) {
		return 0, 0, 0, fmt.Errorf("day %d out of range", day)
	}
	return
}

func findDeathsUL(node *html.Node) *html.Node {
	h2El := findNode(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.H2 && elAttr(n, "id") == "Deaths"
	})
	if h2El == nil {
		return nil
	}
	for sib := h2El.NextSibling; sib != nil; sib = sib.NextSibling {
		if sib.Type == html.ElementNode && sib.DataAtom == atom.Ul {
			return sib
		}
	}
	return nil
}

func findNode(node *html.Node, pred func(*html.Node) bool) *html.Node {
	if pred(node) {
		return node
	}
	if node.Type == html.TextNode {
		return nil
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findNode(child, pred); found != nil {
			return found
		}
	}
	return nil
}

func findInfoBox(node *html.Node) *html.Node {
	return findNode(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.Table && elClassContains(n, "infobox")
	})
}

func findImg(node *html.Node) (src, alt string) {
	found := findNode(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.Img
	})
	if found != nil {
		return elAttr(found, "src"), elAttr(found, "alt")
	}
	return "", ""
}

func toPlainText(w io.Writer, node *html.Node) {
	switch node.Type {
	case html.TextNode:
		w.Write([]byte(html.UnescapeString(node.Data)))

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

func httpGetContext(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return client.Do(req)
}

func getWikiHTML(ctx context.Context, client *http.Client, name string) (*http.Response, error) {
	return httpGetContext(ctx, client, fmt.Sprintf("https://en.wikipedia.org/api/rest_v1/page/html/%s?redirect=false", name))
}
