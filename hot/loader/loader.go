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

// LoadedPlugin contains a loaded Go plugin.
type LoadedPlugin struct {
	hot_plugin.HotPlugin
	hot_plugin.HotResolver

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
	// Load the plugin.
	pg, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	sym, err := pg.Lookup("ControllerBusHotPlugin")
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
	go bus.ExecuteController(
		subCtx,
		resolver.NewController(le, bus, hotResolver),
	)
	return &LoadedPlugin{
		ctx:         subCtx,
		ctxCancel:   subCtxCancel,
		HotPlugin:   hotPlugin,
		HotResolver: hotResolver,
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
