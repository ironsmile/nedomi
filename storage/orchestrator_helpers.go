package storage

import "github.com/ironsmile/nedomi/types"

type storageItem struct {
	Obj   *types.ObjectMetadata
	Parts types.ObjectIndexMap
}

func (o *Orchestrator) startConcurrentIterator() {
	callback := func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {
		select {
		case <-o.doneCh:
			return false
		case o.foundCh <- &storageItem{obj, parts}:
			//!TODO: add throttling here? a simple time.Sleep() should do it :)
			return true
		}
	}

	go func() {
		defer close(o.foundCh) //!TODO: determine if we should close the channel here or somewhere else
		if err := o.storage.Iterate(callback); err != nil {
			o.logger.Errorf("Received iterator error: %s", err)
		}
	}()
}
