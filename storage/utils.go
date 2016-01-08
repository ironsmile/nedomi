package storage

import "github.com/ironsmile/nedomi/types"

// GetExpirationHandler returns a potentially long-lived callback that removes
// the specified object from the storage.
func GetExpirationHandler(cz *types.CacheZone, id *types.ObjectID) func(types.Logger) {
	return func(logger types.Logger) {
		//!TODO: simplify and ignore the cache algorithm when expiring objects.
		// It is only supposed to take into account client interest in the
		// object parts, not whether they are expired due to upstream timeouts
		parts, err := cz.Storage.GetAvailableParts(id)
		if err != nil {
			logger.Errorf("Error while removing expired object %s from zone %s: %s", id, cz.ID, err)
		}

		cz.Algorithm.Remove(parts...)

		//!TODO: make head request to upstream and possibly postpone the
		// removal, if nothing has changed in the file
		if err := cz.Storage.Discard(id); err != nil {
			logger.Errorf("Error while discarding expired object %s from zone %s: %s", id, cz.ID, err)
		}
	}
}
