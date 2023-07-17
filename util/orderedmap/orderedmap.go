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

func (om *OrderedMap[V]) Len() int                     { return len(om.kv) }
func (om *OrderedMap[V]) Front() *Element[string, enc.Name, V] { return om.ll.Front() }
func (om *OrderedMap[V]) Back() *Element[string, enc.Name, V]  { return om.ll.Back() }

func (om *OrderedMap[V]) Copy() *OrderedMap[V] {
	var (
		ret  = New[V](om.oo)
		i, e *Element[string, enc.Name, V]
	)
	for i = om.Front(); i != nil; i = i.Next() {
		e = &Element[string, enc.Name, V]{Kstring: i.Kstring, Kname: i.Kname, Value: i.Value}
		ret.ll.PushBack(e)
		om.kv[e.Kstring] = e
	}
	return ret
}

func (om *OrderedMap[V]) Get(key string) (val V, ok bool) {
	e, ok := om.kv[key]
	if ok {
		val = e.Value
	}
	return
}

func (om *OrderedMap[V]) GetElement(key string) *Element[string, enc.Name, V] {
	e, ok := om.kv[key]
	if ok {
		return e
	}
	return nil
}

func (om *OrderedMap[V]) Remove(key string) bool {
	e, ok := om.kv[key]
	if ok {
		om.ll.Remove(e)
		delete(om.kv, key)
	}
	return ok
}

func (om *OrderedMap[V]) Set(kstring string, kname enc.Name, value V, mv MetaV) bool {
	switch om.oo {
	case LatestEntriesFirst:
		return om.setLatestEntriesFirst(kstring, kname, value, mv.Old)
	default:
		return om.setCanonical(kstring, kname, value)
	}
}

func (om *OrderedMap[V]) setCanonical(kstring string, kname enc.Name, value V) bool {
	e, ok := om.kv[kstring]
	if ok {
		e.Value = value
		return true
	}
	e = &Element[string, enc.Name, V]{Kstring: kstring, Kname: kname, Value: value}
	om.ll.Insert(e, func(e1, e2 *Element[string, enc.Name, V]) bool {
		return e1.Kname.Compare(e2.Kname) == 1
	})
	om.kv[kstring] = e
	return false
}

func (om *OrderedMap[V]) setLatestEntriesFirst(kstring string, kname enc.Name, value V, old bool) bool {
	e, ok := om.kv[kstring]
	if ok {
		e.Value = value
		if !old {
			om.ll.MoveToFront(e)
		}
		return true
	}
	e = &Element[string, enc.Name, V]{Kstring: kstring, Kname: kname, Value: value}
	if old {
		om.ll.PushBack(e)
	} else {
		om.ll.PushFront(e)
	}
	om.kv[kstring] = e
	return false
}
