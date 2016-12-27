package app

import (
	"net/http"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/handler/headers"
)

func headersHandlerFromLocationConfig(next http.Handler, locCfg *config.Location) (*headers.Headers, error) {
	return headers.NewHeaders(next, locCfg.HeadersRewrite, config.HeadersRewrite{})
}
