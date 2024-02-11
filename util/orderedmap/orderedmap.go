package orderedmap

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type OrderedMap[V any] struct {
	kv map[string]*Element[string, enc.Name, V]
	ll list[string, enc.Name, V]
	oo Ordering
}

func New[V any](o Ordering) *OrderedMap[V] {
	return &OrderedMap[V]{
		kv: make(map[string]*Element[string, enc.Name, V]),
		oo: o,
	}
}

func (om *OrderedMap[V]) Len() int                             { return len(om.kv) }
func (om *OrderedMap[V]) Front() *Element[string, enc.Name, V] { return om.ll.Front() }
func (om *OrderedMap[V]) Back() *Element[string, enc.Name, V]  { return om.ll.Back() }

func (om *OrderedMap[V]) Copy() *OrderedMap[V] {
	var (
		ret  = New[V](om.oo)
		i, e *Element[string, enc.Name, V]
	)
	for i = om.Front(); i != nil; i = i.Next() {
		e = &Element[string, enc.Name, V]{Kstr: i.Kstr, Kname: i.Kname, Val: i.Val}
		ret.ll.PushBack(e)
		om.kv[e.Kstr] = e
	}
	return ret
}

func (om *OrderedMap[V]) Get(kstr string) (val V, ok bool) {
	e, ok := om.kv[kstr]
	if ok {
		val = e.Val
	}
	return
}

func (om *OrderedMap[V]) GetElement(kstr string) *Element[string, enc.Name, V] {
	e, ok := om.kv[kstr]
	if ok {
		return e
	}
	return nil
}

func (om *OrderedMap[V]) Remove(kstr string) bool {
	e, ok := om.kv[kstr]
	if ok {
		om.ll.Remove(e)
		delete(om.kv, kstr)
	}
	return ok
}

func (om *OrderedMap[V]) Set(kstr string, kname enc.Name, val V, mv MetaV) bool {
	switch om.oo {
	case LatestEntriesFirst:
		return om.setLatestEntriesFirst(kstr, kname, val, mv.Old)
	default:
		return om.setCanonical(kstr, kname, val)
	}
}

func (om *OrderedMap[V]) setCanonical(kstr string, kname enc.Name, val V) bool {
	e, ok := om.kv[kstr]
	if ok {
		e.Val = val
		return true
	}
	e = &Element[string, enc.Name, V]{Kstr: kstr, Kname: kname, Val: val}
	om.ll.Insert(e, func(e1, e2 *Element[string, enc.Name, V]) bool {
		return e1.Kname.Compare(e2.Kname) != 1
	})
	om.kv[kstr] = e
	return false
}

func (om *OrderedMap[V]) setLatestEntriesFirst(kstr string, kname enc.Name, val V, old bool) bool {
	e, ok := om.kv[kstr]
	if ok {
		e.Val = val
		if !old {
			om.ll.MoveToFront(e)
		}
		return true
	}
	e = &Element[string, enc.Name, V]{Kstr: kstr, Kname: kname, Val: val}
	if old {
		om.ll.PushBack(e)
	} else {
		om.ll.PushFront(e)
	}
	om.kv[kstr] = e
	return false
}
