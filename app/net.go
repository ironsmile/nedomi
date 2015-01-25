package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gophergala/nedomi/config"
	"github.com/gophergala/nedomi/types"
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

func (ph *proxyHandler) DefaultServer(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "" && r.RequestURI == ph.config.StatusPage {
		ph.StatusPage(w, r)
		return
	}

	log.Printf("[%p] 404 %s", r, r.RequestURI)
	http.NotFound(w, r)
	return
}

//!TODO: Add something more than a GET requests
//!TODO: Rewrite Date header
func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	vh := ph.FindVirtualHost(r)

	if vh == nil {
		ph.DefaultServer(w, r)
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

	cl := fileHeaders.Get("Content-Length")
	contentLength, err := strconv.ParseInt(cl, 10, 64)

	if err != nil {
		w.Header().Set("Content-Range", "*/*")
		msg := fmt.Sprintf("File content-length was not parsed: %s. %s", cl, err)
		log.Printf("[%p] %s", r, msg)
		http.Error(w, msg, 416)
		return
	}

	ranges, err := parseRange(r.Header.Get("Range"), contentLength)

	if err != nil {
		w.Header().Set("Content-Range", "*/*")
		msg := fmt.Sprintf("Bytes range error: %s. %s", r.Header.Get("Range"), err)
		log.Printf("[%p] %s", r, msg)
		http.Error(w, msg, 416)
		return
	}

	if len(ranges) != 1 {
		w.Header().Set("Content-Range", "*/*")
		msg := fmt.Sprintf("We support only one set of bytes ranges. Got %d", len(ranges))
		log.Printf("[%p] %s", r, msg)
		http.Error(w, msg, 416)
		return
	}

	httpRng := ranges[0]

	fileReader, err := vh.Storage.Get(objID, uint64(httpRng.start),
		uint64(httpRng.start+httpRng.length-1))

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file reader error. %s\n", r, err)
		return
	}

	defer fileReader.Close()

	respHeaders := w.Header()

	for headerName, headerValue := range fileHeaders {
		respHeaders.Set(headerName, strings.Join(headerValue, ","))
	}

	respHeaders.Set("Content-Range", httpRng.contentRange(contentLength))
	respHeaders.Set("Content-Length", fmt.Sprintf("%d", httpRng.length))

	p.finishRequest(206, w, r, fileReader)
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

	defer fileReader.Close()

	respHeaders := w.Header()
	for headerName, headerValue := range fileHeaders {
		if headerName == "Content-Length" || headerName == "Content-Range" {
			continue
		}
		respHeaders.Set(headerName, strings.Join(headerValue, ","))
	}

	p.finishRequest(200, w, r, fileReader)
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
		req.Header.Set(headerName, strings.Join(headerValue, ","))
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
		respHeaders.Set(headerName, strings.Join(headerValue, ","))
	}

	p.finishRequest(resp.StatusCode, w, r, resp.Body)
}

func (ph *proxyHandler) finishRequest(statusCode int, w http.ResponseWriter,
	r *http.Request, reader io.Reader) {

	rng := r.Header.Get("Range")
	if rng == "" {
		rng = "-"
	}

	log.Printf("[%p] %d %s %s", r, statusCode, rng, r.RequestURI)

	w.WriteHeader(statusCode)
	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("[%p] io.Copy - %s. r.ConLen: %d", r, err, r.ContentLength)
	}
}

func newProxyHandler(app *Application) *proxyHandler {

	return &proxyHandler{
		app:    app,
		config: app.cfg.HTTP,
	}

}
