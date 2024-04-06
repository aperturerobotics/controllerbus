package directive

import "context"

// NewErrChIdleCallback builds an IdleCallback which writes an error to a channel
// if any of the resolvers returned any error other than context.Canceled.
//
// If waitIdle is set, waits for the directive to be idle before checking errs.
// Skips writing to the channel if it is full, use a buffered channel of size 1.
func NewErrChIdleCallback(errCh chan<- error, waitIdle bool) IdleCallback {
	return func(isIdle bool, resolverErrs []error) {
		if waitIdle && !isIdle {
			return
		}
		for _, err := range resolverErrs {
			if err != nil && err != context.Canceled {
				select {
				case errCh <- err:
				default:
				}
			}
		}
	}
}
