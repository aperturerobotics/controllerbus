package plugin_loader_filesystem

import (
	"context"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aperturerobotics/controllerbus/bus"
	shared "github.com/aperturerobotics/controllerbus/plugin/loader/shared-library"
	debounce_fswatcher "github.com/aperturerobotics/util/debounce-fswatcher"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// debounceTime debounces re-sync requests
var debounceTime = time.Second

// Watcher watches a filesystem path to load and reload plugins.
type Watcher struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// mtx guards below fields
	mtx sync.Mutex
	// loadedPlugins is the set of loaded plugins
	loadedPlugins map[string]*shared.LoadedPlugin
}

// NewWatcher builds a new filesystem watcher.
func NewWatcher(le *logrus.Entry, bus bus.Bus) *Watcher {
	return &Watcher{
		le:            le,
		bus:           bus,
		loadedPlugins: make(map[string]*shared.LoadedPlugin),
	}
}

// UnloadPlugin unloads a loaded plugin by ID.
func (w *Watcher) UnloadPlugin(id string) {
	w.mtx.Lock()
	plug, plugOk := w.loadedPlugins[id]
	if !plugOk {
		w.mtx.Unlock()
		return
	}
	w.le.
		WithField("plugin-path", id).
		Info("unloading plugin")
	delete(w.loadedPlugins, id)
	w.mtx.Unlock()
	plug.Close()
}

// LoadPlugin loads a plugin from a path.
//
// If the path was already loaded, no-op.
func (w *Watcher) LoadPlugin(ctx context.Context, plugPath string) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	return w.loadPluginLocked(ctx, plugPath)
}

// loadPluginLocked loads plugin if mtx is locked by caller.
func (w *Watcher) loadPluginLocked(ctx context.Context, plugPath string) error {
	_, plugOk := w.loadedPlugins[plugPath]
	if plugOk {
		return nil
	}

	w.le.
		WithField("plugin-path", plugPath).
		Info("loading plugin")
	lp, err := shared.LoadPluginSharedLibrary(
		ctx,
		w.le.WithField("plugin-id", plugPath),
		w.bus,
		plugPath,
	)
	if err != nil {
		return err
	}
	w.loadedPlugins[plugPath] = lp
	w.le.
		WithField("plugin-path", plugPath).
		WithField("plugin-binary-id", lp.GetBinaryID()).
		WithField("plugin-binary-version", lp.GetBinaryVersion()).
		Info("successfully loaded plugin")
	return nil
}

// SyncPlugins synchronizes the loaded plugins with all in a scan dir.
//
// If any plugin files were removed, unloads those plugins.
func (w *Watcher) SyncPlugins(ctx context.Context, scanDir string) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	dirContents, err := os.ReadDir(scanDir)
	if err != nil {
		return err
	}
	foundNames := make(map[string]*shared.PluginStat)
	for _, df := range dirContents {
		if df.IsDir() || !df.Type().IsRegular() {
			continue
		}
		dfName := df.Name()
		if !strings.HasSuffix(dfName, shared.PluginSuffix) {
			continue
		}
		plugPath := path.Join(scanDir, dfName)
		dfStat, err := shared.NewPluginStat(plugPath)
		if err != nil {
			w.le.
				WithError(err).
				WithField("plugin-path", plugPath).
				Warn("cannot stat plugin path")
			continue
		}
		foundNames[plugPath] = dfStat
	}
	for loadedID, loadedInfo := range w.loadedPlugins {
		foundPlugin, ok := foundNames[loadedID]
		if !ok {
			// plugin no longer exists.
			loadedInfo.Close()
			delete(w.loadedPlugins, loadedID)
			w.le.Debugf(
				"%s: unloading removed plugin, removed(%d){%s}",
				loadedID,
				loadedInfo.GetBinarySize(),
				loadedInfo.GetModificationTime().String(),
			)
			continue
		}

		plugsEqual := foundPlugin.Equal(&loadedInfo.PluginStat)
		if plugsEqual {
			continue
		}
		w.le.Debugf(
			"%s: reloading plugin, discovered(%d){%s} != loaded(%d){%s}",
			loadedID,
			foundPlugin.GetBinarySize(),
			foundPlugin.GetModificationTime().String(),
			loadedInfo.GetBinarySize(),
			loadedInfo.GetModificationTime().String(),
		)
		loadedInfo.Close()
		delete(w.loadedPlugins, loadedID)
	}
	for plugFile := range foundNames {
		if _, ok := w.loadedPlugins[plugFile]; !ok {
			if err := w.loadPluginLocked(ctx, plugFile); err != nil {
				w.le.
					WithError(err).
					Warn("unable to load plugin file")
			}
		}
	}
	return nil
}

// UnloadAll unloads all plugins.
func (w *Watcher) UnloadAll() {
	w.mtx.Lock()
	for id, pg := range w.loadedPlugins {
		pg.Close()
		delete(w.loadedPlugins, id)
	}
	w.mtx.Unlock()
}

// Execute executes the watcher and loads / runs plugins.
func (w *Watcher) Execute(ctx context.Context, syncDir string, watch bool) error {
	defer w.UnloadAll()
	if err := w.SyncPlugins(ctx, syncDir); err != nil {
		return err
	}

	if !watch {
		<-ctx.Done()
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(syncDir); err != nil {
		return err
	}

	for {
		happened, err := debounce_fswatcher.DebounceFSWatcherEvents(ctx, watcher, debounceTime, nil)
		if err != nil {
			return err
		}

		w.le.Debugf("re-syncing plugins after %d filesystem events", len(happened))
		if err := w.SyncPlugins(ctx, syncDir); err != nil {
			return err
		}
	}
}
