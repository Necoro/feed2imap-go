package http

import (
	ctxt "context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"time"
)

// share HTTP clients
var (
	stdClient    *http.Client
	unsafeClient *http.Client
)

// Error represents an HTTP error returned by a server.
type Error struct {
	StatusCode int
	Status     string
}

type Context struct {
	Timeout    int
	DisableTLS bool
	Jar        CookieJar
}

func (err Error) Error() string {
	return fmt.Sprintf("http error: %s", err.Status)
}

func init() {
	// std
	stdClient = &http.Client{Transport: http.DefaultTransport}

	// unsafe
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig
	unsafeClient = &http.Client{Transport: transport}
}

func (ctx Context) StdContext() (ctxt.Context, ctxt.CancelFunc) {
	return ctxt.WithTimeout(ctxt.Background(), time.Duration(ctx.Timeout)*time.Second)
}

func client(disableTLS bool) *http.Client {
	if disableTLS {
		return unsafeClient
	}
	return stdClient
}

var noop ctxt.CancelFunc = func() {}

type Cookie struct {
	Name   string
	Value  string
	Domain string
}

type CookieJar http.CookieJar

func JarOfCookies(cookies []Cookie, url string) (CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	cs := make([]*http.Cookie, len(cookies))
	for i, c := range cookies {
		cs[i] = &http.Cookie{Name: c.Name, Value: c.Value, Domain: c.Domain}
	}

	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, err
	}

	// ignore the path of the URL
	u.Path = ""

	jar.SetCookies(u, cs)

	return jar, nil
}

func Get(url string, ctx Context) (resp *http.Response, cancel ctxt.CancelFunc, err error) {
	prematureExit := true
	stdCtx, ctxCancel := ctx.StdContext()

	cancel = func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
		ctxCancel()
	}

	defer func() {
		if prematureExit {
			cancel()
		}
	}()

	req, err := http.NewRequestWithContext(stdCtx, "GET", url, nil)
	if err != nil {
		return nil, noop, err
	}
	req.Header.Set("User-Agent", "Feed2Imap-Go/1.0")

	if ctx.Jar != nil {
		for _, c := range ctx.Jar.Cookies(req.URL) {
			req.AddCookie(c)
		}
	}

	resp, err = client(ctx.DisableTLS).Do(req)
	if err != nil {
		return nil, noop, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, noop, Error{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
		}
	}

	prematureExit = false
	return resp, cancel, nil
}
