package app

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

// initFromConfig should be called when starting or reloading the app. It makes
// all the connections between cache zones, virtual hosts, storage orchestrators
// and upstreams.
func (a *Application) initFromConfig() (err error) {
	// Make the vhost and storage orchestrator maps
	a.virtualHosts = make(map[string]*types.VirtualHost)
	a.orchestrators = make(map[string]types.StorageOrchestrator)

	// Create a global application context
	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	// Initialize the global logger
	if a.logger, err = logger.New(a.cfg.Logger); err != nil {
		return err
	}

	// Initialize all cache storage orchestrators
	for _, cfgStorage := range a.cfg.CacheZones {
		o, err := storage.NewOrchestrator(a.ctx, cfgStorage, a.logger)
		if err != nil {
			return err
		}
		a.orchestrators[cfgStorage.ID] = o
	}

	// Initialize all vhosts
	for _, cfgVhost := range a.cfg.HTTP.Servers {
		vhost := types.VirtualHost{
			Name:     cfgVhost.Name,
			CacheKey: cfgVhost.CacheKey,
			Logger:   a.logger,
		}
		a.virtualHosts[cfgVhost.Name] = &vhost

		if cfgVhost.Logger != nil {
			if vhost.Logger, err = logger.New(*cfgVhost.Logger); err != nil {
				return err
			}
		}

		if vhost.Handler, err = handler.New(cfgVhost.HandlerType); err != nil {
			return err
		}

		//!TODO: the rest of the initialization should probably be handled by
		// the handler constructor itself, like how each log type handles its
		// own specific settings, not with string comparisons of the type here

		// If this is not a proxy hanlder, there is no need to initialize the rest
		if cfgVhost.HandlerType != "proxy" {
			continue
		}

		if vhost.Upstream, err = upstream.New(cfgVhost.UpstreamType, cfgVhost.UpstreamAddress); err != nil {
			return err
		}

		if cfgVhost.CacheZone == nil {
			return fmt.Errorf("Cache zone for %s was nil", cfgVhost.Name)
		}

		orchestrator, ok := a.orchestrators[cfgVhost.CacheZone.ID]
		if !ok {
			return fmt.Errorf("Could not get the cache zone for vhost %s", cfgVhost.Name)
		}
		vhost.Orchestrator = orchestrator
	}

	a.ctx = contexts.NewStorageOrchestratorsContext(a.ctx, a.orchestrators)

	return nil
}
