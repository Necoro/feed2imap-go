package feed

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/Necoro/gofeed"
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
	"github.com/Necoro/feed2imap-go/pkg/rfc822"
	"github.com/Necoro/feed2imap-go/pkg/version"
)

func address(name, address string) []*mail.Address {
	return []*mail.Address{{Name: name, Address: address}}
}

func author(authors []*gofeed.Person) *gofeed.Person {
	if len(authors) > 0 {
		return authors[0]
	}
	return nil
}

func (item *Item) fromAddress() []*mail.Address {
	itemAuthor := author(item.Authors)
	feedAuthor := author(item.Feed.Authors)
	switch {
	case itemAuthor != nil && itemAuthor.Email != "":
		return address(itemAuthor.Name, itemAuthor.Email)
	case itemAuthor != nil && itemAuthor.Name != "":
		return address(itemAuthor.Name, item.defaultEmail())
	case feedAuthor != nil && feedAuthor.Email != "":
		return address(feedAuthor.Name, feedAuthor.Email)
	case feedAuthor != nil && feedAuthor.Name != "":
		return address(feedAuthor.Name, item.defaultEmail())
	default:
		return address(item.feed.Name, item.defaultEmail())
	}
}

func (item *Item) toAddress() []*mail.Address {
	return address(item.feed.Name, item.defaultEmail())
}

func (item *Item) buildHeader() message.Header {
	var h mail.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetAddressList("From", item.fromAddress())
	h.SetAddressList("To", item.toAddress())
	h.Set("Message-Id", item.messageId())
	h.Set(msg.VersionHeader, version.Version())
	h.Set(msg.ReasonHeader, strings.Join(item.reasons, ","))
	h.Set(msg.IdHeader, item.Id())
	h.Set(msg.CreateHeader, time.Now().Format(time.RFC1123Z))
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

func (item *Item) writeContentPart(w *message.Writer, typ string, tpl template.Template) error {
	var ih message.Header
	ih.SetContentType("text/"+typ, map[string]string{"charset": "utf-8"})
	ih.SetContentDisposition("inline", nil)
	ih.Set("Content-Transfer-Encoding", "8bit")

	partW, err := w.CreatePart(ih)
	if err != nil {
		return err
	}
	defer partW.Close()

	if err = tpl.Execute(rfc822.Writer(w), item); err != nil {
		return fmt.Errorf("writing %s part: %w", typ, err)
	}

	return nil
}

func (item *Item) writeTextPart(w *message.Writer) error {
	return item.writeContentPart(w, "plain", template.Text)
}

func (item *Item) writeHtmlPart(w *message.Writer) error {
	return item.writeContentPart(w, "html", template.Html)
}

func (img *feedImage) buildNameMap(key string) map[string]string {
	if img.name == "" {
		return nil
	}
	return map[string]string{key: img.name}
}

func (img *feedImage) writeImagePart(w *message.Writer, cid string) error {
	var ih message.Header
	// set filename for both Type and Disposition
	// according to standard, it belongs to the latter -- but some clients expect the former
	ih.SetContentType(img.mime, img.buildNameMap("name"))
	ih.SetContentDisposition("inline", img.buildNameMap("filename"))
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

func (item *Item) writeToBuffer(b *bytes.Buffer) error {
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

func (item *Item) message() (msg.Message, error) {
	var b bytes.Buffer

	if err := item.writeToBuffer(&b); err != nil {
		return msg.Message{}, err
	}

	msg := msg.Message{
		Content:  b.String(),
		IsUpdate: item.UpdateOnly,
		ID:       item.Id(),
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

func getImage(src string, ctx http.Context) ([]byte, string, error) {
	resp, cancel, err := http.Get(src, ctx)
	if err != nil {
		return nil, "", fmt.Errorf("fetching from '%s': %w", src, err)
	}
	defer cancel()

	img, err := io.ReadAll(resp.Body)
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

func (item *Item) resolveUrl(otherUrlStr string) string {
	feed := item.feed
	feedUrl := feed.url()

	if feedUrl == nil {
		// no url, just return the original
		return otherUrlStr
	}

	otherUrl, err := url.Parse(otherUrlStr)
	if err != nil {
		log.Errorf("Feed %s: Item %s: Error parsing URL '%s' embedded in item: %s",
			feed.Name, item.Link, otherUrlStr, err)
		return ""
	}

	return feedUrl.ResolveReference(otherUrl).String()
}

func (item *Item) downloadImage(src string) string {
	feed := item.feed

	imgUrl := item.resolveUrl(src)

	img, mime, err := getImage(imgUrl, feed.Context())
	if err != nil {
		log.Errorf("Feed %s: Item %s: Error fetching image: %s",
			feed.Name, item.Link, err)
		return ""
	}
	if img == nil {
		return ""
	}

	if feed.EmbedImages {
		return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(img)
	} else {
		name := path.Base(src)
		if name == "/" || name == "." || name == " " {
			name = ""
		}

		idx := item.addImage(img, mime, name)
		return "cid:" + cidNr(idx)
	}
}

func (item *Item) buildBody() {
	feed := item.feed

	item.Body = getBody(item.Content, item.Description, feed.Body)
	bodyNode, err := html.Parse(strings.NewReader(item.Body))
	if err != nil {
		log.Errorf("Feed %s: Item %s: Error while parsing html: %s", feed.Name, item.Link, err)
		item.TextBody = item.Body
		return
	}

	doc := goquery.NewDocumentFromNode(bodyNode)
	doneAnything := false

	updateBody := func() {
		if doneAnything {
			html, err := goquery.OuterHtml(doc.Selection)
			if err != nil {
				item.clearImages()
				log.Errorf("Feed %s: Item %s: Error during rendering HTML: %s",
					feed.Name, item.Link, err)
			} else {
				item.Body = html
			}
		}
	}

	// make relative links absolute
	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		const attr = "href"

		src, ok := selection.Attr(attr)
		if !ok {
			return
		}

		if src != "" && src[0] == '/' {
			absUrl := item.resolveUrl(src)
			selection.SetAttr(attr, absUrl)
			doneAnything = true
		}
	})

	if feed.Global.WithPartText() {
		if item.TextBody, err = html2text.FromHTMLNode(bodyNode, html2text.Options{CitationStyleLinks: true}); err != nil {
			log.Errorf("Feed %s: Item %s: Error while converting html to text: %s", feed.Name, item.Link, err)
		}
	}

	if !feed.Global.WithPartHtml() || err != nil {
		return
	}

	if !feed.InclImages {
		updateBody()
		return
	}

	// download images
	doc.Find("img").Each(func(i int, selection *goquery.Selection) {
		const attr = "src"

		src, ok := selection.Attr(attr)
		if !ok {
			return
		}

		if !strings.HasPrefix(src, "data:") {
			if imgStr := item.downloadImage(src); imgStr != "" {
				selection.SetAttr(attr, imgStr)
			}
		}

		// srcset overrides src and would reload all the images
		// we do not want to include all images in the srcset either, so just strip it
		selection.RemoveAttr("srcset")

		doneAnything = true
	})

	updateBody()
}
