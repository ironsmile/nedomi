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

	promoted := false
	c2 := NewMock(&MockRepliers{
		Lookup:        func(*types.ObjectIndex) bool { return true },
		ShouldKeep:    func(*types.ObjectIndex) bool { return true },
		AddObject:     func(*types.ObjectIndex) error { return errors.New("ha") },
		PromoteObject: func(*types.ObjectIndex) { promoted = true },
	})
	if !c2.Defaults.Lookup(idx) || !c2.Defaults.ShouldKeep(idx) || c2.Defaults.AddObject(idx) == nil {
		t.Error("Invalid full custom default replies")
	}
	if c2.PromoteObject(idx); !promoted {
		t.Error("Object was not promoted")
	}
}

func TestSettingFakeReplies(t *testing.T) {
	t.Parallel()
	var promotedByDefault, promotedByCustom bool
	ca := NewMock(&MockRepliers{
		Lookup:        func(*types.ObjectIndex) bool { return true },
		AddObject:     func(*types.ObjectIndex) error { return errors.New("pa") },
		PromoteObject: func(*types.ObjectIndex) { promotedByDefault = true },
	})
	if !ca.Lookup(idx) || ca.ShouldKeep(idx) || ca.AddObject(idx) == nil {
		t.Error("Unexpected mock replies")
	}
	if ca.PromoteObject(idx); !promotedByDefault {
		t.Error("Object was not promoted")
	}

	fakeReplies := &MockRepliers{
		Lookup:     func(*types.ObjectIndex) bool { return false },
		ShouldKeep: func(*types.ObjectIndex) bool { return true },
		PromoteObject: func(*types.ObjectIndex) {
			promotedByDefault = false
			promotedByCustom = true
		},
	}
	ca.SetFakeReplies(idx, fakeReplies)
	if ca.Lookup(idx) || !ca.ShouldKeep(idx) || ca.AddObject(idx) == nil {
		t.Error("Unexpected mock replies after setting the fakes")
	}
	if ca.PromoteObject(idx); promotedByDefault || !promotedByCustom {
		t.Error("Object was not promoted correctly")
	}
}
