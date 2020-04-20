package feed

import (
	"bytes"
	"io"
	"time"

	"github.com/emersion/go-message/mail"

	"github.com/Necoro/feed2imap-go/internal/config"
)

func address(name, address string) []*mail.Address {
	return []*mail.Address{{name, address}}
}

func fromAdress(feed Feed, item feeditem, cfg config.Config) []*mail.Address {
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

func asMail(feed Feed, item feeditem, cfg config.Config) (string, error) {
	var b bytes.Buffer

	var h mail.Header
	h.SetAddressList("From", fromAdress(feed, item, cfg))
	h.SetAddressList("To", address(feed.Name, cfg.DefaultEmail))

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

	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		return "", err
	}

	if cfg.WithPartText() {
		tw, err := mw.CreateInline()
		if err != nil {
			return "", err
		}

		var th mail.InlineHeader
		th.SetContentType("text/plain", map[string]string{"charset": "utf-8", "format": "flowed"})

		w, err := tw.CreatePart(th)
		if err != nil {
			return "", err
		}
		_, _ = io.WriteString(w, "Who are you?")

		_ = w.Close()
		_ = tw.Close()
	}

	_ = mw.Close()

	return b.String(), nil
}
