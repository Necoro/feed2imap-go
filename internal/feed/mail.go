package feed

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/Necoro/feed2imap-go/internal/feed/template"
	"github.com/Necoro/feed2imap-go/pkg/config"
)

func address(name, address string) []*mail.Address {
	return []*mail.Address{{Name: name, Address: address}}
}

func fromAdress(feed *Feed, item feeditem, cfg *config.Config) []*mail.Address {
	switch {
	case item.Item.Author != nil && item.Item.Author.Email != "":
		return address(item.Item.Author.Name, item.Item.Author.Email)
	case item.Item.Author != nil && item.Item.Author.Name != "":
		return address(item.Item.Author.Name, cfg.DefaultEmail)
	case item.Feed.Author != nil && item.Feed.Author.Email != "":
		return address(item.Feed.Author.Name, item.Feed.Author.Email)
	case item.Feed.Author != nil && item.Feed.Author.Name != "":
		return address(item.Feed.Author.Name, cfg.DefaultEmail)
	default:
		return address(feed.Name, cfg.DefaultEmail)
	}
}

func writeHtml(writer io.Writer, item feeditem) error {
	return template.Feed.Execute(writer, item)
}

func writeToBuffer(b *bytes.Buffer, feed *Feed, item feeditem, cfg *config.Config) error {
	var h mail.Header
	h.SetAddressList("From", fromAdress(feed, item, cfg))
	h.SetAddressList("To", address(feed.Name, cfg.DefaultEmail))
	h.Add("X-Feed2Imap-Version", config.Version())

	{ // date
		date := item.Item.PublishedParsed
		if date == nil {
			now := time.Now()
			date = &now
		}
		h.SetDate(*date)
	}
	{ // subject
		subject := item.Item.Title
		if subject == "" {
			subject = item.Item.Published
		}
		if subject == "" {
			subject = item.Item.Link
		}
		h.SetSubject(subject)
	}

	tw, err := mail.CreateInlineWriter(b, h)
	if err != nil {
		return err
	}
	defer tw.Close()

	if false /* cfg.WithPartText() */ {
		var th mail.InlineHeader
		th.SetContentType("text/plain", map[string]string{"charset": "utf-8", "format": "flowed"})

		w, err := tw.CreatePart(th)
		if err != nil {
			return err
		}
		defer w.Close()

		_, _ = io.WriteString(w, "Who are you?")
	}

	if cfg.WithPartHtml() {
		var th mail.InlineHeader
		th.SetContentType("text/html", map[string]string{"charset": "utf-8"})

		w, err := tw.CreatePart(th)
		if err != nil {
			return err
		}

		if err = writeHtml(w, item); err != nil {
			return fmt.Errorf("writing html part: %w", err)
		}

		w.Close()
	}

	return nil
}

func asMail(feed *Feed, item feeditem, cfg *config.Config) (string, error) {
	var b bytes.Buffer

	if err := writeToBuffer(&b, feed, item, cfg); err != nil {
		return "", err
	}

	return b.String(), nil
}

func (feed *Feed) ToMails(cfg *config.Config) ([]string, error) {
	var (
		err   error
		mails = make([]string, len(feed.items))
	)
	for idx := range feed.items {
		if mails[idx], err = asMail(feed, feed.items[idx], cfg); err != nil {
			return nil, fmt.Errorf("creating mails for %s: %w", feed.Name, err)
		}
	}
	return mails, nil
}
