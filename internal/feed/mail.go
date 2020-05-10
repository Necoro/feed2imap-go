package feed

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/Necoro/html2text"
	"github.com/PuerkitoBio/goquery"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/net/html"

	"github.com/Necoro/feed2imap-go/internal/feed/template"
	"github.com/Necoro/feed2imap-go/internal/http"
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
	case item.Author != nil && item.Author.Email != "":
		return address(item.Author.Name, item.Author.Email)
	case item.Author != nil && item.Author.Name != "":
		return address(item.Author.Name, item.defaultEmail())
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

func (item *item) buildHeader() message.Header {
	var h mail.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetAddressList("From", item.fromAddress())
	h.SetAddressList("To", item.toAddress())
	h.Set("Message-Id", item.messageId())
	h.Set(msg.VersionHeader, version.Version())
	h.Set(msg.ReasonHeader, strings.Join(item.reasons, ","))
	h.Set(msg.IdHeader, item.id())
	if item.GUID != "" {
		h.Set(msg.GuidHeader, item.GUID)
	}

	{ // date
		date := item.DateParsed()
		if date == nil {
			now := time.Now()
			date = &now
		}
		h.SetDate(*date)
	}
	{ // subject
		subject := item.Title
		if subject == "" {
			subject = item.Date()
		}
		if subject == "" {
			subject = item.Link
		}
		h.SetSubject(subject)
	}

	return h.Header
}

func (item *item) writeContentPart(w *message.Writer, typ string, tpl template.Template) error {
	var ih message.Header
	ih.SetContentType("text/"+typ, map[string]string{"charset": "utf-8"})
	ih.SetContentDisposition("inline", nil)
	ih.Set("Content-Transfer-Encoding", "8bit")

	partW, err := w.CreatePart(ih)
	if err != nil {
		return err
	}
	defer partW.Close()

	if err = tpl.Execute(w, item); err != nil {
		return fmt.Errorf("writing %s part: %w", typ, err)
	}

	return nil
}

func (item *item) writeTextPart(w *message.Writer) error {
	return item.writeContentPart(w, "plain", template.Text)
}

func (item *item) writeHtmlPart(w *message.Writer) error {
	return item.writeContentPart(w, "html", template.Html)
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
	item.buildBody()

	writer, err := message.CreateWriter(b, h)
	if err != nil {
		return err
	}
	defer writer.Close()

	if item.feed.Global.WithPartText() {
		if err = item.writeTextPart(writer); err != nil {
			return err
		}
	}

	if item.feed.Global.WithPartHtml() {
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
	}

	item.clearImages() // safe memory
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

func getImage(src string, timeout int, disableTLS bool) ([]byte, string, error) {
	resp, cancel, err := http.Get(src, timeout, disableTLS)
	if err != nil {
		return nil, "", fmt.Errorf("fetching from '%s': %w", src, err)
	}
	defer cancel()

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

	body := getBody(item.Content, item.Description, feed.Body)
	bodyNode, err := html.Parse(strings.NewReader(body))
	if err != nil {
		log.Errorf("Feed %s: Item %s: Error while parsing html: %s", feed.Name, item.Link, err)
		item.Body = body
		item.TextBody = body
		return
	}

	if feed.Global.WithPartText() {
		if item.TextBody, err = html2text.FromHTMLNode(bodyNode, html2text.Options{CitationStyleLinks: true}); err != nil {
			log.Errorf("Feed %s: Item %s: Error while converting html to text: %s", feed.Name, item.Link, err)
		}
	}

	if !feed.InclImages || !feed.Global.WithPartHtml() || err != nil {
		item.Body = body
		return
	}

	doc := goquery.NewDocumentFromNode(bodyNode)

	doneAnything := true
	nodes := doc.Find("img")
	nodes.Each(func(i int, selection *goquery.Selection) {
		const attr = "src"

		src, ok := selection.Attr(attr)
		if !ok {
			return
		}

		srcUrl, err := url.Parse(src)
		if err != nil {
			log.Errorf("Feed %s: Item %s: Error parsing URL '%s' embedded in item: %s",
				feed.Name, item.Link, src, err)
			return
		}
		imgUrl := feedUrl.ResolveReference(srcUrl)

		img, mime, err := getImage(imgUrl.String(), feed.Global.Timeout, feed.NoTLS)
		if err != nil {
			log.Errorf("Feed %s: Item %s: Error fetching image: %s",
				feed.Name, item.Link, err)
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
				feed.Name, item.Link, err)
		} else {
			body = html
		}
	}

	item.Body = body
}
