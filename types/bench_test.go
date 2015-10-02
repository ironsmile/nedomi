package types

import "testing"

const count = 10000

var (
	id     = NewObjectID("/path/to/somewhere", "key")
	first  = &ObjectIndex{ObjID: id, Part: 0}
	middle = &ObjectIndex{ObjID: id, Part: count / 2}
	last   = &ObjectIndex{ObjID: id, Part: count - 1}
)

func buildMapHash(id *ObjectID, max uint32) map[ObjectIndexHash]*ObjectIndex {
	var m = make(map[ObjectIndexHash]*ObjectIndex)
	for i := uint32(0); max > i; i++ {
		index := &ObjectIndex{
			ObjID: id,
			Part:  i,
		}
		m[index.Hash()] = index
	}

	return m
}

func buildMapString(id *ObjectID, max uint32) map[string]*ObjectIndex {
	var m = make(map[string]*ObjectIndex)
	for i := uint32(0); max > i; i++ {
		index := &ObjectIndex{
			ObjID: id,
			Part:  i,
		}
		m[index.HashStr()] = index
	}

	return m
}

func BenchmarkHash(b *testing.B) {
	var m = buildMapHash(id, count)
	for i := 0; b.N > i; i++ {
		if _, ok := m[first.Hash()]; !ok {
			b.Fail()
		}
		if _, ok := m[middle.Hash()]; !ok {
			b.Fail()
		}
		if _, ok := m[last.Hash()]; !ok {
			b.Fail()
		}
	}

}
func BenchmarkHashStr(b *testing.B) {
	var m = buildMapString(id, count)
	for i := 0; b.N > i; i++ {
		if _, ok := m[first.HashStr()]; !ok {
			b.Fail()
		}
		if _, ok := m[middle.HashStr()]; !ok {
			b.Fail()
		}
		if _, ok := m[last.HashStr()]; !ok {
			b.Fail()
		}
	}

}
