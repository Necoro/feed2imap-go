package http

import (
	ctxt "context"
	"crypto/tls"
	"fmt"
	"net/http"
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
