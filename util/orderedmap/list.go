package orderedmap

type Element[K comparable, V any] struct {
	next, prev *Element[K, V]
	Key        K
	Value      V
}

func (e *Element[K, V]) Next() *Element[K, V] { return e.next }
func (e *Element[K, V]) Prev() *Element[K, V] { return e.prev }

type list[K comparable, V any] struct {
	head, tail *Element[K, V]
}

func (l *list[K, V]) Front() *Element[K, V] { return l.head }
func (l *list[K, V]) Back() *Element[K, V]  { return l.tail }

func (l *list[K, V]) Remove(e *Element[K, V]) {
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

func (l *list[K, V]) PushFront(e *Element[K, V]) {
	if l.head == nil {
		l.head = e
		l.tail = e
		return
	}
	e.next = l.head
	l.head.prev = e
	l.head = e
}

func (l *list[K, V]) PushBack(e *Element[K, V]) {
	if l.tail == nil {
		l.head = e
		l.tail = e
		return
	}
	e.prev = l.tail
	l.tail.next = e
	l.tail = e
}

func (l *list[K, V]) PushAfter(e, i *Element[K, V]) {
	e.prev = i
	if i == l.tail {
		l.tail = e
	} else {
		e.next = i.next
		i.next.prev = e
	}
	i.next = e
}

func (l *list[K, V]) PushBefore(e, i *Element[K, V]) {
	e.next = i
	if i == l.head {
		l.head = e
	} else {
		e.prev = i.prev
		i.prev.next = e
	}
	i.prev = e
}

func (l *list[K, V]) MoveToFront(e *Element[K, V]) {
	if e.prev == nil {
		return
	}
	e.prev.next = e.next
	if e.next == nil {
		l.tail = e.prev
	} else {
		e.next.prev = e.prev
	}
	e.next = l.head
	e.prev = nil
	l.head.prev = e
	l.head = e
}

func (l *list[K, V]) MoveToBack(e *Element[K, V]) {
	if e.next == nil {
		return
	}
	if e.prev == nil {
		l.head = e.next
	} else {
		e.prev.next = e.next
	}
	e.next.prev = e.prev
	e.next = nil
	e.prev = l.tail
	l.tail.next = e
	l.tail = e
}
