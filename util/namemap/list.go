package namemap

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type Element[V any] struct {
	next, prev *Element[V]
	Kstr       string
	Kname      enc.Name
	Val        V
}

func (e *Element[V]) Next() *Element[V] { return e.next }
func (e *Element[V]) Prev() *Element[V] { return e.prev }

type list[V any] struct {
	head, tail *Element[V]
}

func (l *list[V]) front() *Element[V] { return l.head }
func (l *list[V]) back() *Element[V]  { return l.tail }

func (l *list[V]) remove(e *Element[V]) {
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

func (l *list[V]) pushFront(e *Element[V]) {
	if l.head == nil {
		l.head = e
		l.tail = e
		return
	}
	e.next = l.head
	l.head.prev = e
	l.head = e
}

func (l *list[V]) pushBack(e *Element[V]) {
	if l.tail == nil {
		l.head = e
		l.tail = e
		return
	}
	e.prev = l.tail
	l.tail.next = e
	l.tail = e
}

func (l *list[V]) pushAfter(e, i *Element[V]) {
	e.prev = i
	if i == l.tail {
		l.tail = e
	} else {
		e.next = i.next
		i.next.prev = e
	}
	i.next = e
}

func (l *list[V]) pushBefore(e, i *Element[V]) {
	e.next = i
	if i == l.head {
		l.head = e
	} else {
		e.prev = i.prev
		i.prev.next = e
	}
	i.prev = e
}

func (l *list[V]) moveToFront(e *Element[V]) {
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

func (l *list[V]) moveToBack(e *Element[V]) {
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

func (l *list[V]) insert(e *Element[V], f func(e1, e2 *Element[V]) bool) {
	if l.head == nil {
		l.head = e
		l.tail = e
		return
	}
	i := l.head
	if f(e, i) {
		l.head = e
		e.next = i
		i.prev = e
		return
	}
	for i.next != nil && !f(e, i.next) {
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
