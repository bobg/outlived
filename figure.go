package outlived

import (
	"context"
	"encoding/binary"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/appengine/log"
)

// Figure is a historical figure.
type Figure struct {
	Name, Desc, Link string
	Born, Died       Date
	DaysAlive        int
	Pageviews        int

	TimestampMS int64
}

func FiguresAliveFor(ctx context.Context, client *datastore.Client, days int) ([]*Figure, error) {
	q, err := figureQuery(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "constructing query")
	}
	q = q.Filter("DaysAlive =", days)
	var figures []*Figure
	_, err = client.GetAll(ctx, q, &figures)
	return figures, errors.Wrap(err, "querying figures")
}

func FiguresAliveForAtMost(ctx context.Context, client *datastore.Client, days int) ([]*Figure, error) {
	q, err := figureQuery(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "constructing query")
	}
	q = q.Filter("DaysAlive <=", days).Order("-DaysAlive")
	it := client.Run(ctx, q)
	var figures []*Figure
	for {
		var fig Figure
		_, err := it.Next(&fig)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "iterating")
		}
		if len(figures) > 0 && fig.DaysAlive != figures[0].DaysAlive {
			break
		}
		figures = append(figures, &fig)
	}
	return figures, nil
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

const putMultiLimit = 500

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
		if len(figures) > putMultiLimit {
			keys, nextKeys = keys[:putMultiLimit], keys[putMultiLimit:]
			figures, nextFigs = figures[:putMultiLimit], figures[putMultiLimit:]
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
		log.Warningf(ctx, "querying stale figures with TimestampMS < %d: %s", nowMS, err)
		return nil
	}
	err = client.DeleteMulti(ctx, keys)
	if err != nil {
		// Log but otherwise ignore this error.
		log.Warningf(ctx, "deleting %d stale figures with TimestampMS < %d: %s", len(keys), nowMS, err)
	}
	return nil
}
