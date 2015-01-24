package app

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gophergala/nedomi/config"
)

func listen(ht config.HTTPSection) error {

	proxyHandler := newProxyHandler(ht)
	return http.ListenAndServe(ht.Listen, proxyHandler)
}

type proxyHandler struct {
	config config.HTTPSection
}

func (ph *proxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	spew.Dump(r.URL)
	spew.Printf("URL.Path : %s \n", r.URL.Path)
	spew.Printf("URL.Opaque : %s \n", r.URL.Opaque)
	spew.Printf("RequestURI : %s \n", r.RequestURI)
	spew.Printf("REFERrER : %s \n", r.Referer())
	var serverConfig config.ServerSection
	// optimize
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
	serverConfig = ph.config.Servers[0]
	log.Printf("found server config %s", serverConfig)
	client := http.Client{}
	isHTTPS := r.TLS != nil
	schema := "http"
	if isHTTPS {
		schema = "https"
	}
	var newUrl *url.URL
	newUrl, err := url.Parse(schema + "://" + serverConfig.UpstreamAddress)
	if err != nil {
		log.Println("Error while constructing url to proxy to", err)
	}
	newUrl.Path = r.RequestURI
	spew.Dump(newUrl)

	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		log.Printf("Got error\n %s\n while proxying %s to %s", err, r.URL.String(), newUrl.String())
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Got error\n %s\n while proxying %s to %s", err, r.URL.String(), newUrl.String())
		return
	}

	defer resp.Body.Close()
	spew.Dump(resp.Header)
	spew.Dump(resp.Trailer)
	io.Copy(w, resp.Body)
}

func newProxyHandler(ht config.HTTPSection) *proxyHandler {

	return &proxyHandler{
		config: ht,
	}

}
