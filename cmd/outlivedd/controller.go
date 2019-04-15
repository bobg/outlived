package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/smtp"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2beta3"
	"cloud.google.com/go/datastore"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type controller struct {
	dsClient   *datastore.Client
	ctClient   *cloudtasks.Client
	sender     mailSender
	projectID  string
	locationID string
}

type mailSender interface {
	send(ctx context.Context, from string, to []string, subject string, body io.Reader) error
}

func newController(ctx context.Context, creds, projectID, locationID, smtpAddr string) (*controller, error) {
	var options []option.ClientOption

	if creds != "" {
		options = append(options, option.WithCredentialsFile(creds))
	}

	dsClient, err := datastore.NewClient(ctx, projectID, options...)
	if err != nil {
		return nil, errors.Wrap(err, "creating datastore client")
	}

	var ctClient *cloudtasks.Client
	if creds != "" {
		// Create a cloudtasks client only if not in test mode,
		// since there is no cloudtasks emulator (yet?).
		ctClient, err = cloudtasks.NewClient(ctx, options...)
		if err != nil {
			return nil, errors.Wrap(err, "creating cloudtasks client")
		}
	}

	sender := &smtpSender{addr: smtpAddr}
	return &controller{
		dsClient:   dsClient,
		ctClient:   ctClient,
		sender:     sender,
		projectID:  projectID,
		locationID: locationID,
	}, nil
}

type smtpSender struct {
	addr string
	auth smtp.Auth
}

func (s *smtpSender) send(_ context.Context, from string, to []string, subject string, body io.Reader) error {
	msg := new(bytes.Buffer)

	fmt.Fprintf(msg, "From: %s\r\n", from)

	if len(to) > 0 {
		fmt.Fprint(msg, "To: ")
		for i, t := range to {
			if i > 0 {
				fmt.Fprint(msg, ",\r\n\t")
			}
			fmt.Fprint(msg, t)
		}
		fmt.Fprint(msg, "\r\n")
	}
	fmt.Fprintf(msg, "Subject: %s\r\n", subject)
	fmt.Fprintf(msg, "Date: %s\r\n", time.Now().Format(time.RFC822Z))
	fmt.Fprint(msg, "\r\n")

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		fmt.Fprint(msg, scanner.Text())
		fmt.Fprint(msg, "\r\n")
	}

	return smtp.SendMail(s.addr, s.auth, from, to, msg.Bytes())
}
