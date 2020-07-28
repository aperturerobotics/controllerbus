package controller_exec

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
)

// ExecuteController executes a controller and calls the callback with state.
func ExecuteController(
	ctx context.Context,
	cbus bus.Bus,
	conf config.Config,
	cb func(ControllerStatus),
) error {
	if cb == nil {
		cb = func(ControllerStatus) {}
	}
	dir := resolver.NewLoadControllerWithConfig(conf)

	cb(ControllerStatus_ControllerStatus_CONFIGURING)
	_, valRef, err := bus.ExecOneOff(ctx, cbus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	cb(ControllerStatus_ControllerStatus_RUNNING)
	<-ctx.Done()
	return nil
}

// Execute executes the request to apply a config set.
// Cb should not hold ExecControllerResponse after returning.
func (r *ExecControllerRequest) Execute(
	ctx context.Context,
	cbus bus.Bus,
	allowPartialSuccess bool,
	cb func(*ExecControllerResponse) error,
) error {
	var resp ExecControllerResponse
	var err error
	// callCb calls the callback.
	callCb := func() error {
		if cb != nil {
			return cb(&resp)
		}
		return nil
	}

	confsList := r.GetConfigSet().GetConfigurations()
	var confSet configset.ConfigSet
	prevStates := make(map[string]ControllerStatus, len(confsList))
	if !allowPartialSuccess {
		confSet, err = r.GetConfigSet().Resolve(ctx, cbus)
	}
	if err != nil {
		return err
	}

	if confSet == nil {
		confSet = make(configset.ConfigSet, len(confsList))
	}
	niniterr := 0
	var lastniniterr error
	for csID, conf := range confsList {
		if csID == "" {
			continue
		}

		ns := ControllerStatus_ControllerStatus_CONFIGURING
		if _, ok := confSet[csID]; !ok {
			confSet[csID], err = conf.Resolve(ctx, cbus)
			if err != nil {
				ns = ControllerStatus_ControllerStatus_ERROR
				niniterr++
				lastniniterr = err
				delete(confSet, csID)
				resp.ErrorInfo = err.Error()
			} else {
				resp.ErrorInfo = ""
			}
		} else {
			resp.ErrorInfo = ""
		}

		resp.Id = csID
		resp.Status = ns
		if err := callCb(); err != nil {
			return err
		}
		prevStates[csID] = ns
	}
	if niniterr == len(confsList) && len(confsList) != 0 {
		if len(confsList) == 1 && lastniniterr != nil {
			return lastniniterr
		}
		return ErrAllControllersFailed
	}
	resp.Reset()

	// handle results of configset controller apply.
	subCtx, subCtxCancel := context.WithCancel(ctx)
	defer subCtxCancel()

	addedCh := make(chan configset.ApplyConfigSetValue)
	removedCh := make(chan configset.ApplyConfigSetValue)

	_, dirRef, err := cbus.AddDirective(
		configset.NewApplyConfigSet(confSet),
		bus.NewCallbackHandler(
			// value added
			func(val directive.AttachedValue) {
				csVal, csValOk := val.GetValue().(configset.ApplyConfigSetValue)
				if !csValOk || csVal == nil || csVal.GetId() == "" {
					return
				}
				select {
				case <-subCtx.Done():
					return
				case addedCh <- csVal:
				}
			},
			// value removed
			func(val directive.AttachedValue) {
				csVal, csValOk := val.GetValue().(configset.ApplyConfigSetValue)
				if !csValOk || csVal == nil || csVal.GetId() == "" {
					return
				}
				select {
				case <-subCtx.Done():
					return
				case removedCh <- csVal:
				}
			},
			subCtxCancel,
		),
	)
	if err != nil {
		return err
	}
	defer dirRef.Release()

	for {
		select {
		case <-subCtx.Done():
			return subCtx.Err()
		case csv := <-addedCh:
			csvID := csv.GetId()
			csvErr := csv.GetError()
			resp.Id = csvID
			if csvErr != nil {
				resp.Status = ControllerStatus_ControllerStatus_ERROR
				resp.ErrorInfo = csvErr.Error()
			} else if ctrl := csv.GetController(); ctrl != nil {
				ctrlInfo := ctrl.GetControllerInfo()
				resp.Status = ControllerStatus_ControllerStatus_RUNNING
				resp.ControllerInfo = &ctrlInfo
			} else {
				resp.Status = ControllerStatus_ControllerStatus_CONFIGURING
			}
			if prevStates[csvID] != resp.Status ||
				resp.Status != ControllerStatus_ControllerStatus_CONFIGURING {
				prevStates[csvID] = resp.Status
				if err := callCb(); err != nil {
					return err
				}
			}
			resp.Reset()
		case csv := <-removedCh:
			// removed == value is no longer applicable.
			csvID := csv.GetId()
			resp.Status = ControllerStatus_ControllerStatus_CONFIGURING
			if prevStates[csvID] == resp.Status {
				resp.Status = ControllerStatus_ControllerStatus_UNKNOWN
				break
			}
			resp.Id = csvID
			prevStates[csvID] = resp.Status
			if err := callCb(); err != nil {
				return err
			}
			resp.Reset()
		}
	}
}

// DoExecControllerRequest executes a controller via a ExecControllerRequest.
func DoExecControllerRequest(
	ctx context.Context,
	cbus bus.Bus,
	conf config.Config,
	cb func(ControllerStatus),
) error {
	if cb == nil {
		cb = func(ControllerStatus) {}
	}
	dir := resolver.NewLoadControllerWithConfig(conf)

	cb(ControllerStatus_ControllerStatus_CONFIGURING)
	_, valRef, err := bus.ExecOneOff(ctx, cbus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	cb(ControllerStatus_ControllerStatus_RUNNING)
	<-ctx.Done()
	return nil
}
