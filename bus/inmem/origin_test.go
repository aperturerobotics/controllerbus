package inmem

import (
	"context"
	"slices"
	"testing"
	"time"

	cbcontroller "github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	directivecontroller "github.com/aperturerobotics/controllerbus/directive/controller"
	"github.com/sirupsen/logrus"
)

// originTestDirective is a directive equivalent to any other with the same name.
type originTestDirective struct{ name string }

func (d *originTestDirective) GetName() string { return d.name }
func (d *originTestDirective) Validate() error { return nil }
func (d *originTestDirective) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{}
}
func (d *originTestDirective) GetDebugVals() directive.DebugValues { return nil }
func (d *originTestDirective) IsEquivalent(other directive.Directive) bool {
	o, ok := other.(*originTestDirective)
	return ok && o.name == d.name
}

// originTestController is a controller with an ID and optional behavior.
type originTestController struct {
	id        string
	executeFn func(context.Context) error
	handleFn  func(context.Context, directive.Instance) ([]directive.Resolver, error)
}

func (c *originTestController) GetControllerInfo() *cbcontroller.Info {
	return cbcontroller.NewInfo(c.id, cbcontroller.MustParseVersion("0.0.1"), "origin test")
}

func (c *originTestController) Execute(ctx context.Context) error {
	if c.executeFn == nil {
		return nil
	}
	return c.executeFn(ctx)
}

func (c *originTestController) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	if c.handleFn == nil {
		return nil, nil
	}
	return c.handleFn(ctx, di)
}

func (c *originTestController) Close() error { return nil }

// holdDirective adds dir with ctx and holds the reference until ctx is done.
func holdDirective(b *Bus, dir directive.Directive) func(context.Context) error {
	return func(ctx context.Context) error {
		_, ref, err := b.AddDirective(ctx, dir, nil)
		if err != nil {
			return err
		}
		<-ctx.Done()
		ref.Release()
		return nil
	}
}

// findInstance returns the live instance for the named directive, if any.
func findInstance(b *Bus, name string) directive.Instance {
	for _, di := range b.GetDirectives() {
		if di.GetDirective().GetName() == name {
			return di
		}
	}
	return nil
}

// waitFor polls cond until it holds or the test times out.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(time.Millisecond * 5)
	}
}

// originControllerIDs returns the sorted origin controller IDs of di's references.
func originControllerIDs(di directive.Instance) []string {
	var ids []string
	for _, ref := range di.GetReferences() {
		ids = append(ids, ref.Origin.ControllerID)
	}
	slices.Sort(ids)
	return ids
}

// TestOriginGraph checks that references record their origin from the context.
// A deduplicated directive keeps one origin per reference.
func TestOriginGraph(t *testing.T) {
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	b := NewBusWithLogger(directivecontroller.NewController(ctx, le), le)

	d1 := &originTestDirective{name: "d1"}
	d2 := &originTestDirective{name: "d2"}

	ctrlB := &originTestController{id: "b"}
	ctrlB.handleFn = func(_ context.Context, di directive.Instance) ([]directive.Resolver, error) {
		if di.GetDirective().GetName() != "d1" {
			return nil, nil
		}
		return []directive.Resolver{directive.NewFuncResolver(func(rctx context.Context, _ directive.ResolverHandler) error {
			return holdDirective(b, d2)(rctx)
		})}, nil
	}
	ctrlA := &originTestController{id: "a", executeFn: holdDirective(b, d1)}
	ctrlC := &originTestController{id: "c", executeFn: holdDirective(b, d1)}

	for _, ctrl := range []*originTestController{ctrlB, ctrlA, ctrlC} {
		rel, err := b.AddController(ctx, ctrl, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer rel()
	}

	waitFor(t, "two references to d1 and one to d2", func() bool {
		di1, di2 := findInstance(b, "d1"), findInstance(b, "d2")
		return di1 != nil && di2 != nil && len(di1.GetReferences()) == 2 && len(di2.GetReferences()) == 1
	})
	di1, di2 := findInstance(b, "d1"), findInstance(b, "d2")

	if got := originControllerIDs(di1); !slices.Equal(got, []string{"a", "c"}) {
		t.Fatalf("d1 origins: got %v, want [a c]", got)
	}
	for _, ref := range di1.GetReferences() {
		if ref.Origin.Directive != nil {
			t.Fatalf("d1 added from Execute should have no origin directive, got %s", ref.Origin.Directive.GetDirectiveIdent())
		}
	}
	ref2 := di2.GetReferences()[0]
	if ref2.Origin.Directive != di1 || ref2.Origin.ControllerID != "b" {
		t.Fatalf("d2 origin: got directive %v controller %q, want d1 through b", ref2.Origin.Directive, ref2.Origin.ControllerID)
	}

	b.RemoveController(ctrlC)
	waitFor(t, "c's reference to d1 to go away", func() bool {
		return slices.Equal(originControllerIDs(di1), []string{"a"})
	})
}

// TestOriginCycle checks a resolver adding its own directive records the back edge.
// The add does not hang.
func TestOriginCycle(t *testing.T) {
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	b := NewBusWithLogger(directivecontroller.NewController(ctx, le), le)

	loop := &originTestDirective{name: "loop"}
	ctrl := &originTestController{id: "looper"}
	ctrl.handleFn = func(context.Context, directive.Instance) ([]directive.Resolver, error) {
		return []directive.Resolver{directive.NewFuncResolver(func(rctx context.Context, _ directive.ResolverHandler) error {
			return holdDirective(b, loop)(rctx)
		})}, nil
	}
	rel, err := b.AddController(ctx, ctrl, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer rel()

	di, ref, err := b.AddDirective(ctx, loop, nil)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the back edge", func() bool {
		for _, r := range di.GetReferences() {
			if r.Origin.Directive == di && r.Origin.ControllerID == "looper" {
				return true
			}
		}
		return false
	})
	ref.Release()
	di.Close()
}

// TestOriginBroadcast checks that adding and releasing a reference to an
// existing directive signals the directives broadcast.
func TestOriginBroadcast(t *testing.T) {
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	b := NewBusWithLogger(directivecontroller.NewController(ctx, le), le)

	d := &originTestDirective{name: "shared"}
	_, first, err := b.AddDirective(ctx, d, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()

	waitCh := func() <-chan struct{} {
		var ch <-chan struct{}
		b.GetDirectivesBroadcast().HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			ch = getWaitCh()
		})
		return ch
	}
	expectSignal := func(what string, ch <-chan struct{}) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			t.Fatalf("no broadcast after %s", what)
		}
	}

	ch := waitCh()
	_, second, err := b.AddDirective(ctx, d, nil)
	if err != nil {
		t.Fatal(err)
	}
	expectSignal("adding a deduplicated reference", ch)

	ch = waitCh()
	second.Release()
	expectSignal("releasing a reference", ch)
}
