package app

import (
	"fmt"
	"github.com/gophergala/nedomi/types"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
//!TODO: Rewrite Date header
func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	vh := ph.FindVirtualHost(r)

	if vh == nil {
		log.Printf("404 %s", r.RequestURI)
		http.NotFound(w, r)
		return
	}

	log.Printf("[%p] Access %s", r, r.RequestURI)

	rng := r.Header.Get("Range")

	if rng != "" {
		ph.ServerPartialRequest(w, r, vh)
		return
	} else {
		ph.ServeFullRequest(w, r, vh)
		return
	}

}

func (p *proxyHandler) ServerPartialRequest(w http.ResponseWriter, r *http.Request,
	vh *VirtualHost) {
	objID := types.ObjectID{CacheKey: vh.CacheKey, Path: r.URL.String()}

	fileHeaders, err := vh.Storage.Headers(objID)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file headers. %s\n", r, err)
		return
	}

	rng := r.Header.Get("Range")

	parts := strings.Split(rng, "-")

	if len(parts) != 2 {
		http.Error(w, fmt.Sprintf("[%p] Range received: %s", r, rng), 416)
	}

	start, err := strconv.Atoi(parts[0])

	if err != nil {
		http.Error(w, fmt.Sprintf("[%p] Range received: %s", r, rng), 416)
	}

	end, err := strconv.Atoi(parts[0])

	if err != nil {
		http.Error(w, fmt.Sprintf("[%p] Range received: %s", r, rng), 416)
	}

	fileReader, err := vh.Storage.Get(objID, uint64(start), uint64(end))

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file reader error. %s\n", r, err)
		return
	}

	respHeaders := w.Header()
	for headerName, headerValue := range fileHeaders {
		respHeaders.Add(headerName, strings.Join(headerValue, ","))
	}

	log.Printf("[%p] 206 Range %s %s", r, r.Header.Get("Range"), r.RequestURI)
	w.WriteHeader(206)
	io.Copy(w, fileReader)
}

func (p *proxyHandler) ServeFullRequest(w http.ResponseWriter, r *http.Request,
	vh *VirtualHost) {
	objID := types.ObjectID{CacheKey: vh.CacheKey, Path: r.URL.String()}

	fileHeaders, err := vh.Storage.Headers(objID)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file headers. %s\n", r, err)
		return
	}

	fileReader, err := vh.Storage.GetFullFile(objID)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file reader error. %s\n", r, err)
		return
	}

	respHeaders := w.Header()
	for headerName, headerValue := range fileHeaders {
		respHeaders.Add(headerName, strings.Join(headerValue, ","))
	}

	log.Printf("[%p] 200 %s", r, r.RequestURI)
	w.WriteHeader(200)
	_, err = io.Copy(w, fileReader)
	if err != nil {
		log.Println(err)
	}
}

func (p *proxyHandler) ProxyRequest(w http.ResponseWriter, r *http.Request,
	vh *VirtualHost) {
	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return ErrNoRedirects
	}

	newUrl := vh.UpstreamUrl().ResolveReference(r.URL)

	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		log.Printf("[%p] Got error\n %s\n while making request ", r, err)
		return
	}

	for headerName, headerValue := range r.Header {
		req.Header.Add(headerName, strings.Join(headerValue, ","))
	}

	resp, err := client.Do(req)
	if err != nil && err != ErrNoRedirects {
		if urlError, ok := err.(*url.Error); !(ok && urlError.Err == ErrNoRedirects) {
			log.Printf("[%p] Got error\n %s\n while proxying %s to %s", r, err,
				r.URL.String(), newUrl.String())
			return
		}
	}

	defer resp.Body.Close()

	respHeaders := w.Header()
	for headerName, headerValue := range resp.Header {
		respHeaders.Add(headerName, strings.Join(headerValue, ","))
	}

	log.Printf("[%p] %d Proxied %s", r, resp.StatusCode, r.RequestURI)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func newProxyHandler(app *Application) *proxyHandler {

	return &proxyHandler{
		app:    app,
		config: app.cfg.HTTP,
	}

}
