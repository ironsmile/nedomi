package app

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
)

// initFromConfig should be called when starting or reloading the app. It makes
// all the connections between cache zones, virtual hosts, storage objects
// and upstreams.
func (a *Application) initFromConfig() error {
	// vhost_name => types.VirtualHost
	a.virtualHosts = make(map[string]*types.VirtualHost)
	// cache_zone_id => types.Storage
	a.storages = make(map[string]types.Storage)

	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	defaultLogger, err := logger.New(a.cfg.Logger.Type, a.cfg.Logger)
	if err != nil {
		return err
	}
	a.logger = defaultLogger

	for _, cfgVhost := range a.cfg.HTTP.Servers {
		var vhostLogger logger.Logger
		if cfgVhost.Logger != nil {
			vhostLogger, err = logger.New(cfgVhost.Logger.Type, *cfgVhost.Logger)
			if err != nil {
				return err
			}
		} else {
			vhostLogger = a.logger
		}

		if cfgVhost.HandlerType != "proxy" {

			vhostHandler, err := handler.New(cfgVhost.HandlerType)
			if err != nil {
				return err
			}

			a.virtualHosts[cfgVhost.Name] = &types.VirtualHost{
				Name:    cfgVhost.Name,
				Handler: vhostHandler,
			}
			continue
		}

		cz := cfgVhost.CacheZone
		if cz == nil {
			return fmt.Errorf("Cache zone for %s was nil", cfgVhost.Name)
		}

		up, err := upstream.New(cfgVhost.UpstreamType, cfgVhost.UpstreamAddress)
		if err != nil {
			return err
		}

		stor, ok := a.storages[cz.ID]
		if !ok {
			//!TODO: the cache zone should be responsible for it's own algorithm
			ca, err := cache.New(cz)
			if err != nil {
				return err
			}

			removeChan := make(chan types.ObjectIndex, 1000)
			ca.ReplaceRemoveChannel(removeChan)

			stor, err = storage.New(cz.Type, *cz, ca, vhostLogger)

			if err != nil {
				return fmt.Errorf("Creating storage impl: %s", err)
			}

			a.storages[cz.ID] = stor
			go a.cacheToStorageCommunicator(stor, removeChan)

			a.removeChannels = append(a.removeChannels, removeChan)
		}

		vhostHandler, err := handler.New(cfgVhost.HandlerType)
		if err != nil {
			return err
		}

		a.virtualHosts[cfgVhost.Name] = &types.VirtualHost{
			Name:            cfgVhost.Name,
			CacheKey:        cfgVhost.CacheKey,
			Handler:         vhostHandler,
			Storage:         stor,
			Upstream:        up,
			UpstreamAddress: cfgVhost.UpstreamAddress,
		}
	}

	a.ctx = contexts.NewStoragesContext(a.ctx, a.storages)

	return nil
}
