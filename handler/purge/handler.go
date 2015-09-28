package purge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/context"

	"github.com/ironsmile/nedomi/config"
	"github.com/ironsmile/nedomi/contexts"
	"github.com/ironsmile/nedomi/types"
)

// Handler is a simple handler that handles the server purge page.
type Handler struct {
	logger types.Logger
}

const (
	globPattern = "*"
)

var defaultSettings = serverPurgeHandlerSettings{}

type serverPurgeHandlerSettings struct {
}

type purgeRequest struct {
	CacheZoneName string   `json:"cache_zone"`
	CacheZoneKey  string   `json:"cache_zone_key"`
	Objects       []string `json:"objects"`
}

type purgeResult struct {
	CacheZoneName string          `json:"cache_zone"`
	CacheZoneKey  string          `json:"cache_zone_key"`
	Results       map[string]bool `json:"results"`
}

// RequestHandle servers the purge page.
func (ph *Handler) RequestHandle(ctx context.Context,
	w http.ResponseWriter, r *http.Request, l *types.Location) {
	//!TODO authentication
	if r.Method != "POST" {
		http.Error(w, "Wrong method", http.StatusBadRequest)
	}

	var pr purgeRequest
	if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		ph.logger.Errorf("[%p] error on parsing request %s", ph, err)
		return
	}

	czs, _ := contexts.GetCacheZones(ctx)
	for name, zone := range czs {
		if name == pr.CacheZoneName {
			res := ph.purgeAll(zone, pr)
			w.WriteHeader(http.StatusOK)
			encoder := json.NewEncoder(w)
			err := encoder.Encode(res)
			if err != nil {
				ph.logger.Errorf("[%p] error while encodinf respose %s", ph, err)
			}
		}
	}
	return
}

func (ph *Handler) purgeAll(zone types.CacheZone, pr purgeRequest) purgeResult {
	var pres = purgeResult{
		CacheZoneName: pr.CacheZoneName,
		CacheZoneKey:  pr.CacheZoneKey,
		Results:       make(map[string]bool),
	}

	for _, object := range pr.Objects {
		if strings.HasSuffix(object, globPattern) {
			pres.Results[object] =
				zone.Algorithm.RemoveObjectsForKey(
					pr.CacheZoneKey,
					strings.TrimSuffix(object, globPattern))
		} else {
			pres.Results[object] =
				zone.Algorithm.RemoveObject(
					types.NewObjectID(pr.CacheZoneKey, object))
		}
	}
	return pres
}

// New creates and returns a ready to used ServerPurgeHandler.
func New(cfg *config.Handler, l *types.Location, next types.RequestHandler) (*Handler, error) {
	var s = defaultSettings
	if len(cfg.Settings) > 0 {
		if err := json.Unmarshal(cfg.Settings, &s); err != nil {
			return nil, fmt.Errorf("error while parsing settings for handler.purge - %s", err)
		}
	}

	return &Handler{
		logger: l.Logger,
	}, nil
}