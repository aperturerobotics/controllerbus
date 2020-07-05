package debounce_fswatcher

import (
	"context"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

// DebounceFSWatcherEvents debounces a fswatcher event stream.
// Waits for a "quiet period" before syncing.
// Returns when an event happened.
func DebounceFSWatcherEvents(
	ctx context.Context,
	watcher *fsnotify.Watcher,
	debounceDur time.Duration,
) ([]fsnotify.Event, error) {
	var happened []fsnotify.Event
	var nextSyncTicker *time.Timer
	var nextSyncC <-chan time.Time
	defer func() {
		if nextSyncTicker != nil {
			nextSyncTicker.Stop()
		}
	}()
	// flush first
FlushLoop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case _, ok := <-watcher.Events:
			if !ok {
				return nil, nil
			}
		default:
			break FlushLoop
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return happened, nil
			}
			switch event.Op {
			case fsnotify.Create:
			case fsnotify.Rename:
			case fsnotify.Write:
			case fsnotify.Remove:
			default:
				continue
			}
			happened = append(happened, event)
			if nextSyncTicker != nil {
				nextSyncTicker.Stop()
			}
			nextSyncTicker = time.NewTimer(debounceDur)
			nextSyncC = nextSyncTicker.C
		case err, ok := <-watcher.Errors:
			if !ok || err == context.Canceled {
				return happened, nil
			}
			return nil, errors.Wrap(err, "watcher error")
		case <-nextSyncC:
			nextSyncTicker = nil
			return happened, nil
		}
	}
}
