package hot_loader_filesystem

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	hot_loader "github.com/aperturerobotics/controllerbus/hot/loader"
	hot_plugin "github.com/aperturerobotics/controllerbus/hot/plugin"
	"github.com/blang/semver"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var debounceTime = time.Second

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "controllerbus/hot/loader/filesystem/1"

// Controller is the hot plugin filesystem loading controller.
type Controller struct {
	// le is the root logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// dir is the directory to watch
	dir string
	// watch indicates to watch the filesystem
	watch bool
}

// NewController constructs a new controller.
func NewController(le *logrus.Entry, bus bus.Bus, conf *Config) (*Controller, error) {
	dir := path.Clean(conf.GetDir())
	if _, err := os.Stat(dir); err != nil {
		return nil, errors.Wrapf(err, "stat %s", dir)
	}
	return &Controller{
		le:  le,
		bus: bus,
		dir: dir,

		watch: conf.GetWatch(),
	}, nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// Load all plugins (.cbus.so)
	pluginSuffix := hot_plugin.PluginSuffix
	loadedPlugins := make(map[string]*hot_loader.LoadedPlugin)
	unloadPlugin := func(id string) {
		plug, plugOk := loadedPlugins[id]
		if !plugOk {
			return
		}
		c.le.
			WithField("plugin-path", id).
			Info("unloading plugin")
		plug.Close()
		delete(loadedPlugins, id)
	}
	loadPlugin := func(plugPath string) error {
		_, plugOk := loadedPlugins[plugPath]
		if plugOk {
			return nil
		}
		c.le.
			WithField("plugin-path", plugPath).
			Info("loading plugin")
		plugActPath := path.Join(c.dir, plugPath)
		lp, err := hot_loader.LoadPluginSharedLibrary(
			ctx,
			c.le.WithField("plugin-id", plugPath),
			c.bus,
			plugActPath,
		)
		if err != nil {
			return err
		}
		loadedPlugins[plugPath] = lp
		c.le.
			WithField("plugin-id", plugPath).
			WithField("plugin-path", plugActPath).
			WithField("plugin-binary-id", lp.GetBinaryID()).
			WithField("plugin-binary-version", lp.GetBinaryVersion()).
			Info("successfully loaded plugin")
		return nil
	}
	syncPlugins := func() error {
		scanDir := c.dir
		dirContents, err := ioutil.ReadDir(scanDir)
		if err != nil {
			return err
		}
		foundNames := make(map[string]struct{})
		for _, df := range dirContents {
			if df.IsDir() || !df.Mode().IsRegular() {
				continue
			}
			dfName := df.Name()
			if !strings.HasSuffix(dfName, pluginSuffix) {
				continue
			}
			foundNames[dfName] = struct{}{}
		}
		for loadedID := range loadedPlugins {
			if _, ok := foundNames[loadedID]; !ok {
				unloadPlugin(loadedID)
			}
		}
		for plugFile := range foundNames {
			if _, ok := loadedPlugins[plugFile]; !ok {
				if err := loadPlugin(plugFile); err != nil {
					c.le.
						WithError(err).
						Warn("unable to load plugin file")
				}
			}
		}
		return nil
	}

	defer func() {
		for id, pg := range loadedPlugins {
			pg.Close()
			delete(loadedPlugins, id)
		}
	}()

	if err := syncPlugins(); err != nil {
		return err
	}

	if !c.watch {
		<-ctx.Done()
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(c.dir); err != nil {
		return err
	}

	var nextSyncTicker *time.Timer
	var nextSyncC <-chan time.Time
	defer func() {
		if nextSyncTicker != nil {
			nextSyncTicker.Stop()
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			switch event.Op {
			case fsnotify.Create:
			case fsnotify.Rename:
			case fsnotify.Write:
			case fsnotify.Remove:
			default:
				continue
			}
			if nextSyncTicker != nil {
				nextSyncTicker.Stop()
			}
			nextSyncTicker = time.NewTimer(debounceTime)
			nextSyncC = nextSyncTicker.C
		case err, ok := <-watcher.Errors:
			if !ok || err == context.Canceled {
				return nil
			}
			return errors.Wrap(err, "watcher error")
		case <-nextSyncC:
			nextSyncTicker = nil
			nextSyncC = nil
			c.le.Debug("re-syncing plugins list")
			if err := syncPlugins(); err != nil {
				return err
			}
		}
	}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(
	ctx context.Context,
	di directive.Instance,
) (directive.Resolver, error) {
	return nil, nil
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		ControllerID,
		Version,
		"hot plugin filesystem loader: "+c.dir,
	)
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	return nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
