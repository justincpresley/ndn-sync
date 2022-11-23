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

func (l *list[K, V]) PushFront(key K, value V) *Element[K, V] {
	e := &Element[K, V]{Key: key, Value: value}
	if l.head == nil {
		l.head = e
		l.tail = e
		return e
	}
	e.next = l.head
	l.head.prev = e
	l.head = e
	return e
}

func (l *list[K, V]) PushBack(key K, value V) *Element[K, V] {
	e := &Element[K, V]{Key: key, Value: value}
	if l.tail == nil {
		l.head = e
		l.tail = e
		return e
	}
	e.prev = l.tail
	l.tail.next = e
	l.tail = e
	return e
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
