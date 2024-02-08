package bus

import (
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/aperturerobotics/util/routine"
)

// ExecOneOffWatchCb executes a one-off directive and watches for changes.
// Selects one value from the result.
// Calls the callback when the selected value changes.
// Calls with nil when the value becomes unset.
// If the callback returns false, the value is rejected, and the next value will be used instead.
func ExecOneOffWatchCb[T directive.ComparableValue](
	cb func(val directive.TypedAttachedValue[T]) bool,
	b Bus,
	dir directive.Directive,
) (directive.Instance, directive.Reference, error) {
	var currValueID uint32
	vals := make(map[uint32]directive.TypedAttachedValue[T], 1)

	di, ref, err := b.AddDirective(
		dir,
		NewCallbackHandler(
			func(av directive.AttachedValue) {
				val, ok := av.GetValue().(T)
				if !ok {
					return
				}
				vid := av.GetValueID()
				tav := directive.NewTypedAttachedValue[T](vid, val)
				vals[vid] = tav
				if currValueID == 0 {
					currValueID = vid
					if cb != nil && !cb(tav) {
						// Callback rejected the value
						delete(vals, vid)
						currValueID = 0
					}
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
				for valID, v := range vals {
					currValueID = valID
					if cb != nil && !cb(v) {
						// Callback rejected the value
						currValueID = 0
						delete(vals, valID)
						continue
					}
					break
				}
				if currValueID == 0 {
					cb(nil)
				}
			},
			func() {
				if currValueID != 0 {
					currValueID = 0
					cb(nil)
				}
			},
		),
	)

	return di, ref, err
}

// ExecOneOffWatchCtr executes a one-off directive and watches for changes.
// Stores the latest value into the ccontainer.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatchCtr[T directive.ComparableValue](
	ctr *ccontainer.CContainer[T],
	b Bus,
	dir directive.Directive,
) (
	directive.Instance,
	directive.Reference,
	error,
) {
	return ExecOneOffWatchCb[T](func(val directive.TypedAttachedValue[T]) bool {
		ctr.SetValue(val.GetValue())
		return true
	}, b, dir)
}

// ExecOneOffWatchRoutine executes a one-off directive and watches for changes.
//
// Calls SetState on the routine container.
// The routine will be restarted when the value changes.
// Note: the routine will not start until SetContext and SetStateRoutine are called on the routine container.
func ExecOneOffWatchRoutine[T directive.ComparableValue](
	routineCtr *routine.StateRoutineContainer[T],
	b Bus,
	dir directive.Directive,
) (directive.Instance, directive.Reference, error) {
	return ExecOneOffWatchCb[T](func(val directive.TypedAttachedValue[T]) bool {
		routineCtr.SetState(val.GetValue())
		return true
	}, b, dir)
}

// ExecOneOffWatchCh executes a one-off directive and watches for changes.
// Returns a channel with size 1 which will hold the latest value.
// Continues to watch for changes until the directive is released.
func ExecOneOffWatchCh[T directive.ComparableValue](
	b Bus,
	dir directive.Directive,
) (<-chan directive.TypedAttachedValue[T], directive.Instance, directive.Reference, error) {
	valCh := make(chan directive.TypedAttachedValue[T], 1)
	di, diRef, err := ExecOneOffWatchCb[T](func(val directive.TypedAttachedValue[T]) bool {
		for {
			select {
			case valCh <- val:
				return true
			default:
			}
			select {
			case <-valCh:
			default:
			}
		}
	}, b, dir)
	if err != nil {
		return nil, nil, nil, err
	}
	return valCh, di, diRef, nil
}
