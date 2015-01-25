package app

import (
	"html/template"
	"log"
	"net/http"
)

//!TODO: Do not parse the template every request
func (ph *proxyHandler) StatusPage(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/status_page.html")

	if err != nil {
		log.Printf("Error parsing template file: %s", err)
		w.Write([]byte(err.Error()))
		return
	}

	log.Printf("[%p] 200 Status page\n", r)
	w.WriteHeader(200)

	if err := tmpl.Execute(w, ph.app.cacheManagers); err != nil {
		w.Write([]byte(err.Error()))
	}

	return
}
