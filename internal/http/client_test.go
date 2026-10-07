package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackedBody struct {
	io.ReadCloser
	closes int
}

func (b *trackedBody) Close() error {
	b.closes++
	return b.ReadCloser.Close()
}

func TestGetResponseCleanup(t *testing.T) {
	for _, disableTLS := range []bool{false, true} {
		for _, status := range []int{200, 204, 299, 300, 400, 404, 500} {
			t.Run(fmt.Sprintf("disableTLS=%t/status=%d", disableTLS, status), func(t *testing.T) {
				body := &trackedBody{ReadCloser: io.NopCloser(strings.NewReader("response"))}
				var requestContext context.Context
				transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
					requestContext = req.Context()
					return &http.Response{
						StatusCode: status,
						Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
						Body:       body,
						Header:     make(http.Header),
					}, nil
				})
				c := client(disableTLS)
				original := c.Transport
				c.Transport = transport
				t.Cleanup(func() { c.Transport = original })

				resp, cancel, err := Get("https://example.com/feed", Context{Timeout: 60, DisableTLS: disableTLS})
				if cancel == nil {
					t.Fatal("Get returned a nil cleanup function")
				}
				defer cancel()
				if status < 200 || status >= 300 {
					var statusError Error
					if !errors.As(err, &statusError) || statusError.StatusCode != status {
						t.Fatalf("Get error = %v, want status %d", err, status)
					}
					if resp != nil {
						t.Error("Get returned a response for an HTTP error")
					}
					if body.closes != 1 {
						t.Errorf("body closed %d times before returning an error, want 1", body.closes)
					}
					if requestContext.Err() != context.Canceled {
						t.Errorf("request context error = %v, want context.Canceled", requestContext.Err())
					}
					cancel()
					if body.closes != 1 {
						t.Errorf("error cleanup closed the body again: %d calls", body.closes)
					}
					return
				}

				if err != nil {
					t.Fatal(err)
				}
				if resp == nil {
					t.Fatal("Get returned a nil response without an error")
				}
				if body.closes != 0 || requestContext.Err() != nil {
					t.Fatal("Get cleaned up the successful response before returning it")
				}
				data, err := io.ReadAll(resp.Body)
				if err != nil || string(data) != "response" {
					t.Fatalf("response body = %q, error = %v", data, err)
				}
				cancel()
				if body.closes != 1 || requestContext.Err() != context.Canceled {
					t.Errorf("cleanup left body closes=%d, context error=%v", body.closes, requestContext.Err())
				}
			})
		}
	}
}

func TestGetTransportErrorCleanup(t *testing.T) {
	for _, disableTLS := range []bool{false, true} {
		t.Run(fmt.Sprintf("disableTLS=%t", disableTLS), func(t *testing.T) {
			transportError := errors.New("connection failed")
			var requestContext context.Context
			c := client(disableTLS)
			original := c.Transport
			c.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				requestContext = req.Context()
				return nil, transportError
			})
			t.Cleanup(func() { c.Transport = original })

			resp, cancel, err := Get("https://example.com/feed", Context{Timeout: 60, DisableTLS: disableTLS})
			if !errors.Is(err, transportError) || resp != nil {
				t.Fatalf("Get returned response=%v, error=%v", resp, err)
			}
			if cancel == nil {
				t.Fatal("Get returned a nil cleanup function")
			}
			if requestContext.Err() != context.Canceled {
				t.Errorf("request context error = %v, want context.Canceled", requestContext.Err())
			}
			cancel()
		})
	}
}

func TestGetInvalidURL(t *testing.T) {
	resp, cancel, err := Get(":invalid", Context{Timeout: 60})
	if err == nil || resp != nil {
		t.Fatalf("Get returned response=%v, error=%v", resp, err)
	}
	if cancel == nil {
		t.Fatal("Get returned a nil cleanup function")
	}
	cancel()
}

func TestGetRedirectErrorCleanup(t *testing.T) {
	for _, disableTLS := range []bool{false, true} {
		t.Run(fmt.Sprintf("disableTLS=%t", disableTLS), func(t *testing.T) {
			body := &trackedBody{ReadCloser: io.NopCloser(strings.NewReader("redirect"))}
			var requestContext context.Context
			redirectError := errors.New("redirect rejected")
			c := client(disableTLS)
			originalTransport, originalRedirect := c.Transport, c.CheckRedirect
			c.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				requestContext = req.Context()
				return &http.Response{
					StatusCode: http.StatusFound,
					Body:       body,
					Header:     http.Header{"Location": {"https://example.com/redirect"}},
				}, nil
			})
			c.CheckRedirect = func(*http.Request, []*http.Request) error { return redirectError }
			t.Cleanup(func() {
				c.Transport, c.CheckRedirect = originalTransport, originalRedirect
			})

			resp, cancel, err := Get("https://example.com/feed", Context{Timeout: 60, DisableTLS: disableTLS})
			if !errors.Is(err, redirectError) || resp != nil {
				t.Fatalf("Get returned response=%v, error=%v", resp, err)
			}
			if requestContext.Err() != context.Canceled {
				t.Errorf("request context error = %v, want context.Canceled", requestContext.Err())
			}
			if body.closes != 1 {
				t.Errorf("body closed %d times, want 1", body.closes)
			}
			cancel()
		})
	}
}

func TestGetStreamingErrorCleanup(t *testing.T) {
	for _, disableTLS := range []bool{false, true} {
		t.Run(fmt.Sprintf("disableTLS=%t", disableTLS), func(t *testing.T) {
			requestCanceled := make(chan struct{})
			handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(w, "error response")
				w.(http.Flusher).Flush()
				<-req.Context().Done()
				close(requestCanceled)
			})
			var server *httptest.Server
			if disableTLS {
				server = httptest.NewTLSServer(handler)
			} else {
				server = httptest.NewServer(handler)
			}
			t.Cleanup(server.Close)
			t.Cleanup(server.CloseClientConnections)

			c := client(disableTLS)
			original := c.Transport
			var body *trackedBody
			c.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				resp, err := original.RoundTrip(req)
				if err == nil {
					body = &trackedBody{ReadCloser: resp.Body}
					resp.Body = body
				}
				return resp, err
			})
			t.Cleanup(func() { c.Transport = original })

			resp, cancel, err := Get(server.URL, Context{Timeout: 60, DisableTLS: disableTLS})
			var statusError Error
			if !errors.As(err, &statusError) || statusError.StatusCode != http.StatusInternalServerError || resp != nil {
				t.Fatalf("Get returned response=%v, error=%v", resp, err)
			}
			if body.closes != 1 {
				t.Fatalf("body closed %d times before returning an error, want 1", body.closes)
			}
			select {
			case <-requestCanceled:
			case <-time.After(5 * time.Second):
				t.Fatal("server request was not canceled after the failed response")
			}
			cancel()
		})
	}
}
