package namemap

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type NameMap[V any] struct {
	kv map[string]*Element[V]
	ll list[V]
	oo Ordering
}

func New[V any](o Ordering) *NameMap[V] {
	return &NameMap[V]{
		kv: make(map[string]*Element[V]),
		oo: o,
	}
}

func (m *NameMap[V]) Len() int           { return len(m.kv) }
func (m *NameMap[V]) Front() *Element[V] { return m.ll.front() }
func (m *NameMap[V]) Back() *Element[V]  { return m.ll.back() }

func (m *NameMap[V]) Copy() *NameMap[V] {
	ret := &NameMap[V]{
		kv: make(map[string]*Element[V], len(m.kv)),
		oo: m.oo,
	}
	var e *Element[V]
	for i := m.Front(); i != nil; i = i.Next() {
		e = &Element[V]{Kstr: i.Kstr, Kname: i.Kname, Val: i.Val}
		ret.kv[e.Kstr] = e
		ret.ll.pushBack(e)
	}
	return ret
}

func (m *NameMap[V]) Get(kstr string) (V, bool) {
	e, ok := m.kv[kstr]
	if !ok {
		return *new(V), ok
	}
	return e.Val, ok
}

func (m *NameMap[V]) GetElement(kstr string) *Element[V] {
	e, ok := m.kv[kstr]
	if !ok {
		return nil
	}
	return e
}

func (m *NameMap[V]) Remove(kstr string) bool {
	e, ok := m.kv[kstr]
	if ok {
		delete(m.kv, kstr)
		m.ll.remove(e)
	}
	return ok
}

func (m *NameMap[V]) Set(kstr string, kname enc.Name, val V, mv MetaV) bool {
	switch m.oo {
	case LatestEntriesFirst:
		return m.setLatestEntriesFirst(kstr, kname, val, mv.Old)
	default:
		return m.setCanonical(kstr, kname, val)
	}
}

func (m *NameMap[V]) setCanonical(kstr string, kname enc.Name, val V) bool {
	e, ok := m.kv[kstr]
	if ok {
		e.Val = val
		return true
	}
	e = &Element[V]{Kstr: kstr, Kname: kname, Val: val}
	m.kv[kstr] = e
	m.ll.insert(e, func(e1, e2 *Element[V]) bool {
		return e1.Kname.Compare(e2.Kname) != 1
	})
	return false
}

func (m *NameMap[V]) setLatestEntriesFirst(kstr string, kname enc.Name, val V, old bool) bool {
	e, ok := m.kv[kstr]
	if ok {
		e.Val = val
		if !old {
			m.ll.moveToFront(e)
		}
		return true
	}
	e = &Element[V]{Kstr: kstr, Kname: kname, Val: val}
	m.kv[kstr] = e
	if old {
		m.ll.pushBack(e)
	} else {
		m.ll.pushFront(e)
	}
	return false
}
