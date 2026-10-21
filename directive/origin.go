package directive

import "context"

// Origin identifies where a directive reference came from.
//
// AddDirective and AddReference read the Origin from ctx.
// The bus and directive controller set it on the contexts they create.
// Fields are empty if unset.
type Origin struct {
	// Directive is the directive instance whose handler, resolver, or started
	// controller added the reference. Nil when a root context added it.
	Directive Instance
	// ControllerID is the ID of the controller that added the reference.
	ControllerID string
	// ConfigKey is the ConfigSet entry key the adding controller runs under.
	ConfigKey string
}

// ReferenceInfo describes one live reference to a directive instance.
type ReferenceInfo struct {
	// Origin is where the reference came from.
	Origin Origin
	// Weak indicates the reference does not keep the instance alive.
	Weak bool
}

// originCtxKey is the context key for the Origin.
type originCtxKey struct{}

// WithOrigin returns a context carrying origin, replacing any origin on ctx.
func WithOrigin(ctx context.Context, origin Origin) context.Context {
	return context.WithValue(ctx, originCtxKey{}, origin)
}

// OriginFromContext returns the origin carried by ctx.
// Returns an empty Origin if ctx is nil or carries none.
func OriginFromContext(ctx context.Context) Origin {
	if ctx == nil {
		return Origin{}
	}
	origin, _ := ctx.Value(originCtxKey{}).(Origin)
	return origin
}

// WithOriginController returns a context whose origin names controllerID.
// The origin directive and config key are kept.
func WithOriginController(ctx context.Context, controllerID string) context.Context {
	origin := OriginFromContext(ctx)
	origin.ControllerID = controllerID
	return WithOrigin(ctx, origin)
}

// WithConfigKey returns a context whose origin names the ConfigSet entry key.
// The origin directive and controller are kept.
func WithConfigKey(ctx context.Context, configKey string) context.Context {
	origin := OriginFromContext(ctx)
	origin.ConfigKey = configKey
	return WithOrigin(ctx, origin)
}
