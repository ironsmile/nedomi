package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/vhost"
)

// Used to stop following redirects with the default http.Client
var ErrNoRedirects = fmt.Errorf("No redirects")

// Headers in this map will be skipped in the response
var skippedHeaders = map[string]bool{
	"Transfer-Encoding": true,
	"Content-Range":     true,
}

// ProxyHandler is the type resposible for implementing the RequestHandler interface
// in this proxy module.
type ProxyHandler struct {
}

func shouldSkipHeader(header string) bool {
	return skippedHeaders[header]
}

//!TODO: Add something more than a GET requests
//!TODO: Rewrite Date header

// RequestHandle is the main serving function
func (ph *ProxyHandler) RequestHandle(writer http.ResponseWriter,
	req *http.Request, vh *vhost.VirtualHost) {

	log.Printf("[%p] Access %s", req, req.RequestURI)

	rng := req.Header.Get("Range")

	if rng != "" {
		ph.ServerPartialRequest(writer, req, vh)
		return
	} else {
		ph.ServeFullRequest(writer, req, vh)
		return
	}

}

// ServerPartialRequest handles serving client requests that have a specified range.
func (p *ProxyHandler) ServerPartialRequest(w http.ResponseWriter, r *http.Request,
	vh *vhost.VirtualHost) {
	objID := types.ObjectID{CacheKey: vh.CacheKey, Path: r.URL.String()}

	fileHeaders, err := vh.Storage.Headers(&vh.VirtualHost, objID)

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

	fileReader, err := vh.Storage.Get(&vh.VirtualHost, objID, uint64(httpRng.start),
		uint64(httpRng.start+httpRng.length-1))

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file reader error. %s\n", r, err)
		return
	}

	defer fileReader.Close()

	respHeaders := w.Header()

	for headerName, headerValue := range fileHeaders {
		if shouldSkipHeader(headerName) {
			continue
		}
		respHeaders.Set(headerName, strings.Join(headerValue, ","))
	}

	respHeaders.Set("Content-Range", httpRng.contentRange(contentLength))
	respHeaders.Set("Content-Length", fmt.Sprintf("%d", httpRng.length))

	p.finishRequest(206, w, r, fileReader)
}

// ServeFullRequest handles serving client requests that request the whole file.
func (p *ProxyHandler) ServeFullRequest(w http.ResponseWriter, r *http.Request,
	vh *vhost.VirtualHost) {
	objID := types.ObjectID{CacheKey: vh.CacheKey, Path: r.URL.String()}

	fileHeaders, err := vh.Storage.Headers(&vh.VirtualHost, objID)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file headers. %s\n", r, err)
		return
	}

	fileReader, err := vh.Storage.GetFullFile(&vh.VirtualHost, objID)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file reader error. %s\n", r, err)
		return
	}

	defer fileReader.Close()

	respHeaders := w.Header()
	for headerName, headerValue := range fileHeaders {
		if shouldSkipHeader(headerName) {
			continue
		}
		respHeaders.Set(headerName, strings.Join(headerValue, ","))
	}

	p.finishRequest(200, w, r, fileReader)
}

// ProxyRequest does not use the local storage and directly proxies the
// request to the upstream server.
func (p *ProxyHandler) ProxyRequest(w http.ResponseWriter, r *http.Request,
	vh *vhost.VirtualHost) {
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

func (ph *ProxyHandler) finishRequest(statusCode int, w http.ResponseWriter,
	r *http.Request, responseContents io.Reader) {

	rng := r.Header.Get("Range")
	if rng == "" {
		rng = "-"
	}

	log.Printf("[%p] %d %s %s", r, statusCode, rng, r.RequestURI)

	w.WriteHeader(statusCode)
	if _, err := io.Copy(w, responseContents); err != nil {
		log.Printf("[%p] io.Copy - %s. r.ConLen: %d", r, err, r.ContentLength)
	}
}

// New creates and returns a ready to used ProxyHandler.
func New() *ProxyHandler {
	return &ProxyHandler{}
}
