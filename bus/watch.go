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
		if val == nil {
			var empty T
			ctr.SetValue(empty)
		} else {
			ctr.SetValue(val.GetValue())
		}
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
		if val == nil {
			var empty T
			routineCtr.SetState(empty)
		} else {
			routineCtr.SetState(val.GetValue())
		}
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

// ExecWatchEffect calls a callback for each value resolved for a directive.
//
// NOTE: The callbacks are run as part of the Directive Value callback and
// should return as quickly as possible to avoid blocking the other directive
// value callbacks.
//
// The callback can return an optional function to call when the value was removed.
// The callback can return an error to terminate the watch.
// The callback will continue to be called until the ref is removed.
func ExecWatchEffect[T directive.ComparableValue](
	cb func(val directive.TypedAttachedValue[T]) func(),
	b Bus,
	dir directive.Directive,
) (directive.Instance, directive.Reference, error) {
	valRels := make(map[uint32]func(), 1)

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
				rel := cb(tav)
				if rel != nil {
					valRels[vid] = rel
				}
			},
			func(av directive.AttachedValue) {
				_, ok := av.GetValue().(T)
				if !ok {
					return
				}
				vid := av.GetValueID()
				rel := valRels[vid]
				if rel != nil {
					delete(valRels, vid)
					rel()
				}
			},
			func() {
				for k, v := range valRels {
					v()
					delete(valRels, k)
				}
			},
		),
	)

	return di, ref, err
}

// ExecWatchTransformEffect calls a callback to transform each value resolved
// for a directive, then calls a function with each transformed value.
//
// NOTE: The callbacks are run as part of the Directive Value callback and
// should return as quickly as possible to avoid blocking the other directive
// value callbacks.
//
// The transform callback can return nil, false to reject the value.
// The callback can return an optional function to call when the value was removed.
// The callback can return an error to terminate the watch.
// The callback will continue to be called until the ref is removed.
func ExecWatchTransformEffect[T, E directive.ComparableValue](
	transform func(val directive.TypedAttachedValue[T]) (E, bool),
	effect func(val directive.TypedAttachedValue[T], xfrm E) func(),
	b Bus,
	dir directive.Directive,
) (directive.Instance, directive.Reference, error) {
	valRels := make(map[uint32]func(), 1)
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
				xfrm, ok := transform(tav)
				if !ok {
					return
				}
				rel := effect(tav, xfrm)
				if rel != nil {
					valRels[vid] = rel
				}
			},
			func(av directive.AttachedValue) {
				_, ok := av.GetValue().(T)
				if !ok {
					return
				}
				vid := av.GetValueID()
				rel := valRels[vid]
				if rel != nil {
					delete(valRels, vid)
					rel()
				}
			},
			func() {
				for k, v := range valRels {
					v()
					delete(valRels, k)
				}
			},
		),
	)
	return di, ref, err
}

// ExecOneOffWatchEffect calls a callback for a single value resolved for a directive.
//
// The callback can return an optional function to call when the value was removed.
// The callback can return an error to terminate the watch.
// The callback will continue to be called until the ref is removed.
func ExecOneOffWatchEffect[T directive.ComparableValue](
	cb func(val directive.TypedAttachedValue[T]) func(),
	b Bus,
	dir directive.Directive,
) (directive.Instance, directive.Reference, error) {
	var currRel func()
	return ExecOneOffWatchCb(func(val directive.TypedAttachedValue[T]) bool {
		if currRel != nil {
			currRel()
			currRel = nil
		}
		if val != nil {
			currRel = cb(val)
		}
		return true
	}, b, dir)
}
