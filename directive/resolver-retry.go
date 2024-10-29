package directive

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
)

// RetryResolver wraps a Resolver with retry logic.
type RetryResolver struct {
	// le is the logger
	le *logrus.Entry
	// res is the underlying resolver
	res Resolver
	// bo is the backoff
	bo backoff.BackOff
}

// NewRetryResolver constructs a new retry resolver.
func NewRetryResolver(le *logrus.Entry, res Resolver, bo backoff.BackOff) *RetryResolver {
	return &RetryResolver{le: le, res: res, bo: bo}
}

// Resolve resolves the values, emitting them to the handler.
func (r *RetryResolver) Resolve(ctx context.Context, handler ResolverHandler) error {
	for {
		err := r.res.Resolve(ctx, handler)
		if err == nil {
			r.bo.Reset()
			return nil
		}

		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		nextBackOff := r.bo.NextBackOff()
		r.le.
			WithError(err).
			Warnf("resolver returned error: backing off %s", nextBackOff.String())
		timer := time.NewTimer(nextBackOff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return context.Canceled
		case <-timer.C:
		}
	}
}

// _ is a type assertion
var _ Resolver = ((*RetryResolver)(nil))
