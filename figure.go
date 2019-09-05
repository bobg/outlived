package outlived

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

// Figure is a historical figure.
type Figure struct {
	// Link is the path part of the figure's Wikipedia URL.
	// This also serves as the figure's unique datastore key.
	Link string

	Name, Desc     string
	Born, Died     Date
	DaysAlive      int
	Pageviews      int
	ImgSrc, ImgAlt string

	Updated time.Time
}

func (f *Figure) YDAge() string {
	y, d := f.Died.YDSince(f.Born)

	var dstr string
	if d == 1 {
		dstr = fmt.Sprintf("%d day", d)
	} else {
		dstr = fmt.Sprintf("%d days", d)
	}

	if y == 0 {
		return dstr
	}

	var ystr string
	if y == 1 {
		ystr = fmt.Sprintf("%d year", y)
	} else {
		ystr = fmt.Sprintf("%d years", y)
	}

	return fmt.Sprintf("%s, %s", ystr, dstr)
}

func FiguresAliveFor(ctx context.Context, client *datastore.Client, days, limit int) ([]*Figure, error) {
	q := datastore.NewQuery("Figure").Filter("DaysAlive =", days).Order("-Pageviews")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var figures []*Figure
	_, err := client.GetAll(ctx, q, &figures)
	return figures, errors.Wrap(err, "querying figures")
}

func FiguresAliveForAtMost(ctx context.Context, client *datastore.Client, days, limit int) ([]*Figure, error) {
	q := datastore.NewQuery("Figure").Filter("DaysAlive <=", days).Order("-DaysAlive").Order("-Pageviews")
	it := client.Run(ctx, q)
	var figures []*Figure
	for len(figures) < limit {
		var fig Figure
		_, err := it.Next(&fig)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "iterating")
		}
		figures = append(figures, &fig)
	}
	sort.Slice(figures, func(i, j int) bool {
		return figures[i].Pageviews > figures[j].Pageviews
	})
	return figures, nil
}

func FiguresDiedOn(ctx context.Context, client *datastore.Client, mon time.Month, day int, limit int) ([]*Figure, error) {
	q := datastore.NewQuery("Figure").Filter("Died.M =", int(mon)).Filter("Died.D =", day).Order("-Pageviews")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var figures []*Figure
	_, err := client.GetAll(ctx, q, &figures)
	return figures, errors.Wrap(err, "querying figures")
}

const multiLimit = 500

func ReplaceFigures(ctx context.Context, client *datastore.Client, figures []*Figure) error {
	beforeDeduping := len(figures)

	// Remove duplicates from figures.
	var (
		seen    = make(map[string]struct{})
		deduped []*Figure
	)
	for _, fig := range figures {
		if _, ok := seen[fig.Link]; ok {
			continue
		}
		seen[fig.Link] = struct{}{}
		deduped = append(deduped, fig)
	}
	figures = deduped

	afterDeduping := len(figures)

	// TODO(bobg): At least in testing mode, this call to Count (apparently) never returns.
	// before, err := client.Count(ctx, allQ)
	// if err != nil {
	// 	return errors.Wrap(err, "counting figures before replace")
	// }

	keys := make([]*datastore.Key, len(figures))
	for i, fig := range figures {
		keys[i] = &datastore.Key{Kind: "Figure", Name: fig.Link}
	}
	for len(figures) > 0 {
		var (
			nextKeys []*datastore.Key
			nextFigs []*Figure
		)
		if len(figures) > multiLimit {
			keys, nextKeys = keys[:multiLimit], keys[multiLimit:]
			figures, nextFigs = figures[:multiLimit], figures[multiLimit:]
		}
		_, err := client.PutMulti(ctx, keys, figures)
		if err != nil {
			return errors.Wrap(err, "storing figures")
		}
		keys, figures = nextKeys, nextFigs
	}

	// after, err := client.Count(ctx, allQ)
	// if err != nil {
	// 	return errors.Wrap(err, "counting figures after replace")
	// }

	log.Printf("replaced figures, %d before deduping, %d after", beforeDeduping, afterDeduping)

	return nil
}

const stale = 30 * 24 * time.Hour

func ExpireFigures(ctx context.Context, client *datastore.Client) error {
	q := datastore.NewQuery("Figure")
	q = q.Filter("Updated <", time.Now().Add(-stale)).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return errors.Wrap(err, "getting stale figures")
	}
	for len(keys) > 0 {
		var nextKeys []*datastore.Key

		if len(keys) > multiLimit {
			keys, nextKeys = keys[:multiLimit], keys[multiLimit:]
		}
		err = client.DeleteMulti(ctx, keys)
		if err != nil {
			return errors.Wrap(err, "expiring figures")
		}
		keys = nextKeys
	}
	return nil
}
