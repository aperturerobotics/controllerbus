package directive

import (
	"context"
	"errors"
	"testing"

	backoff "github.com/aperturerobotics/util/backoff/cbackoff"
	"github.com/stretchr/testify/require"
)

type retryTestResolver struct{ err error }

func (r retryTestResolver) Resolve(context.Context, ResolverHandler) error { return r.err }

func TestRetryResolverStopsAtBackoffStop(t *testing.T) {
	wantErr := errors.New("resolver failed")
	resolver := NewRetryResolver(nil, retryTestResolver{err: wantErr}, &backoff.StopBackOff{})
	err := resolver.Resolve(context.Background(), nil)
	require.ErrorIs(t, err, wantErr)
}
