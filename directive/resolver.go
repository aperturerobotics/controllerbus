package directive

// Resolvers constructs a resolver slice, ignoring nil entries.
func Resolvers(resolvers ...Resolver) []Resolver {
	out := make([]Resolver, 0, len(resolvers))
	for _, v := range resolvers {
		if v != nil {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// NewResolver checks the error and returns a resolver slice if err == nil.
func NewResolver(res Resolver, err error) ([]Resolver, error) {
	if err != nil {
		return nil, err
	}
	return Resolvers(res), nil
}
