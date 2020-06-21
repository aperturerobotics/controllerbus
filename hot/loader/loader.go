package hot_loader

import (
	"context"
	"errors"
	"plugin"
	"sync"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	hot_plugin "github.com/aperturerobotics/controllerbus/hot/plugin"
	"github.com/sirupsen/logrus"
)

// HotPluginGlobalVar is the global var to load.
const HotPluginGlobalVar = "ControllerBusHotPlugin"

// LoadedPlugin contains a loaded Go plugin.
type LoadedPlugin struct {
	hot_plugin.HotPlugin
	hot_plugin.HotResolver
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
	sym, err := pg.Lookup(HotPluginGlobalVar)
	if err != nil {
		return nil, err
	}
	hotPluginPtr, hotPluginOk := sym.(*hot_plugin.HotPlugin)
	if !hotPluginOk {
		return nil, errors.New("could not load hot plugin, invalid type")
	}
	hotPlugin := *hotPluginPtr
	if hotPlugin == nil {
		return nil, errors.New("could not load hot plugin, nil plugin global")
	}
	hotResolver, err := hotPlugin.NewHotResolver(ctx, bus)
	if err != nil {
		return nil, err
	}
	subCtx, subCtxCancel := context.WithCancel(ctx)
	go func() {
		resolverController := resolver.NewController(le, bus, hotResolver)
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
		ctx:         subCtx,
		ctxCancel:   subCtxCancel,
		HotPlugin:   hotPlugin,
		HotResolver: hotResolver,
		PluginStat:  *pluginSt,
	}, nil
}

// Close closes the loaded plugin.
func (l *LoadedPlugin) Close() {
	l.closeOnce.Do(func() {
		l.ctxCancel()
		if l.HotResolver != nil {
			l.HotResolver.PrePluginUnload()
		}
		if l.HotPlugin != nil {
			l.HotPlugin.PrePluginUnload()
		}
		// TODO unload go plugin (not possible)
	})
}
