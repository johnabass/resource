package resource

import (
	"fmt"
	"strings"
)

const (
	SchemeSeparator = "://"

	StringScheme = "string"
	BytesScheme  = "bytes"
	FileScheme   = "file"
	HTTPScheme   = "http"
	HTTPSScheme  = "https"
)

// Split parses a resource value into its scheme and value.
// This is much more relaxed than url.Parse, as it permits characters
// that are not allowed in URIs.  This fact is importent when allowing
// arbitrary strings or bytes as in-memory resources.
func Split(v string) (scheme, value string) {
	if i := strings.Index(v, SchemeSeparator); i >= 0 {
		return v[0:i], v[i+len(SchemeSeparator):]
	}

	return "", v
}

// DefaultSchemeResolvers produces a Resolvers with the default scheme mappings.
// These mappings are:
//
//   StringScheme is mapped to a StringResolver
//   BytesScheme is mapped to a BytesResolver with standard base64 encoding
//   FileScheme is mapped to a FileResolver with no relative path
//   HTTPScheme and HTTPSScheme are mapped to an HTTPResolver using the default HTTP Client
//
// When constructing custom SchemeResolver instances, this function is useful as a starting point.
func DefaultSchemeResolvers() Resolvers {
	var (
		fr = FileResolver{}
		hr = HTTPResolver{}
	)

	return Resolvers{
		StringScheme: StringResolver{},
		BytesScheme:  BytesResolver{},
		FileScheme:   fr,
		HTTPScheme:   hr,
		HTTPSScheme:  hr,
	}
}

// SchemeError is returned when a scheme had no associated resolver.
type SchemeError struct {
	Value  string
	Scheme string
}

func (e SchemeError) Error() string {
	return fmt.Sprintf("Cannot resolve %s: no resolver registered for scheme %s", e.Value, e.Scheme)
}

// NoSchemeError is returned when no scheme was supplied on a resource and the SchemeResolver
// had no NoScheme resolver configured.
type NoSchemeError struct {
	Value string
}

func (e NoSchemeError) Error() string {
	return fmt.Sprintf("Cannot resolve %s: no scheme supplied", e.Value)
}

// SchemeResolver is a resource resolver that uses URI-style schemes to determine how to
// resolve resources.  For example, "http://localhost/foo" is a resource value with the scheme "http".
// Resource values resolved by this type of resolver do not have to be well-formed URIs unless the
// component resolver expects that.  For example, "string://hello world" would resolve to a string
// resource when using the default scheme resolver.
type SchemeResolver struct {
	Resolvers Resolvers
	NoScheme  Resolver
}

func (sr SchemeResolver) Resolve(v string) (Interface, error) {
	if scheme, _ := Split(v); len(scheme) > 0 {
		resolver, ok := sr.Resolvers.Get(scheme)
		if !ok {
			return nil, SchemeError{Value: v, Scheme: scheme}
		}

		return resolver.Resolve(v)
	}

	if sr.NoScheme == nil {
		return nil, NoSchemeError{Value: v}
	}

	return sr.NoScheme.Resolve(v)
}
