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

// Must panics if err != nil, returning the given resource handle otherwise
func Must(r Interface, err error) Interface {
	if err != nil {
		panic(err)
	}

	return r
}
