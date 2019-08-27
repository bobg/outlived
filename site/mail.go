package site

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"github.com/mailgun/mailgun-go"
	"github.com/pkg/errors"
)

const from = "Outlived <bobg+outlived@emphatic.com>" // xxx

type sender interface {
	send(
		ctx context.Context,
		from string,
		to []string,
		subject string,
		textBody io.Reader,
		htmlBody io.Reader,
	) error
}

type mailgunSender struct {
	mg *mailgun.MailgunImpl
}

func (mg *mailgunSender) send(ctx context.Context, from string, to []string, subject string, textR io.Reader, htmlR io.Reader) error {
	textBody, err := ioutil.ReadAll(textR)
	if err != nil {
		return errors.Wrap(err, "reading text body")
	}
	var htmlBody []byte
	if htmlR != nil {
		htmlBody, err = ioutil.ReadAll(htmlR)
		if err != nil {
			return errors.Wrap(err, "reading html body")
		}
	}

	for len(to) > 0 {
		var nextTo []string
		if len(to) > mailgun.MaxNumberOfRecipients {
			to, nextTo = to[:mailgun.MaxNumberOfRecipients], to[mailgun.MaxNumberOfRecipients:]
		}

		msg := mg.mg.NewMessage(from, subject, string(textBody))
		if htmlBody != nil {
			msg.SetHtml(string(htmlBody))
		}

		for _, recip := range to {
			msg.AddBCC(recip)
		}
		_, _, err = mg.mg.Send(msg)
		if err != nil {
			return errors.Wrapf(err, "sending to %d recipient(s)", len(to))
		}

		to = nextTo
	}

	return nil
}

type testSender struct{}

func (ts *testSender) send(ctx context.Context, from string, to []string, subject string, textR io.Reader, htmlR io.Reader) error {
	log.Printf("sending e-mail from %s, subject %s, to %s", from, subject, strings.Join(to, ", "))
	if textR != nil {
		textBody, err := ioutil.ReadAll(textR)
		if err != nil {
			return errors.Wrap(err, "reading text body")
		}
		log.Printf("text body:\n%s", string(textBody))
	}
	if htmlR != nil {
		htmlBody, err := ioutil.ReadAll(htmlR)
		if err != nil {
			return errors.Wrap(err, "reading html body")
		}
		log.Printf("html body:\n%s", string(htmlBody))
	}
	return nil
}
