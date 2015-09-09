package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils"
)

//!TODO: some unit tests? :)

// Handler is the type resposible for implementing the RequestHandler interface
// in this proxy module.
type Handler struct {
}

//!TODO: Rewrite Date header

// RequestHandle is the main serving function
func (ph *Handler) RequestHandle(ctx context.Context,
	resp http.ResponseWriter, req *http.Request, l *types.Location) {

	if utils.IsRequestCacheable(req) {
		l.Logger.Logf("[%p] Cacheable access: %s", req, req.RequestURI)
		ph.handleCacheableRequest(ctx, resp, req, l)

	} else {
		l.Logger.Logf("[%p] Direct proxy access: %s", req, req.RequestURI)
		l.Upstream.ServeHTTP(resp, req)
	}
}

func (ph *Handler) handleCacheableRequest(ctx context.Context,
	resp http.ResponseWriter, req *http.Request, l *types.Location) {

	obj, reader, err := l.Orchestrator.Handle(ctx, req)
	if err != nil {

	}
	defer reader.Close()

	for h := range obj.Headers {
		resp.Header().Set(h, obj.Headers.Get(h))
	}
	resp.WriteHeader(obj.Code)
	if copied, err := io.Copy(resp, reader); err != nil {
		l.Logger.Logf("[%p] Error copying response: %s. Copied %d from %d", req, err, copied, obj.Size)
	}

}

//const fullContentRange = "*/*"
/*
// Used to stop following redirects with the default http.Client
var ErrNoRedirects = fmt.Errorf("No redirects")

// Headers in this map will be skipped in the response
var skippedHeaders = map[string]bool{
	"Transfer-Encoding": true,
	"Content-Range":     true,
}

func shouldSkipHeader(header string) bool {
	return skippedHeaders[header]
}

// servePartialRequest handles serving client requests that have a specified range.
func (ph *Handler) servePartialRequest(ctx context.Context,
	w http.ResponseWriter, r *http.Request, vh *types.VirtualHost) {

			objID := types.ObjectID{CacheKey: vh.CacheKey, Path: r.URL.String()}
			objMetadata := ph.getMetadata(objID)

			fileHeaders, err := vh.Storage.Headers(ctx, objID)

			if err != nil {
				http.Error(w, fmt.Sprintf("%s", err), 500)
				log.Printf("[%p] Getting file headers. %s\n", r, err)
				return
			}

			cl := fileHeaders.Get("Content-Length")
			contentLength, err := strconv.ParseInt(cl, 10, 64)

			if err != nil {
				w.Header().Set("Content-Range", fullContentRange)
				msg := fmt.Sprintf("File content-length was not parsed: %s. %s", cl, err)
				log.Printf("[%p] %s", r, msg)
				http.Error(w, msg, 416)
				return
			}

			ranges, err := parseRange(r.Header.Get("Range"), contentLength)

			if err != nil {
				w.Header().Set("Content-Range", fullContentRange)
				msg := fmt.Sprintf("Bytes range error: %s. %s", r.Header.Get("Range"), err)
				log.Printf("[%p] %s", r, msg)
				http.Error(w, msg, 416)
				return
			}

			if len(ranges) != 1 {
				w.Header().Set("Content-Range", fullContentRange)
				msg := fmt.Sprintf("We support only one set of bytes ranges. Got %d", len(ranges))
				log.Printf("[%p] %s", r, msg)
				http.Error(w, msg, 416)
				return
			}

			httpRng := ranges[0]

			fileReader, err := vh.Storage.Get(ctx, objID, uint64(httpRng.start),
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

		ph.finishRequest(206, w, r, fileReader)
}

// serveFullRequest handles serving client requests that request the whole file.
func (ph *Handler) serveFullRequest(ctx context.Context,
	w http.ResponseWriter, r *http.Request, vh *types.VirtualHost) {

	objID := types.ObjectID{CacheKey: vh.CacheKey, Path: r.URL.String()}

	fileHeaders, err := vh.Storage.Headers(ctx, objID)

	if err != nil {
		http.Error(w, fmt.Sprintf("%s", err), 500)
		log.Printf("[%p] Getting file headers. %s\n", r, err)
		return
	}

	fileReader, err := vh.Storage.GetFullFile(ctx, objID)

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

	ph.finishRequest(200, w, r, fileReader)
}


func (ph *Handler) finishRequest(statusCode int, w http.ResponseWriter,
	r *http.Request, vh *types.VirtualHost, responseContents io.Reader) {

	rng := r.Header.Get("Range")
	if rng == "" {
		rng = "-"
	}

	vh.Logger.Logf("[%p] %d %s %s", r, statusCode, rng, r.RequestURI)

	w.WriteHeader(statusCode)
	if _, err := io.Copy(w, responseContents); err != nil {
		vh.Logger.Logf("[%p] io.Copy - %s. r.ConLen: %d", r, err, r.ContentLength)
	}
}
*/

// New creates and returns a ready to used Handler.
func New(cfg json.RawMessage, l *types.Location) (*Handler, error) {
	if l.Upstream == nil {
		return nil, fmt.Errorf("proxy handler requires upstream")
	}

	if l.Orchestrator == nil {
		return nil, fmt.Errorf("proxy handler requires orchestrator") // !TODO: this mistake should be more informative
	}
	//!TODO: parse the cfg
	return &Handler{}, nil
}
