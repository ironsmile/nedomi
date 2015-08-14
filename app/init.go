package app

import (
	"fmt"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
	"github.com/ironsmile/nedomi/vhost"
)

// initFromConfig should be called when starting or reloading the app. It makes
// all the connections between cache zones, virtual hosts, storage objects
// and upstreams.
func (a *Application) initFromConfig() error {
	// vhost_name => vhostPair
	a.virtualHosts = make(map[string]*vhostPair)
	// cache_zone_id => cache.Algorithm
	a.cacheAlgorithms = make(map[string]cache.Algorithm)
	// cache_zone_id => storage.Storage
	storages := make(map[string]storage.Storage)

	//!TODO: add logger in app instance, use it after initFromConfig() is done instead of the default logger
	defaultLogger, err := logger.New(a.cfg.Logger.Type, a.cfg.Logger)
	if err != nil {
		return err
	}

	for _, cfgVhost := range a.cfg.HTTP.Servers {
		var vhostLogger logger.Logger
		if cfgVhost.Logger != nil {
			vhostLogger, err = logger.New(cfgVhost.Logger.Type, *cfgVhost.Logger)
			if err != nil {
				return err
			}
		} else {
			vhostLogger = defaultLogger
		}

		var virtualHost *vhost.VirtualHost

		if cfgVhost.HandlerType != "proxy" {

			vhostHandler, err := handler.New(cfgVhost.HandlerType)

			if err != nil {
				return err
			}

			virtualHost = vhost.New(*cfgVhost, nil, nil, nil)
			a.virtualHosts[virtualHost.Name] = &vhostPair{
				vhostStruct:  virtualHost,
				vhostHandler: vhostHandler,
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

		if cm, ok := a.cacheAlgorithms[cz.ID]; ok {
			stor := storages[cz.ID]
			virtualHost = vhost.New(*cfgVhost, cm, stor, up)
		} else {
			cm, err := cache.New(cz.Algorithm, cz)
			if err != nil {
				return err
			}
			a.cacheAlgorithms[cz.ID] = cm

			removeChan := make(chan types.ObjectIndex, 1000)
			cm.ReplaceRemoveChannel(removeChan)

			stor, err := storage.New(cz.Type, *cz, cm, up, vhostLogger)

			if err != nil {
				return fmt.Errorf("Creating storage impl: %s", err)
			}

			storages[cz.ID] = stor
			go a.cacheToStorageCommunicator(stor, removeChan)

			a.removeChannels = append(a.removeChannels, removeChan)

			virtualHost = vhost.New(*cfgVhost, cm, stor, up)
		}

		vhostHandler, err := handler.New(cfgVhost.HandlerType)
		if err != nil {
			return err
		}

		a.virtualHosts[virtualHost.Name] = &vhostPair{
			vhostStruct:  virtualHost,
			vhostHandler: vhostHandler,
		}
	}

	return nil
}
