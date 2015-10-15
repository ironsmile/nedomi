package app

import (
	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/handler/headers"
	"github.com/ironsmile/nedomi/types"
)

func headersHandlerFromLocationConfig(next types.RequestHandler, locCfg *config.Location) (*headers.Headers, error) {
	return headers.NewHeaders(next, locCfg.HeadersRewrite, config.HeadersRewrite{})
}
