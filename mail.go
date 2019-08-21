package outlived

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
)

const maxRecipients = 20

func SendMail(ctx context.Context, client *datastore.Client) error {
	var (
		today    = Today()
		lastBorn Date
		users    []*User
	)
	wrap := func() error {
		defer func() {
			users = nil
		}()

		since := today.Since(users[0].Born)
		figures, err := FiguresAliveFor(ctx, client, since-1, 20)
		if err != nil {
			return errors.Wrapf(err, "looking up figures alive for %d days", since-1)
		}
		if len(figures) == 0 {
			return nil
		}
		for len(users) > 0 {
			var nextUsers []*User
			if len(users) > maxRecipients {
				users, nextUsers = users[:maxRecipients], nextUsers[maxRecipients:]
			}
			// xxx send figures to users
			users = nextUsers
		}
		return nil
	}
	err := ForUserByAge(ctx, client, func(ctx context.Context, user *User) error {
		if user.Born != lastBorn {
			err := wrap()
			if err != nil {
				return errors.Wrapf(err, "processing users born on %s", lastBorn)
			}
		}
		users = append(users, user)
		lastBorn = user.Born
		return nil
	})
	if err != nil {
		return err
	}
	return wrap()
}
