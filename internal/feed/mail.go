package feed

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"

	"github.com/Necoro/feed2imap-go/internal/feed/template"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
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

func buildHeader(feed *Feed, item feeditem, cfg *config.Config) message.Header {
	var h mail.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetAddressList("From", fromAdress(feed, item, cfg))
	h.SetAddressList("To", address(feed.Name, cfg.DefaultEmail))
	h.Add("X-Feed2Imap-Version", config.Version())
	h.Add("X-Feed2Imap-Reason", strings.Join(item.reasons, ","))

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

	return h.Header
}

func writeHtmlPart(w *message.Writer, item feeditem) error {
	var ih message.Header
	ih.SetContentType("text/html", map[string]string{"charset": "utf-8"})
	ih.SetContentDisposition("inline", nil)
	ih.Set("Content-Transfer-Encoding", "8bit")

	partW, err := w.CreatePart(ih)
	if err != nil {
		return err
	}
	defer partW.Close()

	if err = writeHtml(w, item); err != nil {
		return fmt.Errorf("writing html part: %w", err)
	}

	return nil
}

func writeImagePart(w *message.Writer, img feedImage, cid string) error {
	var ih message.Header
	ih.SetContentType(img.mime, nil)
	ih.SetContentDisposition("inline", nil)
	ih.Set("Content-Transfer-Encoding", "base64")
	ih.SetText("Content-ID", fmt.Sprintf("<%s>", cid))

	imgW, err := w.CreatePart(ih)
	if err != nil {
		return err
	}
	defer imgW.Close()

	if _, err = imgW.Write(img.image); err != nil {
		return err
	}

	return nil
}

func writeToBuffer(b *bytes.Buffer, feed *Feed, item feeditem, cfg *config.Config) error {
	h := buildHeader(feed, item, cfg)

	writer, err := message.CreateWriter(b, h)
	if err != nil {
		return err
	}
	defer writer.Close()

	if cfg.WithPartHtml() {
		feed.buildBody(&item)

		var relWriter *message.Writer
		if len(item.images) > 0 {
			var rh message.Header
			rh.SetContentType("multipart/related", map[string]string{"type": "text/html"})
			if relWriter, err = writer.CreatePart(rh); err != nil {
				return err
			}
			defer relWriter.Close()
		} else {
			relWriter = writer
		}

		if err = writeHtmlPart(relWriter, item); err != nil {
			return err
		}

		for idx, img := range item.images {
			cid := cidNr(idx + 1)
			if err = writeImagePart(relWriter, img, cid); err != nil {
				return err
			}
		}

		item.clearImages() // safe memory
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

func getImage(src string) ([]byte, string) {
	resp, err := stdHTTPClient.Get(src)
	if err != nil {
		log.Errorf("Error fetching from '%s': %s", src, err)
		return nil, ""
	}
	defer resp.Body.Close()

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading body from '%s': %s", src, err)
		return nil, ""
	}

	ext := path.Ext(src)
	if ext == "" {
		log.Warnf("Cannot determine extension from '%s', skipping.", src)
		return nil, ""
	}

	mime := mime.TypeByExtension(ext)
	return img, mime
}

func cidNr(idx int) string {
	return fmt.Sprintf("cid_%d", idx)
}

func (feed *Feed) buildBody(item *feeditem) {
	var body string
	var comment string

	if item.Item.Content != "" {
		comment = "<!-- Content -->\n"
		body = item.Item.Content
	} else if item.Item.Description != "" {
		comment = "<!-- Description -->\n"
		body = item.Item.Description
	}

	if !feed.InclImages {
		item.Body = comment + body
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Debugf("Feed %s: Error while parsing html content: %s", feed.Name, err)
		if body != "" {
			item.Body = "<br />" + comment + body
		}
		return
	}

	doneAnything := true
	nodes := doc.Find("img")
	nodes.Each(func(i int, selection *goquery.Selection) {
		const attr = "src"

		src, ok := selection.Attr(attr)
		if !ok {
			return
		}

		img, mime := getImage(src)
		if img == nil {
			return
		}

		if feed.EmbedImages {
			imgStr := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(img)
			selection.SetAttr(attr, imgStr)
		} else {
			idx := item.addImage(img, mime)
			cid := "cid:" + cidNr(idx)
			selection.SetAttr(attr, cid)
		}
		doneAnything = true
	})

	if doneAnything {
		html, err := doc.Find("body").Html()
		if err != nil {
			item.clearImages()
			log.Errorf("Error during rendering HTML, skipping.")
		} else {
			body = html
		}
	}

	item.Body = comment + body
}
