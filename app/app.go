package app

import (
	"github.com/gophergala/nedomi/config"
)

type Application struct {
	cfg *config.Config
}

func (a *Application) Start() error {
	return nil
}

func (a *Application) WaitForSignals() error {
	return nil
}

func New(cfg *config.Config) (*Application, error) {
	return &Application{cfg: cfg}, nil
}
