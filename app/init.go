package app

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

// initFromConfig should be called when starting or reloading the app. It makes
// all the connections between cache zones, virtual hosts and upstreams.
func (a *Application) initFromConfig() (err error) {
	// Make the vhost and cacheZone maps
	a.virtualHosts = make(map[string]*VirtualHost)
	a.cacheZones = make(map[string]types.CacheZone)

	// Create a global application context
	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	// Initialize the global logger
	if a.logger, err = logger.New(&a.cfg.Logger); err != nil {
		return err
	}

	// Initialize all cache zones
	for _, cfgCz := range a.cfg.CacheZones {
		cz := types.CacheZone{ID: cfgCz.ID, PartSize: cfgCz.PartSize}
		// Initialize the storage
		if cz.Storage, err = storage.New(cfgCz, a.logger); err != nil {
			return fmt.Errorf("Could not initialize storage '%s' for cache zone '%s': %s",
				cfgCz.Type, cfgCz.ID, err)
		}

		// Initialize the cache algorithm
		if cz.Algorithm, err = cache.New(cfgCz, cz.Storage.DiscardPart, a.logger); err != nil {
			return fmt.Errorf("Could not initialize algorithm '%s' for cache zone '%s': %s",
				cfgCz.Algorithm, cfgCz.ID, err)
		}

		//!TODO: init sync of previous storage objects to the cache
		a.cacheZones[cfgCz.ID] = cz
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
			cz, ok := a.cacheZones[cfgVhost.CacheZone.ID]
			if !ok {
				return fmt.Errorf("Could not get the cache zone for vhost %s", cfgVhost.Name)
			}
			vhost.Cache = cz
		}

		if vhost.Handler, err = adapt(&vhost.Location, cfgVhost.Handlers); err != nil {
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

	a.ctx = contexts.NewCacheZonesContext(a.ctx, a.cacheZones)

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
			cz, ok := a.cacheZones[locCfg.CacheZone.ID]
			if !ok {
				return nil, fmt.Errorf("Could not get the cache zone for locations[index] %s", locCfg.Name)
			}
			locations[index].Cache = cz
		}

		if locations[index].Handler, err = adapt(locations[index], locCfg.Handlers); err != nil {
			return nil, err
		}

	}

	return locations, nil
}

func adapt(location *types.Location, handlers []config.Handler) (types.RequestHandler, error) {
	var res types.RequestHandler
	var err error
	for index := len(handlers) - 1; index >= 0; index-- {
		if res, err = handler.New(&handlers[index], location, res); err != nil {
			return nil, err
		}
	}
	return res, nil
}
