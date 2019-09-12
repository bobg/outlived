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

	maybeBornDied = regexp.MustCompile(`^\s*\(([^()]+)–([^()]+)\)`)
)

func ScrapeDay(ctx context.Context, client *http.Client, m time.Month, d int, onPerson func(ctx context.Context, href, title, desc string) error) error {
	pageName := fmt.Sprintf("%s_%d", monthName[m], d)
	resp, _, err := getWikiHTML(ctx, client, pageName)
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
		dashNode := findElNode(li, func(n *html.Node) bool {
			if n.DataAtom != atom.Span {
				return false
			}
			return plainTextOf(n) == "–" // dash, not hyphen
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
	resp, updHref, err := getWikiHTML(ctx, client, href)
	if err != nil {
		return errors.Wrapf(err, "getting %s", href)
	}
	defer resp.Body.Close()
	if updHref != href {
		log.Printf("updating href %s -> %s", href, updHref)
		href = updHref
	}

	tree, err := html.Parse(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "parsing HTML of %s", href)
	}

	fullname, imgSrc, imgAlt, bornY, bornM, bornD, diedY, diedM, diedD, err := parsePerson(ctx, tree, href, title)
	if err != nil {
		return errors.Wrapf(err, "parsing content of %s", href)
	}

	if fullname != "" && fullname != title {
		log.Printf("updating %s -> %s", title, fullname)
		title = fullname
		if ind := strings.Index(title, "\n"); ind > 0 {
			title = strings.TrimSpace(title[:ind])
			log.Printf("truncating to %s", title)
		}
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

func parsePerson(ctx context.Context, tree *html.Node, href, title string) (
	fullname, imgSrc, imgAlt string,
	bornY, bornM, bornD, diedY, diedM, diedD int,
	err error,
) {
	infobox := findInfoBox(tree)
	if infobox == nil {
		return parsePersonWithoutInfoBox(ctx, tree, href, title)
	}

	fullname = findFullName(infobox)

	imgSrc, imgAlt = findImg(infobox)

	bornY, bornM, bornD, err = findDateRow(infobox, "Born")
	if err != nil {
		err = errors.Wrap(err, "finding Born row")
		return
	}
	diedY, diedM, diedD, err = findDateRow(infobox, "Died")
	if err != nil {
		err = errors.Wrap(err, "finding Died row")
		return
	}

	return fullname, imgSrc, imgAlt, bornY, bornM, bornD, diedY, diedM, diedD, err
}

func parsePersonWithoutInfoBox(ctx context.Context, tree *html.Node, href, title string) (
	fullname, imgSrc, imgAlt string,
	bornY, bornM, bornD, diedY, diedM, diedD int,
	err error,
) {
	secNode := findElNode(tree, func(n *html.Node) bool {
		return n.DataAtom == atom.Section
	})
	if secNode == nil {
		err = fmt.Errorf("no infobox and no section node in %s", href)
		return
	}

	// Look for the first <p> under secNode.
	// Check it for name and dates.
	// If that's no good, look for the second <p> and check that.
	var (
		pNode *html.Node
		tries int
		found bool
	)
	for pNode = secNode.FirstChild; pNode != nil && tries <= 2; pNode = pNode.NextSibling {
		if pNode.Type != html.ElementNode || pNode.DataAtom != atom.P {
			continue
		}

		tries++

		bNode := findElNode(pNode, func(n *html.Node) bool {
			return n.DataAtom == atom.B
		})
		if bNode == nil {
			continue
		}
		fullname = plainTextOf(bNode)
		if fullname != title {
			continue
		}

		buf := new(bytes.Buffer)
		for tNode := bNode.NextSibling; tNode != nil; tNode = tNode.NextSibling {
			toPlainText(buf, tNode)
		}
		tNodeText := buf.String()
		m := maybeBornDied.FindStringSubmatch(tNodeText)
		if m == nil {
			continue
		}

		bornY, bornM, bornD, err = parseDate(m[1])
		if err != nil {
			continue
		}

		diedY, diedM, diedD, err = parseDate(m[2])
		if err != nil {
			continue
		}

		found = true
		break
	}

	if !found {
		err = fmt.Errorf("no infobox and no suitable intro text in %s", href)
		return
	}

	// Look for the first <figure> under secNode.
	// Look for an <img> inside that, and perhaps also a <figcaption> (for the imgAlt).
	figNode := findElNode(secNode, func(n *html.Node) bool {
		return n.DataAtom == atom.Figure
	})
	if figNode == nil {
		return
	}
	imgSrc, imgAlt = findImg(figNode)
	if imgSrc == "" {
		return
	}
	captionEl := findNode(figNode, func(n *html.Node) bool {
		return n.DataAtom == atom.Figcaption
	})
	if captionEl == nil {
		return
	}
	imgAlt = plainTextOf(captionEl)
	return
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
	fnNode := findElNode(node, func(n *html.Node) bool {
		return elClassContains(n, "fn")
	})
	if fnNode == nil {
		return ""
	}
	return strings.TrimSpace(plainTextOf(fnNode))
}

func findDateRow(node *html.Node, label string) (year, mon, day int, err error) {
	if node.Type == html.ElementNode && node.DataAtom == atom.Th {
		if plainTextOf(node) != label {
			return 0, 0, 0, errNotFound
		}
		td := node.NextSibling
		if td == nil || td.Type != html.ElementNode || td.DataAtom != atom.Td {
			return 0, 0, 0, errNotFound
		}
		return parseDate(plainTextOf(td))
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
	h2El := findElNode(node, func(n *html.Node) bool {
		return n.DataAtom == atom.H2 && elAttr(n, "id") == "Deaths"
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

func findElNode(node *html.Node, pred func(*html.Node) bool) *html.Node {
	return findNode(node, func(n *html.Node) bool {
		return n.Type == html.ElementNode && pred(n)
	})
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
	return findElNode(node, func(n *html.Node) bool {
		return n.DataAtom == atom.Table && elClassContains(n, "infobox")
	})
}

func findImg(node *html.Node) (src, alt string) {
	found := findElNode(node, func(n *html.Node) bool {
		return n.DataAtom == atom.Img
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

func plainTextOf(node *html.Node) string {
	buf := new(bytes.Buffer)
	toPlainText(buf, node)
	return buf.String()
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
	req.Header.Set("User-Agent", "Outlived/1.0")
	return client.Do(req)
}

func getWikiHTML(ctx context.Context, client *http.Client, name string) (*http.Response, string, error) {
	const prefix = "https://en.wikipedia.org/api/rest_v1/page/html/"

	resp, err := httpGetContext(ctx, client, prefix+name)
	if err != nil {
		return nil, "", err
	}
	loc := resp.Header.Get("Content-Location")
	if newName := strings.TrimPrefix(loc, prefix); newName != loc {
		name = newName
	}
	return resp, name, nil
}
