package storage

import "github.com/ironsmile/nedomi/types"

type storageItem struct {
	Obj   *types.ObjectMetadata
	Parts types.ObjectIndexMap
}

func (o *Orchestrator) startConcurrentIterator() {
	counter := 0
	callback := func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {
		select {
		case <-o.done:
			return false
		case o.foundObjects <- &storageItem{obj, parts}:
			counter++
			//!TODO: add throttling here? a simple time.Sleep() should do it :)
			return true
		}
	}

	go func() {
		if err := o.storage.Iterate(callback); err != nil {
			o.logger.Errorf("Received iterator error '%s' after loading %d objects", err, counter)
		} else {
			o.logger.Logf("Loading contents from disk finished: %d objects loaded!", counter)
		}
	}()
}
