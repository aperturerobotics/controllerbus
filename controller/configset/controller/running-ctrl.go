package configset_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// runningController contains information about a running controller.
// it is also the value for the directive ApplyConfigSet
type runningController struct {
	c *Controller
	// le is the logger
	le *logrus.Entry
	// key is the controller key
	key string
	// conf is the controller config
	conf configset.ControllerConfig
	// confRestartCh indicates the controller should be restarted
	confRestartCh chan struct{}

	// mtx guards the state
	mtx sync.Mutex
	// state is the current state
	state runningControllerState
	// refs contains references
	refs []*runningControllerRef
	// prevEmittedState is the previously emitted state
	prevEmittedState runningControllerState
}

func newRunningController(
	c *Controller,
	key string,
	conf configset.ControllerConfig,
) *runningController {
	rc := &runningController{
		c:             c,
		le:            c.le.WithField("config-key", key),
		key:           key,
		conf:          conf,
		confRestartCh: make(chan struct{}, 1),
	}
	rc.state.id = key
	return rc
}

// GetConfigKey returns the configset key for this controller.
func (c *runningController) GetConfigKey() string {
	return c.key
}

// Execute actuates the running controller.
func (c *runningController) Execute(ctx context.Context) (rerr error) {
	c.mtx.Lock()
	conf := c.conf
	c.mtx.Unlock()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ctrlConf := conf.GetConfig()

		// execute the controller with the current config
		execDir := resolver.NewLoadControllerWithConfig(ctrlConf)
		c.le.Debug("executing controller")
		updValCh, updValDir, execRef, err := bus.ExecOneOffWatchCh[resolver.LoadControllerWithConfigValue](c.c.bus, execDir)
		if err != nil {
			if err == context.Canceled {
				return err
			}

			err = errors.Wrap(err, "exec controller")
		}

		disposed := make(chan struct{})
		disposeCb := updValDir.AddDisposeCallback(func() {
			close(disposed)
		})

		clearState := func() {
			c.mtx.Lock()
			c.state.ctrl = nil
			c.state.err = err
			c.state.conf = conf
			s := c.state
			c.pushState(&s)
			c.mtx.Unlock()
		}
		clearState()

	RecheckStateLoop:
		for {
			select {
			case <-ctx.Done():
				break RecheckStateLoop
			case <-c.confRestartCh:
				c.le.Info("restarting with new config")
				break RecheckStateLoop
			case <-disposed:
				break RecheckStateLoop
			case aval := <-updValCh:
				if aval == nil {
					clearState()
					continue RecheckStateLoop
				}
				uval := aval.GetValue()
				c.mtx.Lock()
				uerr := uval.GetError()
				if uerr != nil && uerr != c.state.err {
					c.le.WithError(uerr).Warn("controller error")
				}
				c.state.err = uerr
				c.state.ctrl = uval.GetController()
				c.state.conf = conf
				st := c.state
				c.pushState(&st)
				c.mtx.Unlock()
			}
		}
		execRef.Release()
		disposeCb()
		c.mtx.Lock()
		conf = c.conf
		if c.state.ctrl != nil || c.state.conf != conf {
			c.state.ctrl = nil
			c.state.conf = conf
			s := c.state
			c.pushState(&s)
		}
		c.mtx.Unlock()
	}
}

// GetControllerConfig returns the controller config in use.
// The value will be revoked and re-emitted if this changes.
func (c *runningController) GetControllerConfig() configset.ControllerConfig {
	return c.conf
}

// GetState returns the current state object.
func (c *runningController) GetState() configset.State {
	c.mtx.Lock()
	st := c.state
	c.mtx.Unlock()
	return &st
}

// AddReference adds a reference to the running controller.
func (c *runningController) AddReference(ref *runningControllerRef) {
	c.mtx.Lock()
	c.refs = append(c.refs, ref)
	c.mtx.Unlock()
}

// ApplyReference applies an existing reference to the running controller.
func (c *runningController) ApplyReference(ref *runningControllerRef) {
	if ref.GetRunningController() == c {
		return
	}
	c.mtx.Lock()
	st := c.state
	c.mtx.Unlock()
	if ref.setRunningController(c, &st) {
		c.mtx.Lock()
		c.refs = append(c.refs, ref)
		c.mtx.Unlock()
	}
}

// DelReference removes a reference from the running controller.
func (c *runningController) DelReference(ref *runningControllerRef) {
	c.mtx.Lock()
	for i, r := range c.refs {
		if r == ref {
			c.refs[i] = c.refs[len(c.refs)-1]
			c.refs[len(c.refs)-1] = nil
			c.refs = c.refs[:len(c.refs)-1]
			break
		}
	}
	rc := len(c.refs)
	c.mtx.Unlock()
	if rc == 0 {
		c.c.wake()
	}
}

// ApplyConfig applies a configuration if the revision is newer
func (c *runningController) ApplyConfig(conf configset.ControllerConfig) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.conf.GetRev() >= conf.GetRev() {
		return false
	}

	c.conf = conf
	select {
	case c.confRestartCh <- struct{}{}:
	default:
	}
	return true
}

// pushState pushes the new state to all references
// mtx should be locked by caller
func (c *runningController) pushState(st *runningControllerState) {
	// push only if changed
	if c.prevEmittedState.Equals(st) {
		return
	}
	c.prevEmittedState = *st
	for _, ref := range c.refs {
		ref.pushState(st)
	}
}

// _ is a type assertion
// var _ configset.ApplyConfigSetValue = ((*runningController)(nil))
