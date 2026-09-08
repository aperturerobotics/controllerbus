package directive

import (
	"context"
	"errors"
	"testing"

	backoff "github.com/aperturerobotics/util/backoff/cbackoff"
)

// retryTestResolver returns a terminal resolver error.
type retryTestResolver struct {
	// err is the resolver result.
	err error
}

// Resolve returns the configured error.
func (r retryTestResolver) Resolve(context.Context, ResolverHandler) error {
	return r.err
}

// TestRetryResolverStopsAtBackoffStop checks terminal backoff propagation.
func TestRetryResolverStopsAtBackoffStop(t *testing.T) {
	wantErr := errors.New("resolver failed")
	resolver := NewRetryResolver(nil, retryTestResolver{err: wantErr}, &backoff.StopBackOff{})
	err := resolver.Resolve(t.Context(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("resolver error = %v, want %v", err, wantErr)
	}
}
