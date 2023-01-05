/*
 This module is a modified version of the original work.
 Original work can be found at:
          github.com/elliotchance/orderedmap

 The license provided (copy_of_license.md) covers the files
 within this directory. In addition, the changes or modifications
 done are described in changes.md.
 I do not claim ownership or creation of this module. All
 credit should be given to the original author.
*/

package orderedmap

type OrderedMap[K comparable, V any] struct {
	kv map[K]*Element[K, V]
	ll list[K, V]
}

func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		kv: make(map[K]*Element[K, V]),
	}
}

func (om *OrderedMap[K, V]) Len() int              { return len(om.kv) }
func (om *OrderedMap[K, V]) Front() *Element[K, V] { return om.ll.Front() }
func (om *OrderedMap[K, V]) Back() *Element[K, V]  { return om.ll.Back() }

func (om *OrderedMap[K, V]) Copy() *OrderedMap[K, V] {
	ret := New[K, V]()
	for e := om.Front(); e != nil; e = e.Next() {
		ret.Set(e.Key, e.Value, true)
	}
	return ret
}

func (om *OrderedMap[K, V]) Get(key K) (val V, ok bool) {
	v, ok := om.kv[key]
	if ok {
		val = v.Value
	}
	return
}

func (om *OrderedMap[K, V]) GetElement(key K) *Element[K, V] {
	e, ok := om.kv[key]
	if ok {
		return e
	}
	return nil
}

func (om *OrderedMap[K, V]) Set(key K, value V, old bool) bool {
	e, alreadyExist := om.kv[key]
	if alreadyExist {
		e.Value = value
		if !old {
			om.ll.MoveToFront(e)
		}
		return true
	}
	if old {
		e = om.ll.PushBack(key, value)
	} else {
		e = om.ll.PushFront(key, value)
	}
	om.kv[key] = e
	return false
}

func (om *OrderedMap[K, V]) Remove(key K) bool {
	element, ok := om.kv[key]
	if ok {
		om.ll.Remove(element)
		delete(om.kv, key)
	}
	return ok
}
