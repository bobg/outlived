package outlived

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/bobg/aesite"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

type User struct {
	aesite.User
	Born   Date
	Active bool
}

func (u *User) GetUser() *aesite.User {
	return &u.User
}

func (u *User) SetUser(au *aesite.User) {
	u.User = *au
}

func ForUserByAge(ctx context.Context, client *datastore.Client, f func(context.Context, *User) error) error {
	q := datastore.NewQuery("User").Filter("Verified =", true).Filter("Active =", true).Order("Born")
	it := client.Run(ctx, q)
	for {
		var u User
		_, err := it.Next(&u)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Wrap(err, "iterating")
		}
		err = f(ctx, &u)
		if err != nil {
			return errors.Wrapf(err, "calling callback on user %s", u.Email)
		}
	}
	return nil
}
