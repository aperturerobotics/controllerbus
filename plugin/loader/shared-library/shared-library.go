package plugin_shared_library

import (
	"context"
	"errors"
	"plugin"
	"reflect"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	cbus_plugin "github.com/aperturerobotics/controllerbus/plugin"
	"github.com/sirupsen/logrus"
)

// PluginSuffix is the plugin file suffix.
const PluginSuffix = ".cbus.so"

// LoadedPlugin contains a loaded Go plugin.
type LoadedPlugin struct {
	cbus_plugin.Plugin
	cbus_plugin.PluginResolver
	PluginStat

	closeOnce sync.Once
	ctx       context.Context
	ctxCancel context.CancelFunc
}

// LoadPluginSharedLibrary loads a plugin from a file.
func LoadPluginSharedLibrary(
	ctx context.Context,
	le *logrus.Entry,
	bus bus.Bus,
	pluginPath string,
) (*LoadedPlugin, error) {
	// Get info about the file.
	pluginSt, err := NewPluginStat(pluginPath)
	if err != nil {
		return nil, err
	}

	// Load the plugin.
	pg, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	sym, err := pg.Lookup(cbus_plugin.PluginGlobalVar)
	if err != nil {
		return nil, err
	}
	var loadedPlugin cbus_plugin.Plugin
	loadedPluginPtr, loadedPluginOk := sym.(*cbus_plugin.Plugin)
	if !loadedPluginOk {
		// It's possible the interface type is different:
		// Dereference first with introspection.
		symVal := reflect.ValueOf(sym)
		if symVal.Kind() != reflect.Ptr {
			return nil, errInvalidPluginType(sym)
		}

		sym = symVal.Elem().Interface()
		loadedPlugin, loadedPluginOk = sym.(cbus_plugin.Plugin)
		if !loadedPluginOk {
			return nil, errInvalidPluginType(sym)
		}
	} else {
		loadedPlugin = *loadedPluginPtr
	}
	if loadedPlugin == nil {
		return nil, errors.New("could not load hot plugin, nil plugin global")
	}
	pluginResolver, err := loadedPlugin.NewPluginResolver(ctx, bus)
	if err != nil {
		return nil, err
	}
	subCtx, subCtxCancel := context.WithCancel(ctx)
	go func() {
		resolverController := resolver.NewController(le, bus, pluginResolver)
		err := bus.ExecuteController(
			subCtx,
			resolverController,
		)
		if err != nil {
			// typically not possible
			if err != context.Canceled {
				le.WithError(err).Warn("hot loader controller exited with error")
			}
			return
		}
		<-subCtx.Done()
		_ = bus.RemoveController(ctx, resolverController)
		resolverController.Close()
	}()
	return &LoadedPlugin{
		ctx:            subCtx,
		ctxCancel:      subCtxCancel,
		Plugin:         loadedPlugin,
		PluginResolver: pluginResolver,
		PluginStat:     *pluginSt,
	}, nil
}

// Close closes the loaded plugin.
func (l *LoadedPlugin) Close() {
	l.closeOnce.Do(func() {
		l.ctxCancel()
		if l.PluginResolver != nil {
			l.PluginResolver.PrePluginUnload()
		}
		if l.Plugin != nil {
			l.Plugin.PrePluginUnload()
		}
		// TODO unload shared go plugin (not possible)
	})
}
