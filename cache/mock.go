package cache

import "github.com/ironsmile/nedomi/types"

// MockRepliers are used for setting fake reply for a particular index.
type MockRepliers struct {
	Lookup     func(*types.ObjectIndex) bool
	ShouldKeep func(*types.ObjectIndex) bool
	AddObject  func(*types.ObjectIndex) error
}

// DefaultMockRepliers always return false and nil
var DefaultMockRepliers = &MockRepliers{
	Lookup:     func(*types.ObjectIndex) bool { return false },
	ShouldKeep: func(*types.ObjectIndex) bool { return false },
	AddObject:  func(*types.ObjectIndex) error { return nil },
}

// MockCacheAlgorithm is used in different tests as a cache algorithm substitute
type MockCacheAlgorithm struct {
	Defaults *MockRepliers
	Mapping  map[types.ObjectIndex]*MockRepliers
}

// Remove removes the cpecified object from the cache and returns true
// if it was in the cache. This implementation actually is synonim for Lookup
func (c *MockCacheAlgorithm) Remove(o *types.ObjectIndex) bool {
	return c.Lookup(o)
}

// Lookup returns the specified (if present for this index) or default value
func (c *MockCacheAlgorithm) Lookup(o *types.ObjectIndex) bool {
	if found, ok := c.Mapping[*o]; ok && found.Lookup != nil {
		return found.Lookup(o)
	}
	return c.Defaults.Lookup(o)
}

// ShouldKeep returns the specified (if present for this index) or default value
func (c *MockCacheAlgorithm) ShouldKeep(o *types.ObjectIndex) bool {
	if found, ok := c.Mapping[*o]; ok && found.ShouldKeep != nil {
		return found.ShouldKeep(o)
	}
	return c.Defaults.ShouldKeep(o)
}

// AddObject returns the specified (if present for this index) or default error
func (c *MockCacheAlgorithm) AddObject(o *types.ObjectIndex) error {
	if found, ok := c.Mapping[*o]; ok && found.AddObject != nil {
		return found.AddObject(o)
	}
	return c.Defaults.AddObject(o)
}

// PromoteObject does nothing
func (c *MockCacheAlgorithm) PromoteObject(o *types.ObjectIndex) {}

// ConsumedSize always returns 0
func (c *MockCacheAlgorithm) ConsumedSize() types.BytesSize {
	return 0
}

// Stats always returns nil
func (c *MockCacheAlgorithm) Stats() types.CacheStats {
	return nil
}

// SetFakeReplies is used to customize the replies for certain indexes
func (c *MockCacheAlgorithm) SetFakeReplies(index *types.ObjectIndex, replies *MockRepliers) {
	c.Mapping[*index] = replies
}

// NewMock creates and returns a new mock cache algorithm. The default replies
// can be specified by the `defaults` argument.
func NewMock(defaults *MockRepliers) *MockCacheAlgorithm {
	res := &MockCacheAlgorithm{
		Defaults: DefaultMockRepliers,
		Mapping:  make(map[types.ObjectIndex]*MockRepliers),
	}
	if defaults == nil {
		return res
	}

	res.Defaults = defaults
	if defaults.Lookup == nil {
		res.Defaults.Lookup = DefaultMockRepliers.Lookup
	}
	if defaults.ShouldKeep == nil {
		res.Defaults.ShouldKeep = DefaultMockRepliers.ShouldKeep
	}
	if defaults.AddObject == nil {
		res.Defaults.AddObject = DefaultMockRepliers.AddObject
	}

	return res
}
