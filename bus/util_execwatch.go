package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecOneOffWatchCh executes a one-off directive and watches for changes.
// Returns a channel with size 1 which will hold the latest value.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatchCh(
	b Bus,
	dir directive.Directive,
) (<-chan directive.AttachedValue, directive.Reference, error) {
	valCh := make(chan directive.AttachedValue, 1)
	_, ref, err := b.AddDirective(
		dir,
		NewCallbackHandler(
			func(av directive.AttachedValue) {
			PushLoop:
				for {
					select {
					case valCh <- av:
						break PushLoop
					default:
					}
					select {
					case <-valCh:
						// remove old value
					default:
					}
				}
			},
			nil,
			func() {
				close(valCh)
			},
		),
	)
	return valCh, ref, err
}
