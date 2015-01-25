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

	"github.com/gophergala/nedomi/upstream"

	"github.com/gophergala/nedomi/cache"
	"github.com/gophergala/nedomi/config"
	"github.com/gophergala/nedomi/storage"
	"github.com/gophergala/nedomi/types"
)

/*
   Application is the type which represents the webserver. It is responsible for
   parsing the config and it has Start, Stop, Reload and Wait functions.
*/
type Application struct {
	cfg *config.Config

	handlerWg sync.WaitGroup

	// The http handler for the main server loop
	httpHandler http.Handler

	// The listener for the main server loop
	listener net.Listener

	// HTTP Server which will use the above listener in order to server
	// clients requests.
	httpSrv *http.Server

	virtualHosts  map[string]*VirtualHost
	cacheManagers map[uint32]cache.CacheManager

	removeChannels []chan types.ObjectIndex
}

func (a *Application) initFromConfig() error {
	a.virtualHosts = make(map[string]*VirtualHost)

	// cache_zone_id => CacheManager
	a.cacheManagers = make(map[uint32]cache.CacheManager)

	// cache_zone_id => Storage
	storages := make(map[uint32]storage.Storage)

	up, _ := upstream.New("http://doycho.com:9996")

	for _, vh := range a.cfg.HTTP.Servers {
		cz := vh.GetCacheZoneSection()

		if cz == nil {
			return fmt.Errorf("Cache zone for %s was nil", vh.Name)
		}

		var virtualHost *VirtualHost

		if cm, ok := a.cacheManagers[cz.ID]; ok {
			stor := storages[cz.ID]
			virtualHost = &VirtualHost{*vh, cm, stor}
		} else {
			cm, err := cache.NewCacheManager("lru", cz)
			if err != nil {
				return err
			}
			cm.Init()
			a.cacheManagers[cz.ID] = cm

			removeChan := make(chan types.ObjectIndex, 1000)
			cm.ReplaceRemoveChannel(removeChan)

			stor := storage.NewStorage(*cz, cm, up)

			storages[cz.ID] = stor
			go a.cacheToStorageCommunicator(stor, removeChan)

			a.removeChannels = append(a.removeChannels, removeChan)

			virtualHost = &VirtualHost{*vh, cm, stor}
		}

		a.virtualHosts[virtualHost.Name] = virtualHost
	}

	return nil
}

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

	a.httpHandler = newProxyHandler(a)

	a.httpSrv = &http.Server{
		Addr:           a.cfg.HTTP.Listen,
		Handler:        a.httpHandler,
		ReadTimeout:    time.Duration(a.cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(a.cfg.HTTP.WriteTimeout) * time.Second,
		MaxHeaderBytes: a.cfg.HTTP.MaxHeadersSize,
	}

	err := a.listenAndServe(startErrChan)

	log.Printf("Webserver stopped. %s", err)
}

// Uses our own listener to make our server stoppable. Similar to
// net.http.Server.ListenAndServer only this version saves a reference to the listener
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
	log.Println("Webserver started.")
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
