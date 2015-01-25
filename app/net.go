package app

import (
	"fmt"
	"github.com/gophergala/nedomi/types"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gophergala/nedomi/config"
)

var ErrNoRedirects = fmt.Errorf("No redirects")

type proxyHandler struct {
	config config.HTTPSection
	app    *Application
}

func (ph *proxyHandler) FindVirtualHost(r *http.Request) *VirtualHost {

	split := strings.Split(r.Host, ":")
	vh, ok := ph.app.virtualHosts[split[0]]

	if !ok {
		return nil
	}

	return vh
}

//!TODO: Add something more than a GET requests
func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	vh := ph.FindVirtualHost(r)

	if vh == nil {
		http.NotFound(w, r)
		return
	}

	objID := types.ObjectID{CacheKey: vh.CacheKey, r.URL.String()}

	fileReader, err := vh.Storage.GetFullFile(objID)

	if err != nil {
		http.Error(w, err, 500)
		return
	}

	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return ErrNoRedirects
	}

	newUrl := vh.UpstreamUrl().ResolveReference(r.URL)

	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		log.Printf("Got error\n %s\n while making request ", err)
		return
	}

	for headerName, headerValue := range r.Header {
		req.Header.Add(headerName, strings.Join(headerValue, ","))
	}

	resp, err := client.Do(req)
	if err != nil && err != ErrNoRedirects {
		if urlError, ok := err.(*url.Error); !(ok && urlError.Err == ErrNoRedirects) {
			log.Printf("Got error\n %s\n while proxying %s to %s", err, r.URL.String(),
				newUrl.String())
			return
		}
	}

	defer resp.Body.Close()

	respHeaders := w.Header()
	for headerName, headerValue := range resp.Header {
		respHeaders.Add(headerName, strings.Join(headerValue, ","))
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func newProxyHandler(app *Application) *proxyHandler {

	return &proxyHandler{
		app:    app,
		config: app.cfg.HTTP,
	}

}
