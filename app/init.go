package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/logger"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
	"github.com/ironsmile/nedomi/utils"
)

func (a *Application) reinitFromConfig() (err error) {
	app := *a // copy

	app.virtualHosts = make(map[string]*VirtualHost)
	app.upstreams = make(map[string]types.Upstream)
	app.cacheZones = make(map[string]*types.CacheZone)
	logs := accessLogs{"": nil}
	// Initialize the global logger
	var l types.Logger
	if l, err = logger.New(&app.cfg.Logger); err != nil {
		return err
	}

	app.SetLogger(l)

	var toBeResized = make([]string, 0, len(a.cacheZones))
	// Initialize all cache zones
	for _, cfgCz := range app.cfg.CacheZones {
		if zone, ok := a.cacheZones[cfgCz.ID]; ok {
			app.cacheZones[cfgCz.ID] = zone
			toBeResized = append(toBeResized, cfgCz.ID)
			continue
		}
		if err = app.initCacheZone(cfgCz); err != nil {
			return err
		}
	}

	// Initialize all advanced upstreams
	for _, cfgUp := range app.cfg.HTTP.Upstreams {
		if app.upstreams[cfgUp.ID], err = upstream.New(cfgUp, l); err != nil {
			return err
		}
	}

	app.notConfiguredHandler = newNotConfiguredHandler()
	var accessLog io.Writer
	if accessLog, err = logs.openAccessLog(app.cfg.HTTP.AccessLog); err != nil {
		return err
	}
	app.notConfiguredHandler, _ = loggingHandler(app.notConfiguredHandler, accessLog, false)
	// Initialize all vhosts
	for _, cfgVhost := range app.cfg.HTTP.Servers {
		if err = app.initVirtualHost(cfgVhost, logs); err != nil {
			return err
		}
	}

	a.Lock()
	defer a.Unlock()
	a.SetLogger(app.GetLogger())
	a.virtualHosts = app.virtualHosts
	a.upstreams = app.upstreams
	a.notConfiguredHandler = app.notConfiguredHandler
	for id := range a.cacheZones { // clean the cacheZones
		delete(a.cacheZones, id)
	}
	for _, id := range toBeResized { // resize the to be resized
		var cfgCz = app.cfg.CacheZones[id]
		var zone = app.cacheZones[id]
		zone.Storage.SetLogger(app.GetLogger())
		zone.Scheduler.SetLogger(app.GetLogger())
		zone.Algorithm.SetLogger(app.GetLogger())
		zone.Algorithm.ChangeConfig(cfgCz.BulkRemoveTimeout, cfgCz.BulkRemoveCount, cfgCz.StorageObjects)
	}
	for id, zone := range app.cacheZones { // copy everything
		a.cacheZones[id] = zone
	}

	return nil
}

// initFromConfig should be called when starting or reloading the app. It makes
// all the connections between cache zones, virtual hosts and upstreams.
func (a *Application) initFromConfig() (err error) {
	// Make the vhost and cacheZone maps
	a.cacheZones = make(map[string]*types.CacheZone)

	a.ctx, a.ctxCancel = context.WithCancel(context.Background())
	a.ctx = contexts.NewAppContext(a.ctx, a)
	a.ctx = contexts.NewCacheZonesContext(a.ctx, a.cacheZones)
	return a.reinitFromConfig()
}

func (a *Application) initCacheZone(cfgCz *config.CacheZone) (err error) {
	cz := &types.CacheZone{
		ID:        cfgCz.ID,
		PartSize:  cfgCz.PartSize,
		Scheduler: storage.NewScheduler(a.GetLogger()),
	}
	// Initialize the storage
	if cz.Storage, err = storage.New(cfgCz, a.GetLogger()); err != nil {
		return fmt.Errorf("Could not initialize storage '%s' for cache zone '%s': %s",
			cfgCz.Type, cfgCz.ID, err)
	}

	// Initialize the cache algorithm
	if cz.Algorithm, err = cache.New(cfgCz, cz.Storage.DiscardPart, a.GetLogger()); err != nil {
		return fmt.Errorf("Could not initialize algorithm '%s' for cache zone '%s': %s",
			cfgCz.Algorithm, cfgCz.ID, err)
	}

	a.reloadCache(cz)
	a.cacheZones[cfgCz.ID] = cz

	return nil
}

