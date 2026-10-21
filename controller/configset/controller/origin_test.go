package configset_controller_test

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// originChildDirective is the directive the loaded controller adds.
type originChildDirective struct{}

func (d *originChildDirective) GetName() string { return "OriginChild" }
func (d *originChildDirective) Validate() error { return nil }
func (d *originChildDirective) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}
func (d *originChildDirective) GetDebugVals() directive.DebugValues { return nil }

// originController adds originChildDirective from Execute.
type originController struct {
	b bus.Bus
}

// originControllerID is the controller ID of originController.
const originControllerID = "configset/test/origin"

func (c *originController) GetControllerInfo() *controller.Info {
	return controller.NewInfo(originControllerID, controller.MustParseVersion("0.0.1"), "origin test controller")
}

func (c *originController) HandleDirective(context.Context, directive.Instance) ([]directive.Resolver, error) {
	return nil, nil
}

func (c *originController) Execute(ctx context.Context) error {
	_, ref, err := c.b.AddDirective(ctx, &originChildDirective{}, nil)
	if err != nil {
		return err
	}
	<-ctx.Done()
	ref.Release()
	return nil
}

func (c *originController) Close() error { return nil }

// originFactory constructs originController for config.Placeholder.
type originFactory struct {
	b bus.Bus
}

func (f *originFactory) GetConfigID() string            { return (&config.Placeholder{}).GetConfigID() }
func (f *originFactory) ConstructConfig() config.Config { return &config.Placeholder{} }
func (f *originFactory) GetVersion() controller.Version { return controller.MustParseVersion("0.0.1") }
func (f *originFactory) Construct(context.Context, config.Config, controller.ConstructOpts) (controller.Controller, error) {
	return &originController{b: f.b}, nil
}

// referencedFrom returns the live instances with a reference from parent.
func referencedFrom(b bus.Bus, parent directive.Instance) []directive.Instance {
	var out []directive.Instance
	for _, di := range b.GetDirectives() {
		for _, ref := range di.GetReferences() {
			if ref.Origin.Directive == parent {
				out = append(out, di)
				break
			}
		}
	}
	return out
}

// TestConfigKeyOrigin checks a ConfigSet entry's key marks the reference it adds.
// The loaded controller's directives are reachable from that reference.
func TestConfigKeyOrigin(t *testing.T) {
	le := logrus.NewEntry(logrus.New())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		t.Fatal(err)
	}
	sr.AddFactory(&originFactory{b: b})

	_, _, csCtrlRef, err := bus.ExecOneOff(ctx, b, resolver.NewLoadControllerWithConfig(&configset_controller.Config{}), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer csCtrlRef.Release()

	const key = "fg/node1/port"
	cs := configset.ConfigSet{key: configset.NewControllerConfig(1, &config.Placeholder{})}
	_, applyRef, err := b.AddDirective(ctx, configset.NewApplyConfigSet(cs), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer applyRef.Release()

	// Walk from the reference carrying the key down to the child directive.
	walk := func() (string, bool) {
		var root directive.Instance
		for _, di := range b.GetDirectives() {
			for _, ref := range di.GetReferences() {
				if ref.Origin.ConfigKey == key {
					root = di
				}
			}
		}
		if root == nil {
			return "no reference carries the config key", false
		}
		if _, ok := root.GetDirective().(resolver.LoadControllerWithConfig); !ok {
			return "the keyed reference is on " + root.GetDirectiveIdent(), false
		}
		seen := map[directive.Instance]bool{root: true}
		queue := []directive.Instance{root}
		for len(queue) != 0 {
			di := queue[0]
			queue = queue[1:]
			if _, ok := di.GetDirective().(*originChildDirective); ok {
				for _, ref := range di.GetReferences() {
					if ref.Origin.ControllerID == originControllerID {
						return "", true
					}
				}
				return "the child is not referenced from the loaded controller", false
			}
			for _, child := range referencedFrom(b, di) {
				if !seen[child] {
					seen[child] = true
					queue = append(queue, child)
				}
			}
		}
		return "the child is not reachable from the keyed reference", false
	}

	var msg string
	for {
		var ok bool
		if msg, ok = walk(); ok {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal(msg)
		case <-time.After(10 * time.Millisecond):
		}
	}
}
