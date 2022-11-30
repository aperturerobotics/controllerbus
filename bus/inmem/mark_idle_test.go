package inmem

import (
	"context"
	"errors"
	"testing"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/controllerbus/directive/controller"
	boilerplate_controller "github.com/aperturerobotics/controllerbus/example/boilerplate/controller"
	boilerplate_v1 "github.com/aperturerobotics/controllerbus/example/boilerplate/v1"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// markIdleCtrl is the controller for MarkIdle
type markIdleCtrl struct {
	*bus.BusController[*boilerplate_controller.Config]
}

// HandleDirective asks if the handler can resolve the directive.
func (c *markIdleCtrl) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	switch di.GetDirective().(type) {
	case *boilerplate_v1.Boilerplate:
		return directive.Resolvers(&markIdleRes{}), nil
	}
	return nil, nil
}

// markIdleRes is a resolver that marks itself as idle then waits for ctx to be canceled.
type markIdleRes struct{}

func (m *markIdleRes) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	handler.MarkIdle()
	<-ctx.Done()
	return ctx.Err()
}

// _ is a type assertion
var _ directive.Resolver = ((*markIdleRes)(nil))

func TestMarkIdle(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	b := NewBus(controller.NewController(ctx, le))

	// add test controller
	boilerplateConf := &boilerplate_controller.Config{}
	testMarkIdle := bus.NewBusController(le, b, boilerplateConf, "test-mark-idle", semver.MustParse("0.0.1"), "")
	markIdleCtrl := &markIdleCtrl{BusController: testMarkIdle}
	rel, err := b.AddController(ctx, markIdleCtrl, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer rel()

	res, resRef, err := bus.ExecOneOff(ctx, b, &boilerplate_v1.Boilerplate{
		MessageText: "hello world",
	}, true, nil)
	if res != nil || resRef != nil {
		err = errors.New("expected idle with no values")
	}
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("successfully marked as idle")
}
