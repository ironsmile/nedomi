package cache

import (
	"errors"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

var idx = &types.ObjectIndex{
	ObjID: types.NewObjectID("test", "/path/to/object"),
	Part:  14,
}

func TestMockCacheAlgorithm(t *testing.T) {
	t.Parallel()
	d := NewMock(nil)
	if d.Defaults.Lookup(idx) || d.Defaults.ShouldKeep(idx) || d.Defaults.AddObject(idx) != nil {
		t.Errorf("Invalid default default replies %#v", d.Defaults)
	}

	c1 := NewMock(&MockRepliers{
		Lookup: func(*types.ObjectIndex) bool { return true },
	})
	if !c1.Defaults.Lookup(idx) || c1.Defaults.ShouldKeep(idx) || c1.Defaults.AddObject(idx) != nil {
		t.Error("Invalid partial custom default replies")
	}

	c2 := NewMock(&MockRepliers{
		Lookup:     func(*types.ObjectIndex) bool { return true },
		ShouldKeep: func(*types.ObjectIndex) bool { return true },
		AddObject:  func(*types.ObjectIndex) error { return errors.New("ha") },
	})
	if !c2.Defaults.Lookup(idx) || !c2.Defaults.ShouldKeep(idx) || c2.Defaults.AddObject(idx) == nil {
		t.Error("Invalid full custom default replies")
	}
}

func TestSettingFakeReplies(t *testing.T) {
	t.Parallel()
	ca := NewMock(&MockRepliers{
		Lookup:    func(*types.ObjectIndex) bool { return true },
		AddObject: func(*types.ObjectIndex) error { return errors.New("pa") },
	})
	if !ca.Lookup(idx) || ca.ShouldKeep(idx) || ca.AddObject(idx) == nil {
		t.Error("Unexpected mock replies")
	}

	fakeReplies := &MockRepliers{
		Lookup:     func(*types.ObjectIndex) bool { return false },
		ShouldKeep: func(*types.ObjectIndex) bool { return true },
	}
	ca.SetFakeReplies(idx, fakeReplies)
	if ca.Lookup(idx) || !ca.ShouldKeep(idx) || ca.AddObject(idx) == nil {
		t.Error("Unexpected mock replies after setting the fakes")
	}
}
