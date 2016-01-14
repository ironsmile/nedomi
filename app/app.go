// Package app contains the main Application struct. This struct represents the
// application and is resposible for creating and connecting all other parts of
// the software.
package app

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"github.com/MStoykov/grace/gracehttp"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/netutils"
)

// Application is the type which represents the webserver. It is responsible for
// parsing the config and it has Start, Stop, Reload and Wait functions.
type Application struct {
	types.SyncLogger
	sync.RWMutex
	// the way for the application to get it's config if it's changed
	configGetter config.Getter

	// Parsed config
	cfg *config.Config

	// Used to wait for the main serving goroutine to finish
	handlerWg sync.WaitGroup

	// HTTP Server which will use the above listener in order to server
	// clients requests.
	httpSrv *http.Server

	// This is a map from Host names to virtual host pairs. The host names which will be
	// matched against the Host heder are used as keys in this map.
	// Virtual host pair is a struct which has a *VirtualHost struct and
	// a types.RequestHandler.
	virtualHosts map[string]*VirtualHost

	notConfiguredHandler types.RequestHandler

	// A map from cache zone ID (from the config) to types.CacheZone
	// that is resposible for this cache zone.
	cacheZones map[string]*types.CacheZone

	// A map with all simple and advanced upstream transports
	upstreams map[string]types.Upstream

	// The global application context. It is cancelled when stopping or
	// reloading the application.
	ctx context.Context

	// The cancel function for the global application context.
	ctxCancel func()

	stats *applicationStats

	started time.Time

	version types.AppVersion

	conns *connections
}

// Stats returns application wide stats
func (a *Application) Stats() types.AppStats {
	return (types.AppStats)(*a.stats)
}

// Run fires up the application. And Blocks until it ends
func (a *Application) Run() error {
	if err := SetupEnv(a.cfg); err != nil {
		return err
	}
	a.started = time.Now()
	if a.cfg == nil {
		return errors.New("Cannot start application with emtpy config")
	}

	if err := a.reinitFromConfig(a.cfg, false); err != nil {
		return err
	}

	a.handlerWg.Add(1)
	go a.doServing()

	a.GetLogger().Logf("Application %d started", os.Getpid())

	defer func() {
		if err := CleanupEnv(a.cfg); err != nil {
			a.GetLogger().Logf("error on env cleanup %s", err)
		}
	}()
	return a.Wait()
}

// This routine actually starts listening and working on clients requests.
func (a *Application) doServing() {
	defer a.handlerWg.Done()

	a.httpSrv = &http.Server{
		Addr:           a.cfg.HTTP.Listen,
		Handler:        a,
		ReadTimeout:    time.Duration(a.cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(a.cfg.HTTP.WriteTimeout) * time.Second,
		MaxHeaderBytes: a.cfg.HTTP.MaxHeadersSize,
		ConnState: func(input net.Conn, state http.ConnState) {
			conn := input.(types.IncomingConn)
			switch state {
			case http.StateNew:
				a.conns.add(conn)
			case http.StateClosed:
				a.conns.remove(conn)
			}
		},
	}

	err := a.listenAndServe()

	a.GetLogger().Logf("Webserver stopped. %s", err)
}

// Uses our own listener to make our server stoppable. Similar to
// net.http.Server.ListenAndServer only this version saves a reference to the listener
func (a *Application) listenAndServe() error {
	var deadlineToTimeoutListener = netutils.DeadlineToTimeoutListenerConstructor(
		int64(a.cfg.HTTP.MaxIOTransferSize),
		int64(a.cfg.HTTP.MinIOTransferSize),
	)
	// Serve accepts incoming connections on the Listener lsn, creating a
	// new service goroutine for each.  The service goroutines read requests and
	// then call the handler (i.e. ServeHTTP() ) to reply to them.
	return gracehttp.ServeWithWrapper(deadlineToTimeoutListener, a.httpSrv)
}

// Stop makes sure the application is completely stopped and all of its
// goroutines and channels are finished and closed.
func (a *Application) Stop() error {
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	err = process.Signal(syscall.SIGTERM)
	a.handlerWg.Wait()
	a.ctxCancel()
	return err
}

// Reload takse a new configuration and replaces the old one with it. After succesful
// reload the things that are written in the new config will be in use.
func (a *Application) Reload(cfg *config.Config) error {
	if err := a.checkConfigCouldBeReloaded(cfg); err != nil {
		return err
	}
	return a.reinitFromConfig(cfg, false)
}

// Wait subscribes iteself to few signals and waits for any of them to be received.
// When Wait returns it is the end of the application.
func (a *Application) Wait() error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM)

	for sig := range signalChan {
		if sig == syscall.SIGHUP {
			newConfig, err := a.configGetter()
			if err != nil {
				a.GetLogger().Errorf("Getting new config error: %s", err)
				continue
			}
			err = a.Reload(newConfig)
			if err != nil {
				a.GetLogger().Errorf("Reloading failed: %s", err)
			}
		} else {
			a.GetLogger().Logf("Stopping %d: %s", os.Getpid(), sig)
			break
		}
	}

	if err := a.Stop(); err != nil {
		return err
	}

	return nil
}

// New creates and returns a new Application with the specified config.
func New(version types.AppVersion, configGetter config.Getter) (*Application, error) {
	var cfg, err = configGetter()
	if err != nil {
		return nil, err
	}
	var a = &Application{
		version:      version,
		cfg:          cfg,
		stats:        new(applicationStats),
		configGetter: configGetter,
		conns:        newConnections(),
		cacheZones:   make(map[string]*types.CacheZone),
	}
	a.ctx, a.ctxCancel = context.WithCancel(context.Background())
	a.ctx = contexts.NewAppContext(a.ctx, a)
	a.ctx = contexts.NewCacheZonesContext(a.ctx, a.cacheZones)
	if err = a.reinitFromConfig(a.cfg, true); err != nil {
		return nil, err
	}

	return a, nil
}

// Version returns application version
func (a *Application) Version() types.AppVersion {
	return a.version
}

// Started returns when the application was started
func (a *Application) Started() time.Time {
	return a.started
}

// GetUpstream gets a configured upstream by it's id
func (a *Application) GetUpstream(id string) types.Upstream {
	return a.upstreams[id]
}
