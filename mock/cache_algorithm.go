package mock

import "github.com/ironsmile/nedomi/types"

// CacheAlgorithmRepliers are used for setting fake reply for a particular index.
type CacheAlgorithmRepliers struct {
	Lookup        func(*types.ObjectIndex) bool
	ShouldKeep    func(*types.ObjectIndex) bool
	AddObject     func(*types.ObjectIndex) error
	Remove        func(...*types.ObjectIndex)
	PromoteObject func(*types.ObjectIndex)
}

// DefaultCacheAlgorithmRepliers always return false and nil
var DefaultCacheAlgorithmRepliers = CacheAlgorithmRepliers{
	Lookup:        func(*types.ObjectIndex) bool { return false },
	ShouldKeep:    func(*types.ObjectIndex) bool { return false },
	AddObject:     func(*types.ObjectIndex) error { return nil },
	PromoteObject: func(*types.ObjectIndex) {},
	Remove:        func(...*types.ObjectIndex) {},
}

// CacheAlgorithm is used in different tests as a cache algorithm substitute
type CacheAlgorithm struct {
	Defaults CacheAlgorithmRepliers
	Mapping  map[types.ObjectIndex]*CacheAlgorithmRepliers
}

// Remove removes the cpecified objects from the cache. Currently only the
// default MockRepliers are being used to implement the call
func (c *CacheAlgorithm) Remove(os ...*types.ObjectIndex) {
	c.Defaults.Remove(os...)
}

// Lookup returns the specified (if present for this index) or default value
func (c *CacheAlgorithm) Lookup(o *types.ObjectIndex) bool {
	if found, ok := c.Mapping[*o]; ok && found.Lookup != nil {
		return found.Lookup(o)
	}
	return c.Defaults.Lookup(o)
}

// ShouldKeep returns the specified (if present for this index) or default value
func (c *CacheAlgorithm) ShouldKeep(o *types.ObjectIndex) bool {
	if found, ok := c.Mapping[*o]; ok && found.ShouldKeep != nil {
		return found.ShouldKeep(o)
	}
	return c.Defaults.ShouldKeep(o)
}

// AddObject returns the specified (if present for this index) or default error
func (c *CacheAlgorithm) AddObject(o *types.ObjectIndex) error {
	if found, ok := c.Mapping[*o]; ok && found.AddObject != nil {
		return found.AddObject(o)
	}
	return c.Defaults.AddObject(o)
}

// PromoteObject calls the specified (if present for this index) or default callback
func (c *CacheAlgorithm) PromoteObject(o *types.ObjectIndex) {
	if found, ok := c.Mapping[*o]; ok && found.PromoteObject != nil {
		found.PromoteObject(o)
		return
	}
	c.Defaults.PromoteObject(o)
}

// ConsumedSize always returns 0
func (c *CacheAlgorithm) ConsumedSize() types.BytesSize {
	return 0
}

// Stats always returns nil
func (c *CacheAlgorithm) Stats() types.CacheStats {
	return nil
}

// Resize does nothing
func (c *CacheAlgorithm) Resize(_ uint64) {
}

// SetFakeReplies is used to customize the replies for certain indexes
func (c *CacheAlgorithm) SetFakeReplies(index *types.ObjectIndex, replies *CacheAlgorithmRepliers) {
	c.Mapping[*index] = replies
}

// NewCacheAlgorithm creates and returns a new mock cache algorithm. The default replies
// can be specified by the `defaults` argument.
func NewCacheAlgorithm(defaults *CacheAlgorithmRepliers) *CacheAlgorithm {
	res := &CacheAlgorithm{
		Defaults: DefaultCacheAlgorithmRepliers,
		Mapping:  make(map[types.ObjectIndex]*CacheAlgorithmRepliers),
	}
	if defaults == nil {
		return res
	}

	if defaults.Lookup != nil {
		res.Defaults.Lookup = defaults.Lookup
	}
	if defaults.ShouldKeep != nil {
		res.Defaults.ShouldKeep = defaults.ShouldKeep
	}
	if defaults.AddObject != nil {
		res.Defaults.AddObject = defaults.AddObject
	}
	if defaults.PromoteObject != nil {
		res.Defaults.PromoteObject = defaults.PromoteObject
	}

	return res
}
