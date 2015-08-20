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
	// Make the vhost and storage maps
	a.virtualHosts = make(map[string]*types.VirtualHost)
	a.storages = make(map[string]types.Storage)

	// Create a global application context
	a.ctx, a.ctxCancel = context.WithCancel(context.Background())

	// Initialize the global logger
	defaultLogger, err := logger.New(a.cfg.Logger)
	if err != nil {
		return err
	}
	a.logger = defaultLogger

	// Initialize all cache storages
	for _, cfgStorage := range a.cfg.CacheZones {
		//!TODO: the cache zone should be responsible for it's own algorithm
		ca, err := cache.New(cfgStorage)
		if err != nil {
			return err
		}

		removeChan := make(chan types.ObjectIndex, 1000)
		ca.ReplaceRemoveChannel(removeChan)
		stor, err := storage.New(*cfgStorage, ca, a.logger)
		if err != nil {
			return fmt.Errorf("Could not initialize storage '%s' impl: %s", cfgStorage.Type, err)
		}

		a.storages[cfgStorage.ID] = stor
		go a.cacheToStorageCommunicator(stor, removeChan)
		a.removeChannels = append(a.removeChannels, removeChan)
	}

	// Initialize all vhosts
	for _, cfgVhost := range a.cfg.HTTP.Servers {
		vhost := types.VirtualHost{
			Name:            cfgVhost.Name,
			CacheKey:        cfgVhost.CacheKey,
			UpstreamAddress: cfgVhost.UpstreamAddress,
			Logger:          a.logger,
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

		stor, ok := a.storages[cfgVhost.CacheZone.ID]
		if !ok {
			return fmt.Errorf("Could not get the cache zone for vhost %s", cfgVhost.Name)
		}
		vhost.Storage = stor
	}

	a.ctx = contexts.NewStoragesContext(a.ctx, a.storages)

	return nil
}
