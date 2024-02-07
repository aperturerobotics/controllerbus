package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/aperturerobotics/util/routine"
)

// ExecOneOffWatchCtr executes a one-off directive and watches for changes.
// Stores the latest value into the ccontainer.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatchCtr(
	ctr *ccontainer.CContainer[*directive.AttachedValue],
	b Bus,
	dir directive.Directive,
) (
	directive.Instance,
	directive.Reference,
	error,
) {
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
	return di, ref, err
}

// ExecOneOffWatchRoutine executes a one-off directive and watches for changes.
//
// Calls SetState on the routine container.
// The routine will be restarted when the value changes.
// The routine will not start until SetContext is called on the routine container.
// If the routine returns nil or any error, the function returns that value.
// If the routine context is canceled, return context.Canceled.
// If the directive is disposed, the routine will return ErrDirectiveDisposed.
func ExecOneOffWatchRoutine[T directive.ComparableValue](
	routineCtr *routine.StateRoutineContainer[T],
	b Bus,
	dir directive.Directive,
) (directive.Instance, directive.Reference, error) {
	type attachedValue struct {
		vid uint32
		val T
	}
	var currValueID uint32
	vals := make(map[uint32]attachedValue, 1)

	di, ref, err := b.AddDirective(
		dir,
		NewCallbackHandler(
			func(av directive.AttachedValue) {
				val, ok := av.GetValue().(T)
				if !ok {
					return
				}
				vid := av.GetValueID()
				vals[vid] = attachedValue{vid: vid, val: val}
				if currValueID == 0 {
					currValueID = vid
					routineCtr.SetState(val)
				}
			},
			func(av directive.AttachedValue) {
				_, ok := av.GetValue().(T)
				if !ok {
					return
				}
				vid := av.GetValueID()
				delete(vals, vid)
				if currValueID != vid {
					return
				}
				currValueID = 0
				for _, v := range vals {
					routineCtr.SetState(v.val)
					currValueID = v.vid
					break
				}
				if currValueID == 0 {
					var empty T
					routineCtr.SetState(empty)
				}
			},
			func() {
				currValueID = 0
				var empty T
				routineCtr.SetState(empty)
			},
		),
	)

	return di, ref, err
}

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
