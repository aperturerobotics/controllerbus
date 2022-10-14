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
