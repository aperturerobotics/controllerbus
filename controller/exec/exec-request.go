package controller_exec

import (
	"context"
	"errors"
	"sort"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
	configset_proto "github.com/aperturerobotics/controllerbus/controller/configset/proto"
	"github.com/aperturerobotics/controllerbus/directive"
)

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

	var confSet configset.ConfigSet

	rConfSet := r.GetConfigSet() // proto.Clone(r.GetConfigSet()).(*configset_proto.ConfigSet)
	if rConfSet == nil && r.GetConfigSetYaml() == "" {
		return errors.New("at least one config must be specified")
	}
	confsList := rConfSet.GetConfigurations()
	prevStates := make(map[string]ControllerStatus, len(confsList))
	if !allowPartialSuccess && len(rConfSet.GetConfigurations()) != 0 {
		confSet, err = rConfSet.Resolve(ctx, cbus)
	}
	if err != nil {
		return err
	}
	if confSet == nil {
		confSet = make(configset.ConfigSet, len(confsList))
	}
	if confsList == nil {
		confsList = make(map[string]*configset_proto.ControllerConfig)
	}

	confsYAML := r.GetConfigSetYaml()
	if confsYAML != "" {
		addedConfs, err := configset_json.UnmarshalYAML(
			ctx,
			cbus,
			[]byte(confsYAML),
			confSet,
			r.GetConfigSetYamlOverwrite(),
		)
		if err != nil {
			return err
		}
		sort.Strings(addedConfs)
	}

	niniterr := 0
	var lastniniterr error
	// TODO: sort and send sorted
	for csID, conf := range confsList {
		if csID == "" {
			continue
		}

		if _, ok := confSet[csID]; !ok {
			confSet[csID], err = conf.Resolve(ctx, cbus)
			if err != nil {
				resp.Id = csID
				resp.Status = ControllerStatus_ControllerStatus_ERROR
				resp.ErrorInfo = err.Error()
				prevStates[csID] = resp.Status
				if err := callCb(); err != nil {
					return err
				}
				resp.Reset()
				niniterr++
				lastniniterr = err
				delete(confSet, csID)
			}
		}
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

	addedCh := make(chan configset.ApplyConfigSetValue, 1)
	removedCh := make(chan configset.ApplyConfigSetValue, 1)

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
