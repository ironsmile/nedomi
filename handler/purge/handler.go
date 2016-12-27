package purge

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
	"github.com/ironsmile/nedomi/utils/httputils"
)

// Handler is a simple handler that handles the server purge page.
type Handler struct {
	logger types.Logger
}

type purgeRequest config.StringSlice
type purgeResult map[string]bool

// ServeHTTP servers the purge page.
func (ph *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID, _ := contexts.GetRequestID(r.Context())
	//!TODO authentication
	if r.Method != "POST" {
		httputils.Error(w, http.StatusMethodNotAllowed)
		return
	}

	var pr = new(purgeRequest)
	if err := json.NewDecoder(r.Body).Decode(pr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		ph.logger.Errorf("[%s] error on parsing request %s",
			reqID, err)
		return
	}

	var app, ok = contexts.GetApp(r.Context())
	if !ok {
		httputils.Error(w, http.StatusInternalServerError)
		ph.logger.Errorf("[%s] no app in context", reqID)
		return
	}
	var res, err = ph.purgeAll(reqID, app, *pr)
	if err != nil {
		httputils.Error(w, http.StatusInternalServerError)
		// previosly logged
		return
	}
	if err := json.NewEncoder(w).Encode(res); err != nil {
		ph.logger.Errorf("[%s] error while encoding response %s",
			reqID, err)
	}
}

func (ph *Handler) purgeAll(reqID types.RequestID, app types.App, pr purgeRequest) (purgeResult, error) {
	var pres = purgeResult(make(map[string]bool))

	for _, uString := range pr {
		pres[uString] = false
		var u, err = url.Parse(uString)
		if err != nil {
			continue
		}
		var location = app.GetLocationFor(u.Host, u.Path)
		if location == nil {
			ph.logger.Logf(
				"[%s] got request to purge an object (%s) that is for a not configured location",
				reqID, uString)
			continue
		}

		var oid = location.NewObjectIDForURL(u)

		parts, err := location.Cache.Storage.GetAvailableParts(oid)

		if err != nil {
			if !os.IsNotExist(err) {
				ph.logger.Errorf(
					"[%s] got error while gettings parts of object '%s' - %s",
					reqID, oid, err)
				return nil, err
			}
		}

		if len(parts) == 0 {
			continue
		}

		if err = location.Cache.Storage.Discard(oid); err != nil {
			if !os.IsNotExist(err) {
				ph.logger.Errorf(
					"[%s] got error while purging object '%s' - %s",
					reqID, oid, err)
				return nil, err
			}
		}

		location.Cache.Algorithm.Remove(parts...)
		pres[uString] = err == nil // err is os.ErrNotExist
	}
	return pres, nil
}

// New creates and returns a ready to used ServerPurgeHandler.
func New(cfg *config.Handler, l *types.Location, next http.Handler) (*Handler, error) {
	return &Handler{
		logger: l.Logger,
	}, nil
}
