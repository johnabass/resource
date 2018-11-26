package resource

import (
	"encoding/base64"
	"net/url"
	"path/filepath"
)

// Resolver is the strategy used to turn strings into resource handles.
type Resolver interface {
	// Resolve accepts a resource string and returns a handle that can load data
	// from that resource.  All Resolver implementations must safely permit this
	// method to be called concurrently.
	Resolve(string) (Interface, error)
}

// ResolverFunc is a function type that can resolve resources
type ResolverFunc func(string) (Interface, error)

func (rf ResolverFunc) Resolve(v string) (Interface, error) {
	return rf(v)
}

// StringResolver resolves values as in-memory strings rather than external locations.
// This resolver ignores any scheme associated with the value, allowing it to be mapped
// to any desired scheme.
type StringResolver struct{}

func (r StringResolver) Resolve(v string) (Interface, error) {
	_, v = Split(v)
	return String(v), nil
}

// BytesResolver resolves values as in-memory bytes encoding as base64 strings.  The choice
// of encoding is configurable, and defaults to base64.StdEncoding.  Any scheme is ignored by
// this resolver, allowing it to be mapped to any desired scheme.
type BytesResolver struct {
	// Encoding is the base64 encoding to use.  If not supplied, base64.StdEncoding is used.
	Encoding *base64.Encoding
}

func (r BytesResolver) Resolve(v string) (Interface, error) {
	enc := r.Encoding
	if enc == nil {
		enc = base64.StdEncoding
	}

	_, v = Split(v)
	b, err := enc.DecodeString(v)
	if err != nil {
		return nil, err
	}

	return Bytes(b), nil
}

// FileResolver resolves values as file system paths, relative to an optional Root directory.
// Any scheme is ignored by this resolver.
type FileResolver struct {
	// Root is the optional file system path that acts as the logical root directory
	// for any resource strings this instance resolves.  If not supplied, no root is assumed.
	Root string
}

func (r FileResolver) Resolve(v string) (Interface, error) {
	_, v = Split(v)
	p, err := filepath.Abs(filepath.Join(r.Root, v))
	if err != nil {
		return nil, err
	}

	return File(p), nil
}

// HTTPResolver uses an HTTP client to resolve resources.  Resource strings are expected to be
// valid URIs resolvable by the net/http package.
type HTTPResolver struct {
	OpenMethod string
	Client     HTTPClient
}

func (r HTTPResolver) Resolve(v string) (Interface, error) {
	if _, err := url.Parse(v); err != nil {
		return nil, err
	}

	return HTTP{URL: v, OpenMethod: r.OpenMethod, Client: r.Client}, nil
}

// Resolvers represents a mapping of component resolvers by an arbitrary string key.
// The most common usage is looking up a resolver by the scheme that it is mapped to.
type Resolvers map[string]Resolver

func (rs Resolvers) Get(k string) (Resolver, bool) {
	if len(rs) > 0 {
		r, ok := rs[k]
		return r, ok
	}

	return nil, false
}

func (rs *Resolvers) Set(k string, r Resolver) {
	if rs == nil {
		*rs = make(Resolvers)
	}

	(*rs)[k] = r
}
