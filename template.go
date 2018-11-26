package resource

import (
	"bytes"
	"errors"
	"html/template"
	"os"
	"sync"
)

var ErrTooManyDefaults = errors.New("Too many default values")

// DefaultEnvFunc is the default key in a template.FuncMap that maps to Getenv.
const DefaultEnvFunc = "env"

// Getenv is an analog to os.Getenv that allows for an optional default value to be used when
// the given environment variable is not present.  Supplying multiple default values raises an error.
func Getenv(key string, def ...string) (string, error) {
	if len(def) > 1 {
		return "", ErrTooManyDefaults
	}

	v := os.Getenv(key)
	if len(v) > 0 {
		return v, nil
	}

	if len(def) > 0 {
		return def[0], nil
	}

	return "", nil
}

// ConfigureTemplateDefaults is used to set up the defaults for a resource template.  This function is
// useful when using an arbitrary template as the parent for parsing.
func ConfigureTemplateDefaults(t *template.Template) *template.Template {
	return t.Funcs(template.FuncMap{DefaultEnvFunc: Getenv})
}

// TemplateResolver is a decorator that expands resource strings as text templates and passes the
// results to another Resolver.  An arbitrary template can be used for parsing, which allows customization
// of delimiters, functions, etc.
type TemplateResolver struct {
	// Resolver is the decorated Resolver.  This resolver will receive expanded resource strings.
	// This field is required.
	Resolver Resolver

	// Template is the optional template for parsing.  If supplied, this template's Parse method is used
	// to expand resource strings.  If not supplied, a simple default template is created and used each
	// time a resource string needs to be resolved.
	//
	// When supplied, this template's Parse method is guarded by an internal mutex.  Care must be taken that
	// the same template is not used with multiple TemplateResolver instances.  It's safest to use template.Clone
	// for each TemplateResolver's Template field, thus ensuring no race conditions will occur.
	Template *template.Template

	// Data is the optional data passed to each template execution.  If supplied, this value is passed as is
	// to template.Execute.
	Data interface{}

	parseLock sync.Mutex
}

func (tr *TemplateResolver) parse(v string) (t *template.Template, err error) {
	if tr.Template != nil {
		tr.parseLock.Lock()
		t, err = tr.Template.Parse(v)
		tr.parseLock.Unlock()
	} else {
		t, err = ConfigureTemplateDefaults(template.New("")).Parse(v)
	}

	return
}

// Resolve expands v using the configured templating (or a default) and passes the result
// to the decorated Resolver.
func (tr *TemplateResolver) Resolve(v string) (Interface, error) {
	t, err := tr.parse(v)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	if err := t.Execute(&output, tr.Data); err != nil {
		return nil, err
	}

	return tr.Resolver.Resolve(output.String())
}
