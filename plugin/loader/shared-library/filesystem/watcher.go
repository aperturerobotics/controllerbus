//go:build !js && !wasm

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
	"github.com/aperturerobotics/fsnotify"
	debounce_fswatcher "github.com/aperturerobotics/util/debounce-fswatcher"
	"github.com/sirupsen/logrus"
)

// debounceTime coalesces filesystem events before reconciling plugins.
const debounceTime = time.Second

// Watcher watches a filesystem path to load and reload plugins.
type Watcher struct {
	// le is the logger.
	le *logrus.Entry
	// bus is the controller bus.
	bus bus.Bus
	// mtx guards loadedPlugins.
	mtx sync.Mutex
	// loadedPlugins holds the plugins attached by this watcher.
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
	// Remove the tracked plugin before running its unload hooks.
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

	// Unload hooks run without the watcher lock.
	plug.Close()
}

// LoadPlugin loads a plugin from a path.
//
// A path already loaded is left unchanged.
func (w *Watcher) LoadPlugin(ctx context.Context, plugPath string) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	return w.loadPluginLocked(ctx, plugPath)
}

// loadPluginLocked loads plugin if mtx is locked by caller.
func (w *Watcher) loadPluginLocked(ctx context.Context, plugPath string) error {
	// Keep the existing resolver when this path is already loaded.
	_, plugOk := w.loadedPlugins[plugPath]
	if plugOk {
		return nil
	}

	// Attach the shared library and retain it for later reconciliation.
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

	// Report the binary identity supplied by the loaded plugin.
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
	// Serialize discovery and reconciliation against manual loads and unloads.
	w.mtx.Lock()
	defer w.mtx.Unlock()

	// Discover the shared libraries and snapshot their file metadata.
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

	// Unload removed or changed binaries before loading replacements.
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

		// Compare through the loaded plugin's existing stats API.
		if loadedInfo.Equal(foundPlugin) {
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

	// Attach discovered paths that are no longer in the loaded set.
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
	// Keep every loaded resolver within this watcher run.
	defer w.UnloadAll()
	if err := w.SyncPlugins(ctx, syncDir); err != nil {
		return err
	}

	// A one-shot scan keeps its plugins until cancellation.
	if !watch {
		<-ctx.Done()
		return nil
	}

	// Subscribe to directory changes for subsequent reconciliation.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	if err := watcher.Add(syncDir); err != nil {
		return err
	}

	// Reconcile each coalesced batch of filesystem changes.
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
