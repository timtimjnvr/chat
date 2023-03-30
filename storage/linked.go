package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
)

var (
	NotFound          = errors.New("Not found")
	InvalidIdentifier = errors.New("Invalid identifier")
)

type (
	element struct {
		key  uuid.UUID
		chat *crdt.Chat
		next *element
	}

	list struct {
		length int
		head   *element
		tail   *element
	}
)

func NewList() *list {
	return &list{}
}

func newElement(chat *crdt.Chat) *element {
	id := uuid.New()
	return &element{
		key:  id,
		chat: chat,
	}
}
func (l *list) Len() int {
	return l.length
}

func (l *list) Display() {
	fmt.Printf("%d chats\n", l.length)

	tmp := l.head
	for tmp != nil {
		fmt.Printf("%s", tmp.chat.Name)
		tmp = tmp.next
	}
}

// Add insert chat at the end of the list and return the key of the inserted chat
func (l *list) Add(chat *crdt.Chat) uuid.UUID {
	e := newElement(chat)
	if l.length == 0 {
		l.head = e
		l.tail = e
		l.length++

		return l.tail.key
	}

	var (
		ptr    = l.head
		length = l.length
	)

	for i := 0; i < length; i++ {
		if ptr.next == nil {
			ptr.next = e
			l.tail = e
			l.length++
		}

		ptr = ptr.next
	}

	return l.tail.key
}

func (l *list) Contains(id uuid.UUID) bool {
	first := l.head
	for first.next != nil && first.key != id {
		first = first.next
	}

	if first.key == id {
		return true
	}

	return false
}

func (l *list) Update(id uuid.UUID, chat *crdt.Chat) {
	first := l.head
	for first.next != nil && first.key != id {
		first = first.next
	}

	if first.key == id {
		first.chat = chat
	}
}

func (l *list) GetByIndex(index int) (*crdt.Chat, error) {
	if index >= l.Len() {
		return nil, NotFound
	}
	i := 0
	for l.head != nil && i != index {
		l.head = l.head.next
	}

	return l.head.chat, nil
}

func (l *list) GetById(id uuid.UUID) (*crdt.Chat, error) {
	first := l.head
	for first.next != nil && first.key != id {
		first = first.next
	}

	if first.key == id {
		return first.chat, nil
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
