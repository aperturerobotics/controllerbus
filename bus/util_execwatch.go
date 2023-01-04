package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
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

// ExecOneOffWatch executes a one-off directive and watches for changes.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatch(
	b Bus,
	dir directive.Directive,
) (*ccontainer.CContainer[*directive.AttachedValue], directive.Reference, error) {
	ctr := ccontainer.NewCContainer[*directive.AttachedValue](nil)
	_, ref, err := b.AddDirective(
		dir,
		NewCallbackHandler(
			func(av directive.AttachedValue) {
				ctr.SetValue(&av)
			},
			func(av directive.AttachedValue) {
				ctr.SwapValue(func(val *directive.AttachedValue) *directive.AttachedValue {
					if val != nil && (*val).GetValueID() == av.GetValueID() {
						val = nil
					}
					return val
				})
			},
			func() {
				ctr.SetValue(nil)
			},
		),
	)
	return ctr, ref, err
}
