package resource

var defaultResolver Resolver = &TemplateResolver{
	Resolver: SchemeResolver{
		Resolvers: NewDefaultSchemeResolvers(),
		NoScheme:  FileResolver{},
	},
}

// DefaultResolver returns the default Resolver implementation, which is a TemplateResolver that
// delegates to a default SchemeResolver.
func DefaultResolver() Resolver {
	return defaultResolver
}

// Resolve uses the default resolver to resolve a resource value.
func Resolve(v string) (Interface, error) {
	return defaultResolver.Resolve(v)
}

// MustResolve resolves a given resource string, panicing upon an error.
func MustResolve(r Resolver, v string) Interface {
	res, err := r.Resolve(v)
	if err != nil {
		panic(err)
	}

	return res
}
