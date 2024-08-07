package bus

import (
	"context"
	"slices"
	"sync"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/aperturerobotics/util/keyed"
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
				tav := directive.NewTypedAttachedValue(vid, val)
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
				if cb != nil && currValueID == 0 {
					cb(nil)
				}
			},
			func() {
				if currValueID != 0 {
					currValueID = 0
					if cb != nil {
						cb(nil)
					}
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
	return ExecOneOffWatchCb(func(val directive.TypedAttachedValue[T]) bool {
		if val == nil {
			var empty T
			ctr.SetValue(empty)
		} else {
			ctr.SetValue(val.GetValue())
		}
		return true
	}, b, dir)
}

// ExecOneOffWatchSelectCb executes a one-off directive and watches for changes.
//
// Uses the selectValue callback to select which value to pass to the callback.
// selectValue can return -1 to select none of the values.
// if selectValue is nil, the value with the lowest id will be selected.
//
// Calls the callback when the selected value changes.
// Calls with nil when the value becomes unset.
func ExecOneOffWatchSelectCb[T directive.ComparableValue](
	b Bus,
	dir directive.Directive,
	selectValue func(vals []directive.TypedAttachedValue[T]) int,
	cb func(val directive.TypedAttachedValue[T]),
) (directive.Instance, directive.Reference, error) {
	var sortedVals []directive.TypedAttachedValue[T]
	var currValueID uint32

	selectNextValue := func() {
		var selectedVal directive.TypedAttachedValue[T]
		if len(sortedVals) > 0 {
			selectedIdx := 0
			if selectValue != nil {
				selectedIdx = selectValue(sortedVals)
			}
			if selectedIdx >= 0 && selectedIdx < len(sortedVals) {
				selectedVal = sortedVals[selectedIdx]
			}
		}

		if selectedVal == nil {
			if currValueID != 0 {
				currValueID = 0
				if cb != nil {
					cb(nil)
				}
			}
		} else if selectedValID := selectedVal.GetValueID(); selectedValID != currValueID {
			currValueID = selectedValID
			if cb != nil {
				cb(selectedVal)
			}
		}
	}

	return b.AddDirective(
		dir,
		directive.NewTypedCallbackHandler(
			func(av directive.TypedAttachedValue[T]) {
				val := av.GetValue()
				vid := av.GetValueID()
				tav := directive.NewTypedAttachedValue(vid, val)
				sortedVals = append(sortedVals, tav)
				slices.SortFunc(sortedVals, func(a, b directive.TypedAttachedValue[T]) int {
					return int(a.GetValueID() - b.GetValueID())
				})
				selectNextValue()
			},
			func(av directive.TypedAttachedValue[T]) {
				vid := av.GetValueID()
				sortedVals = slices.DeleteFunc(sortedVals, func(v directive.TypedAttachedValue[T]) bool {
					return v.GetValueID() == vid
				})
				selectNextValue()
			},
			func() {
				// Clear all values
				sortedVals = nil
				if currValueID != 0 {
					currValueID = 0
					if cb != nil {
						cb(nil)
					}
				}
			},
			nil,
		),
	)
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
	return ExecOneOffWatchCb(func(val directive.TypedAttachedValue[T]) bool {
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
// Returns a nil value if the value becomes unset.
func ExecOneOffWatchCh[T directive.ComparableValue](
	b Bus,
	dir directive.Directive,
) (<-chan directive.TypedAttachedValue[T], directive.Instance, directive.Reference, error) {
	valCh := make(chan directive.TypedAttachedValue[T], 1)
	di, diRef, err := ExecOneOffWatchCb(func(val directive.TypedAttachedValue[T]) bool {
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

// ExecOneOffWatchLatestCb executes a one-off directive and watches for changes.
// The init callback is called with the directive instance and reference once set.
// The callback is called with the latest value or nil if the value is removed.
// The callback will not be called concurrently (only one instance at a time).
// Values may be dropped if the callback takes longer to execute than a value is available.
// Continues to watch for changes until the directive is released.
// Both callbacks are optional.
// Calls the callback with a nil value if the value becomes unset.
func ExecOneOffWatchLatestCb[T directive.ComparableValue](
	ctx context.Context,
	b Bus,
	dir directive.Directive,
	initCb func(di directive.Instance, ref directive.Reference) error,
	valCb func(val directive.TypedAttachedValue[T]) error,
) error {
	valCh, di, ref, err := ExecOneOffWatchCh[T](b, dir)
	if err != nil {
		return err
	}
	defer ref.Release()

	if initCb != nil {
		if err := initCb(di, ref); err != nil {
			return err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case val := <-valCh:
			if valCb != nil {
				if err := valCb(val); err != nil {
					return err
				}
			}
		}
	}
}

// ExecWatchEffect calls a callback for each value resolved for a directive.
//
// NOTE: The callbacks are run as part of the Directive Value callback and
// should return as quickly as possible to avoid blocking the other directive
// value callbacks.
//
// The callback can return an optional function to call when the value was removed.
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
				tav := directive.NewTypedAttachedValue(vid, val)
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

// ExecWatchTransformEffect calls a callback for each value resolved for a directive.
//
// The transform function will be called concurrently for each value resolved by the directive.
// The transform callback can return nil, false to reject the value.
// The transform function can return a transformed value.
//
// The callback will be called with the transformed values.
// The callback can return an optional function to call when the value was removed or changed.
// The callback will continue to be called until the ref is removed.
func ExecWatchTransformEffect[T, E directive.ComparableValue](
	ctx context.Context,
	transform func(ctx context.Context, val directive.TypedAttachedValue[T]) (E, bool, error),
	effect func(val directive.TransformedAttachedValue[T, E]) func(),
	b Bus,
	dir directive.Directive,
	keyedOpts ...keyed.Option[uint32, directive.TypedAttachedValue[T]],
) (directive.Instance, directive.Reference, error) {
	var mtx sync.Mutex
	vals := make(map[uint32]directive.TypedAttachedValue[T])
	rels := make(map[uint32]func())

	wrapTransform := func(val directive.TypedAttachedValue[T]) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			out, outOk, err := transform(ctx, val)
			if err != nil || !outOk {
				return err
			}

			if ctx.Err() != nil {
				return context.Canceled
			}

			xfrmValue := directive.NewTransformedAttachedValue(val, out)
			relFn := effect(xfrmValue)
			if relFn == nil {
				return nil
			}
			if ctx.Err() != nil {
				relFn()
				return context.Canceled
			}

			valueID := val.GetValueID()
			mtx.Lock()
			if vals[valueID] == val {
				rels[valueID] = relFn
			} else {
				defer relFn()
			}
			mtx.Unlock()

			return nil
		}
	}

	transforms := keyed.NewKeyed(
		func(key uint32) (keyed.Routine, directive.TypedAttachedValue[T]) {
			// NOTE: no mtx.Lock here since we already locked before SetKey below!
			val, exists := vals[key]
			if !exists {
				return nil, val
			}
			return wrapTransform(val), val
		},
		keyedOpts...,
	)
	transforms.SetContext(ctx, true)

	return ExecWatchEffect(func(val directive.TypedAttachedValue[T]) func() {
		mtx.Lock()
		vals[val.GetValueID()] = val
		transforms.SetKey(val.GetValueID(), true)
		mtx.Unlock()

		return func() {
			valueID := val.GetValueID()
			mtx.Lock()
			delete(vals, valueID)
			transforms.RemoveKey(valueID)
			relFn := rels[valueID]
			if relFn != nil {
				delete(rels, valueID)
				defer relFn()
			}
			mtx.Unlock()
		}
	}, b, dir)
}

// ExecOneOffWatchEffect calls a callback for a single value resolved for a directive.
//
// The callback can return an optional function to call when the value was removed.
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

// ExecOneOffWatchTransformEffect calls a callback for a single value resolved for a directive.
//
// The transform function will be called concurrently for each value resolved by the directive.
// The transform callback can return nil, false to reject the value.
// The transform function can return a transformed value.
//
// The select function is called with all values produced by the transform
// functions so far, and should select which value to emit to the effect based
// on a stable sort. It can return an index in the slice or -1 to select none.
//
// If selectValue is nil, the first returned value (lowest value id) will be used.
//
// The callback will be called with the most recently selected value.
// The callback can return an optional function to call when the value was removed or changed.
// The callback will continue to be called until the ref is removed.
func ExecOneOffWatchTransformEffect[T, E directive.ComparableValue](
	ctx context.Context,
	transform func(ctx context.Context, val directive.TypedAttachedValue[T]) (E, bool, error),
	selectValue func(vals []directive.TransformedAttachedValue[T, E]) int,
	effect func(val directive.TransformedAttachedValue[T, E]) func(),
	b Bus,
	dir directive.Directive,
	keyedOpts ...keyed.Option[uint32, directive.TypedAttachedValue[T]],
) (directive.Instance, directive.Reference, error) {
	var mtx sync.Mutex
	var xfrmVals []directive.TransformedAttachedValue[T, E]
	sortXfrmVals := func() {
		slices.SortFunc(xfrmVals, func(a, b directive.TransformedAttachedValue[T, E]) int {
			return int(a.GetValueID() - b.GetValueID())
		})
	}

	var effectVal uint32
	var effectRel func()
	maybeCallEffect := func() {
		var selectedValID uint32
		var selectedVal directive.TransformedAttachedValue[T, E]

		var valueIdx int
		if selectValue != nil {
			valueIdx = selectValue(xfrmVals)
		}
		if valueIdx >= 0 && valueIdx < len(xfrmVals) {
			selectedVal = xfrmVals[valueIdx]
			selectedValID = selectedVal.GetValueID()
		}
		if selectedValID == effectVal {
			// no change
			return
		}
		if effectRel != nil {
			effectRel()
		}
		effectVal = selectedValID
		effectRel = effect(selectedVal)
	}

	return ExecWatchTransformEffect(
		ctx,
		transform,
		func(val directive.TransformedAttachedValue[T, E]) func() {
			mtx.Lock()
			xfrmVals = append(xfrmVals, val)
			sortXfrmVals()
			maybeCallEffect()
			mtx.Unlock()

			return func() {
				mtx.Lock()
				idx := slices.Index(xfrmVals, val)
				if idx >= 0 {
					xfrmVals = slices.Delete(xfrmVals, idx, idx+1)
					if effectVal == val.GetValueID() {
						effectVal = 0
						if effectRel != nil {
							effectRel()
							effectRel = nil
						}
					}
				}
				mtx.Unlock()
			}
		},
		b,
		dir,
		keyedOpts...,
	)
}