func (a *Application) getUpstream(upID string) (types.Upstream, error) {
	if upID == "" {
		return nil, nil
	}

	if up, ok := a.upstreams[upID]; ok {
		return up, nil
	}

	if upURL, err := url.Parse(upID); err == nil {
		up, err := upstream.NewSimple(upURL)
		if err != nil {
			return nil, err
		}
		a.upstreams[upID] = up
		return up, nil
	}

	return nil, fmt.Errorf("Invalid upstream %s", upID)
}

func (a *Application) initVirtualHost(cfgVhost *config.VirtualHost, logs accessLogs) (err error) {
	var accessLog io.Writer
	if cfgVhost.AccessLog != "" {
		if accessLog, err = logs.openAccessLog(cfgVhost.AccessLog); err != nil {
			return fmt.Errorf("error opening access log for virtual host %s - %s",
				cfgVhost.Name, err)
		}
	}

	vhost := VirtualHost{
		Location: types.Location{
			Name:                  cfgVhost.Name,
			CacheKey:              cfgVhost.CacheKey,
			CacheKeyIncludesQuery: cfgVhost.CacheKeyIncludesQuery,
			CacheDefaultDuration:  cfgVhost.CacheDefaultDuration,
		},
	}
	if vhost.Upstream, err = a.getUpstream(cfgVhost.Upstream); err != nil {
		return err
	}

	if _, ok := a.virtualHosts[cfgVhost.Name]; ok {
		return fmt.Errorf("Virtual host or alias %s already exists", cfgVhost.Name)
	}
	a.virtualHosts[cfgVhost.Name] = &vhost

	for _, alias := range cfgVhost.Aliases {
		if _, ok := a.virtualHosts[alias]; ok {
			return fmt.Errorf("Virtual host or alias %s already exists, duplicated by alias for %s",
				alias, cfgVhost.Name)
		}
		a.virtualHosts[alias] = &vhost
	}

	if vhost.Logger, err = logger.New(&cfgVhost.Logger); err != nil {
		return err
	}

	if cfgVhost.CacheZone != nil {
		cz, ok := a.cacheZones[cfgVhost.CacheZone.ID]
		if !ok {
			return fmt.Errorf("Could not get the cache zone for vhost %s", cfgVhost.Name)
		}
		vhost.Cache = cz
	}

	if vhost.Handler, err = chainHandlers(&vhost.Location, &cfgVhost.Location, accessLog); err != nil {
		return err
	}
	var locations []*types.Location
	if locations, err = a.initFromConfigLocationsForVHost(cfgVhost.Locations, accessLog); err != nil {
		return err
	}

	if vhost.Muxer, err = NewLocationMuxer(locations); err != nil {
		return fmt.Errorf("Could not create location muxer for vhost %s - %s", cfgVhost.Name, err)
	}

	return nil
}

func (a *Application) initFromConfigLocationsForVHost(cfgLocations []*config.Location, accessLog io.Writer) ([]*types.Location, error) {
	var err error
	var locations = make([]*types.Location, len(cfgLocations))
	for index, locCfg := range cfgLocations {
		locations[index] = &types.Location{
			Name:                  locCfg.Name,
			CacheKey:              locCfg.CacheKey,
			CacheKeyIncludesQuery: locCfg.CacheKeyIncludesQuery,
			CacheDefaultDuration:  locCfg.CacheDefaultDuration,
		}
		if locations[index].Upstream, err = a.getUpstream(locCfg.Upstream); err != nil {
			return nil, err
		}

		if locations[index].Logger, err = logger.New(&locCfg.Logger); err != nil {
			return nil, err
		}

		if locCfg.CacheZone != nil {
			cz, ok := a.cacheZones[locCfg.CacheZone.ID]
			if !ok {
				return nil, fmt.Errorf("Could not get the cache zone for locations[index] %s", locCfg.Name)
			}
			locations[index].Cache = cz
		}

		if locations[index].Handler, err = chainHandlers(locations[index], locCfg, accessLog); err != nil {
			return nil, err
		}

	}

	return locations, nil
}

