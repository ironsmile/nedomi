package app

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gophergala/nedomi/config"
)

/*
   Application is the type which represents the webserver. It is responsible for
   parsing the config and it has Start, Stop, Reload and Wait functions.
*/
type Application struct {
	cfg *config.Config
}

/*
   Start fires up the application.
*/
func (a *Application) Start() error {
	if a.cfg == nil {
		return errors.New("Cannot start application with emtpy config")
	}

	log.Printf("Application %d started\n", os.Getpid())
	return nil
}

/*
   Stop makes sure the application is completely stopped and all of its
   goroutines and channels are finished and closed.
*/
func (a *Application) Stop() error {
	return nil
}

/*
   Reload takse a new configuration and replaces the old one with it. After succesful
   reload the things that are written in the new config will be in use.
*/
func (a *Application) Reload(cfg *config.Config) error {
	a.cfg = cfg
	return nil
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
