package outlived

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"sort"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

// Figure is a historical figure.
type Figure struct {
	Name, Desc, Link string
	Born, Died       Date
	DaysAlive        int
	Pageviews        int

	TimestampMS int64
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

func FiguresAliveFor(ctx context.Context, client *datastore.Client, days int) ([]*Figure, error) {
	q, err := figureQuery(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "constructing query")
	}
	q = q.Filter("DaysAlive =", days).Order("-Pageviews")
	var figures []*Figure
	_, err = client.GetAll(ctx, q, &figures)
	return figures, errors.Wrap(err, "querying figures")
}

func FiguresAliveForAtMost(ctx context.Context, client *datastore.Client, days, limit int) ([]*Figure, error) {
	q, err := figureQuery(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "constructing query")
	}
	q = q.Filter("DaysAlive <=", days).Order("-DaysAlive").Order("-Pageviews")
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
	q, err := figureQuery(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "constructing query")
	}
	q = q.Filter("Died.M =", int(mon)).Filter("Died.D =", day).Order("-Pageviews")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var figures []*Figure
	_, err = client.GetAll(ctx, q, &figures)
	return figures, errors.Wrap(err, "querying figures")
}

func figureQuery(ctx context.Context, client *datastore.Client) (*datastore.Query, error) {
	timestampMSBytes, err := aesite.GetSetting(ctx, client, "generation")
	if err != nil {
		return nil, errors.Wrap(err, "getting generation setting")
	}
	timestampMS, n := binary.Varint(timestampMSBytes)
	if n <= 0 {
		return nil, errors.Wrapf(err, "decoding generation setting %x", timestampMSBytes)
	}
	return datastore.NewQuery("Figure").Filter("TimestampMS = ", timestampMS), nil
}

const multiLimit = 500

func ReplaceFigures(ctx context.Context, client *datastore.Client, figures []*Figure) error {
	nowMS := time.Now().UnixNano() * int64(time.Nanosecond) / int64(time.Millisecond)
	keys := make([]*datastore.Key, len(figures))
	for i := range figures {
		keys[i] = &datastore.Key{Kind: "Figure"}
		figures[i].TimestampMS = nowMS
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
	var nowMSBytes [binary.MaxVarintLen64]byte
	n := binary.PutVarint(nowMSBytes[:], nowMS)
	err := aesite.SetSetting(ctx, client, "generation", nowMSBytes[:n])
	if err != nil {
		return errors.Wrap(err, "finalizing figure storage")
	}
	q := datastore.NewQuery("Figure").Filter("TimestampMS <", nowMS).KeysOnly()
	keys, err = client.GetAll(ctx, q, nil)
	if err != nil {
		// Log but otherwise ignore this error.
		log.Printf("querying stale figures with TimestampMS < %d: %s", nowMS, err) // xxx use "appengine/log".Warningf when in the appengine context?
		return nil
	}

	for len(keys) > 0 {
		var nextKeys []*datastore.Key
		if len(keys) > multiLimit {
			keys, nextKeys = keys[:multiLimit], keys[multiLimit:]
		}
		err = client.DeleteMulti(ctx, keys)
		if err != nil {
			// Log but otherwise ignore this error.
			log.Printf("deleting %d stale figures with TimestampMS < %d: %s", len(keys), nowMS, err) // xxx use "appengine/log".Warningf when in the appengine context?
			break
		}
		keys = nextKeys
	}
	return nil
}
