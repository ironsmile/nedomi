package app

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
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
	a.virtualHosts = make(map[string]*VirtualHost)
	a.orchestrators = make(map[string]types.StorageOrchestrator)

	// Create a global application context
	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	// Initialize the global logger
	if a.logger, err = logger.New(&a.cfg.Logger); err != nil {
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
		vhost := VirtualHost{
			Location: types.Location{
				Name:     cfgVhost.Name,
				CacheKey: cfgVhost.CacheKey,
			},
		}
		a.virtualHosts[cfgVhost.Name] = &vhost

		if vhost.Logger, err = logger.New(cfgVhost.Logger); err != nil {
			return err
		}

		if cfgVhost.UpstreamType != "" || cfgVhost.UpstreamAddress != nil {
			if vhost.Upstream, err = upstream.New(cfgVhost.UpstreamType, cfgVhost.UpstreamAddress); err != nil {
				return err
			}
		}

		if cfgVhost.CacheZone != nil {
			orchestrator, ok := a.orchestrators[cfgVhost.CacheZone.ID]
			if !ok {
				return fmt.Errorf("Could not get the cache zone for vhost %s", cfgVhost.Name)
			}
			vhost.Orchestrator = orchestrator
		}

		if vhost.Handler, err = handler.New(cfgVhost.HandlerType.Type, cfgVhost.HandlerType.Settings, &vhost.Location); err != nil {
			return err
		}
		var locations []*types.Location
		if locations, err = a.initFromConfigLocationsForVHost(cfgVhost.Locations); err != nil {
			return err
		}

		if vhost.Muxer, err = NewLocationMuxer(locations); err != nil {
			return fmt.Errorf("Could not create location muxer for vhost %s", cfgVhost.Name)
		}
	}

	a.ctx = contexts.NewStorageOrchestratorsContext(a.ctx, a.orchestrators)

	return nil
}

func (a *Application) initFromConfigLocationsForVHost(cfgLocations []*config.Location) ([]*types.Location, error) {
	var err error
	var locations = make([]*types.Location, len(cfgLocations))
	for index, locCfg := range cfgLocations {
		locations[index] = &types.Location{
			Name:     locCfg.Name,
			CacheKey: locCfg.CacheKey,
		}

		if locations[index].Logger, err = logger.New(locCfg.Logger); err != nil {
			return nil, err
		}

		if locCfg.UpstreamType != "" || locCfg.UpstreamAddress != nil {
			if locations[index].Upstream, err = upstream.New(locCfg.UpstreamType, locCfg.UpstreamAddress); err != nil {
				return nil, err
			}
		}

		if locCfg.CacheZone != nil {
			orchestrator, ok := a.orchestrators[locCfg.CacheZone.ID]
			if !ok {
				return nil, fmt.Errorf("Could not get the cache zone for locations[index] %s", locCfg)
			}
			locations[index].Orchestrator = orchestrator
		}

		if locations[index].Handler, err = handler.New(locCfg.HandlerType.Type, locCfg.HandlerType.Settings, locations[index]); err != nil {
			return nil, err
		}

	}

	return locations, nil
}
