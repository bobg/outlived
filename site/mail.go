package site

import (
	"context"
	"io"
	"io/ioutil"
	"net/smtp"

	"github.com/mailgun/mailgun-go"
	"github.com/pkg/errors"
)

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

const maxSMTPRecips = 20

type smtpSender struct {
	addr string
	auth smtp.Auth
}

func (s *smtpSender) send(ctx context.Context, from string, to []string, subject string, textR io.Reader, htmlR io.Reader) error {
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
		if len(to) > maxSMTPRecips {
			to, nextTo = to[:maxSMTPRecips], to[maxSMTPRecips:]
		}

		client, err := smtp.Dial(s.addr)
		if err != nil {
			return errors.Wrapf(err, "connecting to %s", s.addr)
		}

		err = s.sendBatch(ctx, client, from, to, subject, textBody, htmlBody)
		if err != nil {
			return err
		}

		to = nextTo
	}

	return nil
}

func (s *smtpSender) sendBatch(ctx context.Context, client *smtp.Client, from string, to []string, subject string, textBody, htmlBody []byte) error {
	defer client.Close()

	if s.auth != nil {
		err := client.Auth(s.auth)
		if err != nil {
			return errors.Wrapf(err, "authenticating to %s", s.addr)
		}
	}

	err := client.Mail(from)
	if err != nil {
		return errors.Wrapf(err, "sending MAIL FROM to %s", s.addr)
	}

	for _, recip := range to {
		err = client.Rcpt(recip)
		if err != nil {
			return errors.Wrapf(err, "sending RCPT TO:%s to %s", recip, s.addr)
		}
	}

	w, err := client.Data()
	if err != nil {
		return errors.Wrapf(err, "sending DATA command to %s", s.addr)
	}
	defer w.Close()

	// TODO: enforce \r\n line-endings

	// xxx write header to w

	if len(htmlBody) > 0 {
		// xxx write preamble, boundary, and subpart header to w
	}
	_, err = w.Write(textBody)
	if err != nil {
		return errors.Wrap(err, "writing text body")
	}

	if len(htmlBody) > 0 {
		// xxx write boundary
		_, err = w.Write(htmlBody)
		if err != nil {
			return errors.Wrap(err, "writing html body")
		}
		// xxx write boundary
	}

	err = w.Close()
	if err != nil {
		return errors.Wrap(err, "sending CLOSE command")
	}

	err = client.Quit()
	if err != nil {
		return errors.Wrap(err, "sending QUIT command")
	}

	err = client.Close()
	return errors.Wrap(err, "closing SMTP client")
}
