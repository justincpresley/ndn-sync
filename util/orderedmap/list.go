package orderedmap

type Element[K, Z, V any] struct {
	next, prev *Element[K, Z, V]
	Kstring    K
	Kname      Z
	Value      V
}

func (e *Element[K, Z, V]) Next() *Element[K, Z, V] { return e.next }
func (e *Element[K, Z, V]) Prev() *Element[K, Z, V] { return e.prev }

type list[K, Z, V any] struct {
	head, tail *Element[K, Z, V]
}

func (l *list[K, Z, V]) Front() *Element[K, Z, V] { return l.head }
func (l *list[K, Z, V]) Back() *Element[K, Z, V]  { return l.tail }

func (l *list[K, Z, V]) Remove(e *Element[K, Z, V]) {
	if e.prev == nil {
		l.head = e.next
	} else {
		e.prev.next = e.next
	}
	if e.next == nil {
		l.tail = e.prev
	} else {
		e.next.prev = e.prev
	}
	e.next = nil
	e.prev = nil
}

func (l *list[K, Z, V]) PushBack(e *Element[K, Z, V]) {
	if l.tail == nil {
		l.head = e
		l.tail = e
		return
	}
	e.prev = l.tail
	l.tail.next = e
	l.tail = e
}

func (l *list[K, Z, V]) Insert(e *Element[K, Z, V], f func(e1, e2 *Element[K, Z, V]) bool) {
	if l.head == nil {
		l.head = e
		l.tail = e
		return
	}
	i := l.head
	if !f(i, e) {
		l.head = e
		e.next = i
		i.prev = e
		return
	}
	for i.next != nil && f(i.next, e) {
		i = i.next
	}
	if i.next == nil {
		e.next = nil
		i.next = e
		l.tail = e
	} else {
		e.next = i.next
		i.next.prev = e
		i.next = e
	}
}
