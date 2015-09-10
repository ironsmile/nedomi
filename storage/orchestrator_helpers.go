package storage

import "github.com/ironsmile/nedomi/types"

func isMetadataFresh(obj *types.ObjectMetadata) bool {
	//!TODO: implement
	return true
}

type storageItem struct {
	Obj   *types.ObjectMetadata
	Parts types.ObjectIndexMap
}

func (o *Orchestrator) startConcurrentIterator() {
	counter := 0
	callback := func(obj *types.ObjectMetadata, parts types.ObjectIndexMap) bool {

		//!TODO: implement proper stop
		//select {
		//case <-o.done:
		//	return false
		//}
		counter++
		for n := range parts {
			o.algorithm.AddObject(&types.ObjectIndex{ObjID: obj.ID, Part: n})
		}

		//!TODO: add throttling here? a simple time.Sleep() should do it :)
		return true
	}

	go func() {
		if err := o.storage.Iterate(callback); err != nil {
			o.logger.Errorf("Received iterator error '%s' after loading %d objects", err, counter)
		} else {
			o.logger.Logf("Loading contents from disk finished: %d objects loaded!", counter)
		}
	}()
}
