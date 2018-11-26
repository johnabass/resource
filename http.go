package resource

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"time"
)

// HTTPClient is the method set expected of an object which can transact with an HTTP server.
// http.Client implements this interface.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type HTTPClientFunc func(*http.Request) (*http.Response, error)

func (f HTTPClientFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func WithMethod(method string, c HTTPClient) HTTPClient {
	return HTTPClientFunc(func(request *http.Request) (*http.Response, error) {
		request.Method = method
		return c.Do(request)
	})
}

func WithHeader(name, value string, c HTTPClient) HTTPClient {
	return HTTPClientFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header == nil {
			request.Header = make(http.Header)
		}

		request.Header.Set(name, value)
		return c.Do(request)
	})
}

func WithHeaders(h http.Header, c HTTPClient) HTTPClient {
	if len(h) == 0 {
		return c
	}

	// make a safe, deep copy to use
	clone := make(http.Header, len(h))
	for k, v := range h {
		k = textproto.CanonicalMIMEHeaderKey(k)
		clone[k] = append(clone[k], v...)
	}

	h = clone
	return HTTPClientFunc(func(request *http.Request) (*http.Response, error) {
		if request.Header == nil {
			request.Header = make(http.Header, len(h))
		}

		for k, v := range h {
			request.Header[k] = v
		}

		return c.Do(request)
	})
}

func WithClose(c HTTPClient) HTTPClient {
	return HTTPClientFunc(func(request *http.Request) (*http.Response, error) {
		request.Close = true
		return c.Do(request)
	})
}

func WithTimeout(d time.Duration, c HTTPClient) HTTPClient {
	return HTTPClientFunc(func(request *http.Request) (*http.Response, error) {
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()

		return c.Do(request.WithContext(ctx))
	})
}

type drainOnClose struct {
	io.ReadCloser
}

func (doc drainOnClose) Close() error {
	io.Copy(ioutil.Discard, doc.ReadCloser)
	return doc.ReadCloser.Close()
}

// DrainOnClose decorates an existing io.ReadCloser such that invoking Close on the
// returned io.ReadCloser causes any remaining contents to be discarded.  Used by HTTP
// resources to ensure that the HTTP response body is fully read when client code closes
// the resource's io.ReadCloser returned by Open.
func DrainOnClose(rc io.ReadCloser) io.ReadCloser {
	return drainOnClose{rc}
}

// HTTPError represents a failure to obtain an HTTP resource.  This error indicates that a successful
// HTTP transaction occurred with a non-2XX response code.
type HTTPError struct {
	URL  string
	Code int
}

func (he HTTPError) Error() string {
	return fmt.Sprintf("HTTP resource %s failed with status code %d", he.URL, he.Code)
}

// StatusCode is supplied to implement go-kit's StatusCoder interface
func (he HTTPError) StatusCode() int {
	return he.Code
}
