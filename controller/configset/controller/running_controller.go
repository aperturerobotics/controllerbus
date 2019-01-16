package configset_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
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
}

func newRunningController(
	c *Controller,
	key string,
	conf configset.ControllerConfig,
) *runningController {
	return &runningController{
		c:             c,
		le:            c.le.WithField("config-key", key),
		key:           key,
		conf:          conf,
		confRestartCh: make(chan struct{}, 1),
	}
}

// GetConfigKey returns the configset key for this controller.
func (c *runningController) GetConfigKey() string {
	return c.key
}

// Execute actuates the running controller.
func (c *runningController) Execute(ctx context.Context) (rerr error) {
	defer func() {
		if rerr != nil {
			if rerr != context.Canceled {
				c.le.WithError(rerr).Warn("controller exited with error")
			}
			c.mtx.Lock()
			c.state.ctrl = nil
			c.state.err = rerr
			s := c.state
			c.pushState(&s)
			c.mtx.Unlock()
		}
	}()

	for {
		c.mtx.Lock()
		conf := c.conf
		c.mtx.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// load config constructor by id
		/*
			c.le.Debug("loading config constructor")
			configCtorDir := resolver.NewLoadConfigConstructorByID(conf.GetConfigId())
			configCtorVal, configCtorRef, err := bus.ExecOneOff(ctx, c.c.bus, configCtorDir, nil)
			if err != nil {
				return errors.WithMessage(err, "resolve config object")
			}

			confObj := configCtorVal.(config.Constructor).ConstructConfig()
			if err := proto.Unmarshal(conf.GetData(), confObj); err != nil {
				configCtorRef.Release()
				return errors.WithMessage(err, "unmarshal config")
			}
		*/

		// execute the controller with the current config
		valCtx, valCtxCancel := context.WithCancel(ctx)
		execDir := resolver.NewLoadControllerWithConfig(conf.GetConfig())
		c.le.Debug("executing controller")
		ev, execRef, err := bus.ExecOneOff(valCtx, c.c.bus, execDir, valCtxCancel)
		// configCtorRef.Release()
		if err != nil {
			valCtxCancel()
			if err == context.Canceled {
				return err
			}

			return errors.WithMessage(err, "exec controller")
		}
		c.mtx.Lock()
		c.state.ctrl, _ = ev.GetValue().(controller.Controller)
		s := c.state
		c.pushState(&s)
		c.mtx.Unlock()
		select {
		case <-ctx.Done():
			execRef.Release()
			return ctx.Err()
		case <-valCtx.Done():
			execRef.Release()
			continue
		case <-c.confRestartCh:
			valCtxCancel()
			execRef.Release()
			c.le.Debug("restarting with new config")
			c.mtx.Lock()
			c.state.ctrl = nil
			s := c.state
			c.pushState(&s)
			c.mtx.Unlock()
		}
	}
}

// GetControllerConfig returns the controller config in use.
// The value will be revoked and re-emitted if this changes.
func (c *runningController) GetControllerConfig() configset.ControllerConfig {
	return c.conf
}

// GetState returns the current state object.
func (c *runningController) GetState() configset.State {
	return nil
}

// AddReference adds a reference to the running controller.
func (c *runningController) AddReference(ref *runningControllerRef) {
	c.mtx.Lock()
	c.refs = append(c.refs, ref)
	c.mtx.Unlock()
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

	if c.conf.GetRevision() >= conf.GetRevision() {
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
func (c *runningController) pushState(st configset.State) {
	for _, ref := range c.refs {
		ref.pushState(st)
	}
}

// _ is a type assertion
// var _ configset.ApplyConfigSetValue = ((*runningController)(nil))
