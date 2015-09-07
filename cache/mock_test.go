package cache

import (
	"errors"
	"testing"

	"github.com/ironsmile/nedomi/types"
)

func TestMockCacheAlgorithm(t *testing.T) {
	d := NewMock(nil)
	if d.Defaults.Lookup || d.Defaults.ShouldKeep || d.Defaults.AddObject != nil {
		t.Errorf("Invalid default default replies %#v", d.Defaults)
	}

	customReplies := &MockReplies{Lookup: true, ShouldKeep: true, AddObject: errors.New("ha")}
	c := NewMock(customReplies)
	if !c.Defaults.Lookup || !c.Defaults.ShouldKeep || c.Defaults.AddObject == nil {
		t.Errorf("Invalid custom default replies %#v", c.Defaults)
	}
}

func TestSettingFakeReplies(t *testing.T) {
	defaultReplies := &MockReplies{Lookup: true, ShouldKeep: false, AddObject: nil}
	ca := NewMock(defaultReplies)
	idx := &types.ObjectIndex{
		ObjID: types.NewObjectID("test", "/path/to/object"),
		Part:  14,
	}

	if !ca.Lookup(idx) || ca.ShouldKeep(idx) || ca.AddObject(idx) != nil {
		t.Error("Unexpected mock replies")
	}

	fakeReplies := MockReplies{Lookup: false, ShouldKeep: true, AddObject: errors.New("pa")}
	ca.SetFakeReplies(idx, fakeReplies)
	if ca.Lookup(idx) || !ca.ShouldKeep(idx) || ca.AddObject(idx) == nil {
		t.Error("Unexpected mock replies after setting the fakes")
	}
}
