package feed

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/gabriel-vasile/mimetype"

	"github.com/Necoro/feed2imap-go/internal/feed/template"
	"github.com/Necoro/feed2imap-go/internal/msg"
	"github.com/Necoro/feed2imap-go/pkg/config"
	"github.com/Necoro/feed2imap-go/pkg/log"
	"github.com/Necoro/feed2imap-go/pkg/version"
)

func address(name, address string) []*mail.Address {
	return []*mail.Address{{Name: name, Address: address}}
}

func (item *item) fromAddress() []*mail.Address {
	switch {
	case item.Item.Author != nil && item.Item.Author.Email != "":
		return address(item.Item.Author.Name, item.Item.Author.Email)
	case item.Item.Author != nil && item.Item.Author.Name != "":
		return address(item.Item.Author.Name, item.defaultEmail())
	case item.Feed.Author != nil && item.Feed.Author.Email != "":
		return address(item.Feed.Author.Name, item.Feed.Author.Email)
	case item.Feed.Author != nil && item.Feed.Author.Name != "":
		return address(item.Feed.Author.Name, item.defaultEmail())
	default:
		return address(item.feed.Name, item.defaultEmail())
	}
}

func (item *item) toAddress() []*mail.Address {
	return address(item.feed.Name, item.defaultEmail())
}

func (item *item) writeHtml(writer io.Writer) error {
	return template.Feed.Execute(writer, item)
}

func (item *item) buildHeader() message.Header {
	var h mail.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetAddressList("From", item.fromAddress())
	h.SetAddressList("To", item.toAddress())
	h.Set(msg.VersionHeader, version.Version())
	h.Set(msg.ReasonHeader, strings.Join(item.reasons, ","))
	h.Set(msg.IdHeader, item.id())
	h.Set("Message-Id", item.messageId())

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

func (item *item) writeHtmlPart(w *message.Writer) error {
	var ih message.Header
	ih.SetContentType("text/html", map[string]string{"charset": "utf-8"})
	ih.SetContentDisposition("inline", nil)
	ih.Set("Content-Transfer-Encoding", "8bit")

	partW, err := w.CreatePart(ih)
	if err != nil {
		return err
	}
	defer partW.Close()

	if err = item.writeHtml(w); err != nil {
		return fmt.Errorf("writing html part: %w", err)
	}

	return nil
}

func (img *feedImage) writeImagePart(w *message.Writer, cid string) error {
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

func (item *item) writeToBuffer(b *bytes.Buffer) error {
	h := item.buildHeader()

	writer, err := message.CreateWriter(b, h)
	if err != nil {
		return err
	}
	defer writer.Close()

	if item.feed.Global.WithPartHtml() {
		item.buildBody()

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

		if err = item.writeHtmlPart(relWriter); err != nil {
			return err
		}

		for idx, img := range item.images {
			cid := cidNr(idx + 1)
			if err = img.writeImagePart(relWriter, cid); err != nil {
				return err
			}
		}

		item.clearImages() // safe memory
	}

	return nil
}

func (item *item) message() (msg.Message, error) {
	var b bytes.Buffer

	if err := item.writeToBuffer(&b); err != nil {
		return msg.Message{}, err
	}

	msg := msg.Message{
		Content:  b.String(),
		IsUpdate: item.updateOnly,
		ID:       item.id(),
	}

	return msg, nil
}

func (feed *Feed) Messages() (msg.Messages, error) {
	var (
		err   error
		mails = make([]msg.Message, len(feed.items))
	)
	for idx := range feed.items {
		if mails[idx], err = feed.items[idx].message(); err != nil {
			return nil, fmt.Errorf("creating mails for %s: %w", feed.Name, err)
		}
	}
	return mails, nil
}

func getImage(src string, client *http.Client) ([]byte, string, error) {
	resp, err := client.Get(src)
	if err != nil {
		return nil, "", fmt.Errorf("fetching from '%s': %w", src, err)
	}
	defer resp.Body.Close()

	img, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading from '%s': %w", src, err)
	}

	var mimeStr string
	ext := path.Ext(src)
	if ext == "" {
		mimeStr = mimetype.Detect(img).String()
	} else {
		mimeStr = mime.TypeByExtension(ext)
	}
	return img, mimeStr, nil
}

func cidNr(idx int) string {
	return fmt.Sprintf("cid_%d", idx)
}

func getBody(content, description string, bodyCfg config.Body) string {
	switch bodyCfg {
	case "default":
		if content != "" {
			return content
		}
		return description
	case "description":
		return description
	case "content":
		return content
	case "both":
		return description + content
	default:
		panic(fmt.Sprintf("Unknown value for Body: %v", bodyCfg))
	}
}

func (item *item) buildBody() {
	feed := item.feed
	feedUrl, err := url.Parse(feed.Url)
	if err != nil {
		panic(fmt.Sprintf("URL '%s' of feed '%s' is not a valid URL. How have we ended up here?", feed.Url, feed.Name))
	}

	body := getBody(item.Item.Content, item.Item.Description, feed.Body)

	if !feed.InclImages {
		item.Body = body
		return
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Errorf("Feed %s: Item %s: Error while parsing html content: %s", feed.Name, item.Item.Link, err)
		if body != "" {
			item.Body = "<br />" + body
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

		srcUrl,err := url.Parse(src)
		if err != nil {
			log.Errorf("Feed %s: Item %s: Error parsing URL '%s' embedded in item: %s",
				feed.Name, item.Item.Link, src, err)
			return
		}
		imgUrl := feedUrl.ResolveReference(srcUrl)

		img, mime, err := getImage(imgUrl.String(), httpClient(feed.NoTLS))
		if err != nil {
			log.Errorf("Feed %s: Item %s: Error fetching image: %s",
				feed.Name, item.Item.Link, err)
			return
		}
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
			log.Errorf("Feed %s: Item %s: Error during rendering HTML: %s",
				feed.Name, item.Item.Link, err)
		} else {
			body = html
		}
	}

	item.Body = body
}
