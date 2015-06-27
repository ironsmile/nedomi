/*
	Package app contains the main Application struct. This struct represents the
	application and is resposible for creating and connecting all other parts of the
	software.
*/
package app

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ironsmile/nedomi/cache"
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/handler"
	"github.com/ironsmile/nedomi/storage"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/upstream"
	"github.com/ironsmile/nedomi/vhost"
)

/*
   Application is the type which represents the webserver. It is responsible for
   parsing the config and it has Start, Stop, Reload and Wait functions.
*/
type Application struct {

	// Parsed config
	cfg *config.Config

	// Used to wait for the main serving goroutine to finish
	handlerWg sync.WaitGroup

	// The listener for the main server loop
	listener net.Listener

	// HTTP Server which will use the above listener in order to server
	// clients requests.
	httpSrv *http.Server

	// This is a map from Host names to virtual host pairs. The host names which will be
	// matched against the Host heder are used as keys in this map.
	// Virtual host pair is a struct which has a *vhost.VirtualHost struct and
	// a handler.RequestHandler.
	virtualHosts map[string]*vhostPair

	// A map from cache zone ID (from the config) to CacheManager resposible for this
	// cache zone.
	cacheManagers map[uint32]cache.CacheManager

	// Channels used to signal Storage objects that files have been evicted from the
	// cache.
	removeChannels []chan types.ObjectIndex
}

type vhostPair struct {
	vhostStruct  *vhost.VirtualHost
	vhostHandler handler.RequestHandler
}

/*
	initFromConfig should be called once when starting the app. It makes all the
	connections between cache zones, virtual hosts, storage objects and upstreams.
*/
func (a *Application) initFromConfig() error {
	a.virtualHosts = make(map[string]*vhostPair)

	// cache_zone_id => CacheManager
	a.cacheManagers = make(map[uint32]cache.CacheManager)

	// cache_zone_id => Storage
	storages := make(map[uint32]storage.Storage)

	up := upstream.New(a.cfg)

	defaultCacheAlgo := a.cfg.HTTP.CacheAlgo

	for _, cfgVhost := range a.cfg.HTTP.Servers {
		var virtualHost *vhost.VirtualHost

		if !cfgVhost.IsForProxyModule() {

			vhostHandler, err := handler.New(cfgVhost.HandlerType)

			if err != nil {
				return err
			}

			virtualHost = vhost.New(*cfgVhost, nil, nil)
			a.virtualHosts[virtualHost.Name] = &vhostPair{
				vhostStruct:  virtualHost,
				vhostHandler: vhostHandler,
			}
			continue
		}

		cz := cfgVhost.GetCacheZoneSection()

		if cz == nil {
			return fmt.Errorf("Cache zone for %s was nil", cfgVhost.Name)
		}

		if cm, ok := a.cacheManagers[cz.ID]; ok {
			stor := storages[cz.ID]
			virtualHost = vhost.New(*cfgVhost, cm, stor)
		} else {
			cacheManagerAlgo := defaultCacheAlgo

			if cz.CacheAlgo != "" {
				cacheManagerAlgo = cz.CacheAlgo
			}

			cm, err := cache.NewCacheManager(cacheManagerAlgo, cz)
			if err != nil {
				return err
			}
			cm.Init()
			a.cacheManagers[cz.ID] = cm

			removeChan := make(chan types.ObjectIndex, 1000)
			cm.ReplaceRemoveChannel(removeChan)

			stor, err := storage.New("disk", *cz, cm, up)

			if err != nil {
				return fmt.Errorf("Creating storage impl: %s", err)
			}

			storages[cz.ID] = stor
			go a.cacheToStorageCommunicator(stor, removeChan)

			a.removeChannels = append(a.removeChannels, removeChan)

			virtualHost = vhost.New(*cfgVhost, cm, stor)
		}

		handlerType := cfgVhost.HandlerType

		if handlerType == "" {
			handlerType = "proxy"
		}

		vhostHandler, err := handler.New(handlerType)

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

/*
	A single goroutine running this function is created for every storage.
	CacheManagers will send to the com channel files which they wish to be removed
	from the storage.
*/
func (a *Application) cacheToStorageCommunicator(stor storage.Storage,
	com chan types.ObjectIndex) {
	for oi := range com {
		stor.DiscardIndex(oi)
	}
}

/*
   Start fires up the application.
*/
func (a *Application) Start() error {
	if a.cfg == nil {
		return errors.New("Cannot start application with emtpy config")
	}

	if err := a.initFromConfig(); err != nil {
		return err
	}

	startError := make(chan error)

	a.handlerWg.Add(1)
	go a.doServing(startError)

	if err := <-startError; err != nil {
		return err
	}

	log.Printf("Application %d started\n", os.Getpid())

	return nil
}

/*
   This routine actually starts listening and working on clients requests.
*/
func (a *Application) doServing(startErrChan chan<- error) {
	defer a.handlerWg.Done()

	a.httpSrv = &http.Server{
		Addr:           a.cfg.HTTP.Listen,
		Handler:        a,
		ReadTimeout:    time.Duration(a.cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(a.cfg.HTTP.WriteTimeout) * time.Second,
		MaxHeaderBytes: a.cfg.HTTP.MaxHeadersSize,
	}

	err := a.listenAndServe(startErrChan)

	log.Printf("Webserver stopped. %s", err)
}

/*
   Uses our own listener to make our server stoppable. Similar to
   net.http.Server.ListenAndServer only this version saves a reference to the listener
*/
func (a *Application) listenAndServe(startErrChan chan<- error) error {
	addr := a.httpSrv.Addr
	if addr == "" {
		addr = ":http"
	}
	lsn, err := net.Listen("tcp", addr)
	if err != nil {
		startErrChan <- err
		return err
	}
	a.listener = lsn
	startErrChan <- nil
	log.Printf("Webserver started on %s\n", addr)
	return a.httpSrv.Serve(lsn)
}

/*
   Stop makes sure the application is completely stopped and all of its
   goroutines and channels are finished and closed.
*/
func (a *Application) Stop() error {
	a.closeRemoveChannels()
	a.listener.Close()
	a.handlerWg.Wait()
	return nil
}

/*
   Closes all channels used for sending evicted storage objects.
*/
func (a *Application) closeRemoveChannels() {
	for _, chn := range a.removeChannels {
		close(chn)
	}
}

/*
   Reload takse a new configuration and replaces the old one with it. After succesful
   reload the things that are written in the new config will be in use.
*/
func (a *Application) Reload(cfg *config.Config) error {
	if cfg == nil {
		return errors.New("Config for realoding was nil. Reloading aborted.")
	}
	//!TODO: save the listening handler if needed
	if err := a.Stop(); err != nil {
		return err
	}
	a.cfg = cfg
	return a.Start()
}

/*
   Wait subscribes iteself to few signals and waits for any of them to be received.
   When Wait returns it is the end of the application.
*/
func (a *Application) Wait() error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM)

	for sig := range signalChan {
		if sig == syscall.SIGHUP {
			newConfig, err := config.Get()
			if err != nil {
				log.Printf("Gettin new config error: %s", err)
				continue
			}
			err = a.Reload(newConfig)
			if err != nil {
				log.Printf("Reloading failed: %s", err)
			}
		} else {
			log.Printf("Stopping %d: %s", os.Getpid(), sig)
			break
		}
	}

	if err := a.Stop(); err != nil {
		return err
	}

	return nil
}

func New(cfg *config.Config) (*Application, error) {
	return &Application{cfg: cfg}, nil
}
