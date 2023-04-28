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
	var (
		ret  = New[K, V]()
		i, e *Element[K, V]
	)
	for i = om.Front(); i != nil; i = i.Next() {
		e = &Element[K, V]{Key: i.Key, Value: i.Value}
		ret.ll.PushBack(e)
		om.kv[e.Key] = e
	}
	return ret
}

func (om *OrderedMap[K, V]) Get(key K) (val V, ok bool) {
	e, ok := om.kv[key]
	if ok {
		val = e.Value
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
	e, ok := om.kv[key]
	if ok {
		e.Value = value
		if !old {
			om.ll.MoveToFront(e)
		}
		return true
	}
	e = &Element[K, V]{Key: key, Value: value}
	if old {
		om.ll.PushBack(e)
	} else {
		om.ll.PushFront(e)
	}
	om.kv[key] = e
	return false
}

func (om *OrderedMap[K, V]) Remove(key K) bool {
	e, ok := om.kv[key]
	if ok {
		om.ll.Remove(e)
		delete(om.kv, key)
	}
	return ok
}
