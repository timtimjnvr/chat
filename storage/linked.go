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
		chat *crdt.Chat
		next *element
	}

	list struct {
		length int
		head   *element
		tail   *element
	}
)

func NewList() List {
	return &list{}
}

func newElement(chat *crdt.Chat) *element {
	return &element{
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
		fmt.Printf("- %s\n", tmp.chat.Name)
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
		id, _ := uuid.Parse(l.tail.chat.Id)
		return id
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

	id, _ := uuid.Parse(l.tail.chat.Id)
	return id
}

func (l *list) Contains(id uuid.UUID) bool {
	if l.length == 0 {
		return false
	}

	first := l.head
	for first.next != nil && first.chat.Id != id.String() {
		first = first.next
	}

	if first.chat.Id == id.String() {
		return true
	}

	return false
}

func (l *list) Update(id uuid.UUID, chat *crdt.Chat) {
	first := l.head
	for first.next != nil && first.chat.Id != id.String() {
		first = first.next
	}

	if first.chat.Id == id.String() {
		first.chat = chat
	}
}

func (l *list) GetByIndex(index int) (*crdt.Chat, error) {
	if index >= l.Len() {
		return nil, NotFound
	}
	var (
		i   = 0
		tmp = l.head
	)
	for tmp.next != nil && i != index {
		tmp = tmp.next
		i++
	}

	return tmp.chat, nil
}

func (l *list) GetById(id uuid.UUID) (*crdt.Chat, error) {
	first := l.head
	for first.next != nil && first.chat.Id != id.String() {
		first = first.next
	}

	if first.chat.Id == id.String() {
		return first.chat, nil
	}

	return nil, NotFound
}

func (l *list) Delete(id uuid.UUID) {
	var previous, tmp *element

	// remove first element
	if l.head.chat.Id == id.String() {
		l.head = l.head.next
		l.length--
		return
	}

	// second or more
	previous = l.head
	tmp = l.head.next
	for tmp != nil && tmp.next != nil && tmp.chat.Id != id.String() {
		previous = tmp
		tmp = tmp.next
		tmp = previous.next
	}

	if tmp != nil && tmp.chat.Id == id.String() {
		previous.next = tmp.next
		l.length--
	}
}
