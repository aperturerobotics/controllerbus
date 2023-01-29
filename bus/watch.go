package bus

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/aperturerobotics/util/routine"
)

// ExecOneOffWatchCh executes a one-off directive and watches for changes.
// Returns a channel with size 1 which will hold the latest value.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatchCh(
	b Bus,
	dir directive.Directive,
) (<-chan directive.AttachedValue, directive.Instance, directive.Reference, error) {
	valCh := make(chan directive.AttachedValue, 1)
	di, ref, err := b.AddDirective(
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
	return valCh, di, ref, err
}

// ExecOneOffWatch executes a one-off directive and watches for changes.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatch(
	b Bus,
	dir directive.Directive,
) (*ccontainer.CContainer[*directive.AttachedValue], directive.Instance, directive.Reference, error) {
	ctr := ccontainer.NewCContainer[*directive.AttachedValue](nil)
	di, ref, err := b.AddDirective(
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
	return ctr, di, ref, err
}

// ExecOneOffWatchRoutine executes a one-off directive and watches for changes.
//
// The routine will not start until SetContext is called.
// The routine will be restarted when the value changes.
// If the routine returns nil or any error, the function returns that value.
// If the routine context is canceled, return context.Canceled.
// If the directive is disposed, the routine will return ErrDirectiveDisposed.
func ExecOneOffWatchRoutine[T directive.Value](
	b Bus,
	dir directive.Directive,
	cb func(ctx context.Context, value T) error,
) (*routine.RoutineContainer, directive.Instance, directive.Reference, error) {
	var mtx sync.Mutex
	var currValueID uint32

	routineCtr := routine.NewRoutineContainer()
	di, ref, err := b.AddDirective(
		dir,
		NewCallbackHandler(
			func(av directive.AttachedValue) {
				val, ok := av.GetValue().(T)
				if !ok {
					return
				}
				mtx.Lock()
				currValueID = av.GetValueID()
				routineCtr.SetRoutine(func(ctx context.Context) error {
					return cb(ctx, val)
				})
				mtx.Unlock()
			},
			func(av directive.AttachedValue) {
				mtx.Lock()
				if currValueID == av.GetValueID() {
					routineCtr.SetRoutine(nil)
				}
				mtx.Unlock()
			},
			func() {
				mtx.Lock()
				currValueID = 0
				routineCtr.SetRoutine(func(ctx context.Context) error {
					return ErrDirectiveDisposed
				})
				mtx.Unlock()
			},
		),
	)
	return routineCtr, di, ref, err
}
