package resource

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Interface represents a handle to a resource.
//
// All resource handles implement io.WriterTo, giving each type of resource the ability
// to optimize how transfers of data occur.  For example, in-memory resources can perform
// a simple copy without creating intermediate objects.
type Interface interface {
	io.WriterTo

	Location() string
	Open() (io.ReadCloser, error)
}

// String represents an in-memory resource backed by a golang string.
type String string

func (s String) Location() string {
	return "string"
}

func (s String) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader(string(s))), nil
}

func (s String) WriteTo(w io.Writer) (int64, error) {
	count, err := io.WriteString(w, string(s))
	return int64(count), err
}

// Bytes represents an in-memory resource backed by a byte slice.
type Bytes []byte

func (b Bytes) Location() string {
	return "bytes"
}

func (b Bytes) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(bytes.NewReader([]byte(b))), nil
}

func (b Bytes) WriteTo(w io.Writer) (int64, error) {
	count, err := w.Write([]byte(b))
	return int64(count), err
}

// File represents a resource backed by a system file.
type File string

func (f File) Location() string {
	return string(f)
}

func (f File) Open() (io.ReadCloser, error) {
	return os.Open(string(f))
}

func (f File) WriteTo(w io.Writer) (int64, error) {
	rc, err := f.Open()
	if err != nil {
		return int64(0), err
	}

	defer rc.Close()
	count, err := io.Copy(w, rc)
	return int64(count), err
}

// HTTP represents a resource backed by an HTTP or HTTPS URL.
type HTTP struct {
	// URL is the required URL of the resource
	URL string

	// OpenMethod is the HTTP verb used to request the resource's data.  If not
	// supplied, GET is used.
	OpenMethod string

	// Client is the HTTP client to use to obtain the resource.  If not supplied,
	// http.DefaultClient is used.
	Client HTTPClient
}

func (h HTTP) Location() string {
	return h.URL
}

// transact performs an HTTP transaction using this resource's configuration
func (h HTTP) transact() (*http.Response, error) {
	method := h.OpenMethod
	if len(method) == 0 {
		method = http.MethodGet
	}

	request, err := http.NewRequest(method, h.URL, nil)
	if err != nil {
		return nil, err
	}

	c := h.Client
	if c == nil {
		c = http.DefaultClient
	}

	return c.Do(request)
}

func (h HTTP) Open() (io.ReadCloser, error) {
	response, err := h.transact()
	if err != nil {
		return nil, err
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		io.Copy(ioutil.Discard, response.Body)
		response.Body.Close()
		return nil, HTTPError{h.URL, response.StatusCode}
	}

	return DrainOnClose(response.Body), nil
}

func (h HTTP) WriteTo(w io.Writer) (int64, error) {
	response, err := h.transact()
	if err != nil {
		return int64(0), err
	}

	defer response.Body.Close()
	return io.Copy(w, response.Body)
}
