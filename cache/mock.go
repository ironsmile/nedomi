package cache

import "github.com/ironsmile/nedomi/types"

// MockReplies are used for setting fake reply for a particular index.
type MockReplies struct {
	Lookup     bool
	ShouldKeep bool
	AddObject  error
}

// MockCacheAlgorithm is used in different tests as a cache algorithm substitute
type MockCacheAlgorithm struct {
	Defaults MockReplies
	Mapping  map[types.ObjectIndex]MockReplies
}

// Lookup returns the specified (if present for this index) or default value
func (c *MockCacheAlgorithm) Lookup(o *types.ObjectIndex) bool {
	if found, ok := c.Mapping[*o]; ok {
		return found.Lookup
	}
	return c.Defaults.Lookup
}

// ShouldKeep returns the specified (if present for this index) or default value
func (c *MockCacheAlgorithm) ShouldKeep(o *types.ObjectIndex) bool {
	if found, ok := c.Mapping[*o]; ok {
		return found.ShouldKeep
	}
	return c.Defaults.ShouldKeep
}

// AddObject returns the specified (if present for this index) or default error
func (c *MockCacheAlgorithm) AddObject(o *types.ObjectIndex) error {
	if found, ok := c.Mapping[*o]; ok {
		return found.AddObject
	}
	return c.Defaults.AddObject
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
func (c *MockCacheAlgorithm) SetFakeReplies(index *types.ObjectIndex, replies MockReplies) {
	c.Mapping[*index] = replies
}

// NewMock creates and returns a new mock cache algorithm. The default replies
// can be specified by the `defaults` argument.
func NewMock(defaults *MockReplies) *MockCacheAlgorithm {
	res := &MockCacheAlgorithm{
		Mapping: make(map[types.ObjectIndex]MockReplies),
	}
	if defaults != nil {
		res.Defaults = *defaults
	}

	return res
}