func (a *Application) reloadCache(cz *types.CacheZone) {
	counter := 0
	callback := func(obj *types.ObjectMetadata, parts ...*types.ObjectIndex) bool {
		counter++
		//!TODO: remove hardcoded periods and timeout, get them from config
		if counter%100 == 0 {
			select {
			case <-a.ctx.Done():
				return false
			case <-time.After(100 * time.Millisecond):
			}
		}

		if !utils.IsMetadataFresh(obj) {
			if err := cz.Storage.Discard(obj.ID); err != nil {
				a.GetLogger().Errorf("Error for cache zone `%s` on discarding objID `%s` in reloadCache: %s", cz.ID, obj.ID, err)
			}
		} else {
			cz.Scheduler.AddEvent(
				obj.ID.Hash(),
				storage.GetExpirationHandler(cz, obj.ID),
				//!TODO: Maybe do not use time.Now but cached time. See the todo comment
				// in utils.IsMetadataFresh.
				time.Unix(obj.ExpiresAt, 0).Sub(time.Now()),
			)

			for _, idx := range parts {
				if err := cz.Algorithm.AddObject(idx); err != nil && err != types.ErrAlreadyInCache {
					a.GetLogger().Errorf("Error for cache zone `%s` on adding objID `%s` in reloadCache: %s", cz.ID, obj.ID, err)
				}
			}
		}

		return true
	}

	go func() {
		var ch = make(chan struct{})
		defer close(ch)
		go func() {
			const tick = 10 * time.Second

			var ticker = time.NewTicker(tick)
			var ticks int64
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					ticks++
					a.GetLogger().Logf("Storage reload for cache zone `%s` has reloaded %d for %s and is still going", cz.ID, counter, time.Duration(ticks)*tick)
				case <-ch:
					return
				}
			}
		}()
		a.GetLogger().Logf("Start storage reload for cache zone `%s`", cz.ID)
		if err := cz.Storage.Iterate(callback); err != nil {
			a.GetLogger().Errorf("For cache zone `%s` received iterator error '%s' after loading %d objects", cz.ID, err, counter)
		} else {
			a.GetLogger().Logf("Loading contents from disk for cache zone `%s` finished: %d objects loaded!", cz.ID, counter)
		}
	}()
}

func chainHandlers(location *types.Location, locCfg *config.Location, accessLog io.Writer) (types.RequestHandler, error) {
	var res types.RequestHandler
	var err error
	var handlers = locCfg.Handlers
	for index := len(handlers) - 1; index >= 0; index-- {
		if res, err = handler.New(&handlers[index], location, res); err != nil {
			return nil, err
		}
	}
	res, err = headersHandlerFromLocationConfig(res, locCfg)
	if err != nil {
		return nil, err
	}
	return loggingHandler(res, accessLog, true)
}

// loggingHandler will write to accessLog each and every request to it while proxing
// it to next
func loggingHandler(next types.RequestHandler, accessLog io.Writer, knownVhost bool) (
	types.RequestHandler,
	error,
) {

	if next == nil {
		return nil, types.NilNextHandler("accessLog")
	}

	if accessLog == nil {
		return next, nil
	}

	return types.RequestHandlerFunc(
		func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			t := time.Now()
			l := &responseLogger{ResponseWriter: w}
			url := *r.URL
			reqID, _ := contexts.GetRequestID(ctx)

			vhostID := r.Host

			if !knownVhost {
				vhostID += unknownVhostLogSuffix
			}

			defer func(vhostID string) {
				go func() {
					writeLog(accessLog, r, vhostID, reqID, url, t, l.Status(), l.Size())
				}()
			}(vhostID)
			next.RequestHandle(ctx, l, r)
		}), nil
}

// This will make the access log line for uknown vhosts to include something like
// 127.0.0.1 -> z9-u19.ucdn-domains.com.[unknown-location].~* \.flv$
// This can be useful for grepping through the access logs
const unknownVhostLogSuffix = ".[unknown-location]"
