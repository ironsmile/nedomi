package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gophergala/nedomi/config"
)

func listen(ht config.HTTPSection) error {

	proxyHandler := newProxyHandler(ht)
	return http.ListenAndServe(ht.Listen, proxyHandler)
}

type proxyHandler struct {
	config config.HTTPSection
}

var ErrNoRedirects = fmt.Errorf("No redirects")

func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var serverConfig config.ServerSection

	//!TODO: optimize
	split := strings.Split(r.Host, ":")
	for _, server := range ph.config.Servers {
		if server.Name == split[0] {
			serverConfig = server
			break
		}
	}

	if serverConfig.Name == "" {
		http.NotFound(w, r)
		return
	}

	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return ErrNoRedirects
	}

	var newUrl *url.URL
	newUrl, err := url.Parse(serverConfig.UpstreamAddress)
	if err != nil {
		log.Println("Error while constructing url to proxy to", err)
	}
	newUrl = newUrl.ResolveReference(r.URL)

	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		log.Printf("Got error\n %s\n while making request ", err)
		return
	}

	for headerName, headerValue := range r.Header {
		req.Header.Add(headerName, strings.Join(headerValue, ","))
	}

	log.Printf("%s", req)

	resp, err := client.Do(req)
	if err != nil && err != ErrNoRedirects {
		if urlError, ok := err.(*url.Error); !(ok && urlError.Err == ErrNoRedirects) {
			log.Printf("Got error\n %s\n while proxying %s to %s", err, r.URL.String(), newUrl.String())
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

func newProxyHandler(ht config.HTTPSection) *proxyHandler {

	return &proxyHandler{
		config: ht,
	}

}
