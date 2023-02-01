package linked

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var NotFound = errors.New("Not found")

type (
	element struct {
		key   uuid.UUID
		value interface{}
		next  *element
	}

	list struct {
		length int
		head   *element
		tail   *element
	}

	List interface {
		Len() int
		Add(value interface{}) uuid.UUID
		GetByIndex(index int) (*element, error)
		GetById(id uuid.UUID) (*element, error)
		Delete(key uuid.UUID)
	}

	Element interface {
	}
)

func NewList() List {
	return &list{}
}

func newElement(value interface{}) *element {
	id := uuid.New()
	return &element{
		key:   id,
		value: value,
	}
}
func (l *list) Len() int {
	return l.length
}

func (l *list) Display() {
	for l.head != nil {
		fmt.Printf("%v->", l.head.value)
		l.head = l.head.next
	}
}

// Add insert value at the end of the list and return the key of the inserted value
func (l *list) Add(value interface{}) uuid.UUID {
	if l.head == nil {
		l.head = newElement(value)
		l.tail = l.head
		l.length++
	} else {
		l.tail.next = newElement(value)
		l.tail = newElement(value)
		l.length++
	}
	return l.tail.key
}

func (l *list) GetByIndex(index int) (*element, error) {
	if index >= l.Len() {
		return nil, NotFound
	}
	i := 0
	for l.head != nil && i != index {
		l.head = l.head.next
	}

	return l.head, nil
}

func (l *list) GetById(id uuid.UUID) (*element, error) {
	for l.head != nil && l.head.key != id {
		l.head = l.head.next
	}

	if l.head.key == id {
		return l.head, nil
	}

	return nil, NotFound
}

func (l *list) Delete(key uuid.UUID) {
	var previous, tmp *element

	// remove first element
	if l.head.key == key {
		l.head = l.head.next
		l.length--
		return
	}

	// second or more
	previous = l.head
	tmp = l.head.next
	for tmp != nil && tmp.next != nil && key != tmp.key {
		previous = tmp
		tmp = tmp.next
		tmp = previous.next
	}

	if tmp != nil && key != tmp.key {
		previous.next = tmp.next
	}

	l.length--
}
