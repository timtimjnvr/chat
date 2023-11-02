package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
)

var (
	AlreadyInListWithNameErr = errors.New("already a chat with this name in the list")
	AlreadyInListWithIDErr   = errors.New("already a chat with this ID in the list")
	InvalidChatErr           = errors.New("invalid chat")
	NotFoundErr              = errors.New("not found")
	InvalidIdentifierErr     = errors.New("invalid identifier")
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

func NewList() *list {
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
func (l *list) Add(chat *crdt.Chat) (uuid.UUID, error) {
	e := newElement(chat)
	id, err := uuid.Parse(chat.Id.String())
	if err != nil {
		return uuid.UUID{}, InvalidIdentifierErr
	}

	if l.length == 0 {
		l.head = e
		l.tail = e
		l.length++
		return id, nil
	}

	var (
		ptr    = l.head
		length = l.length
	)

	for i := 0; i < length; i++ {
		if ptr.chat.Name == chat.Name {
			return uuid.UUID{}, AlreadyInListWithNameErr
		}

		if ptr.chat.Id == chat.Id {
			return uuid.UUID{}, AlreadyInListWithIDErr
		}

		if ptr.next == nil {
			ptr.next = e
			l.tail = e
			l.length++
		}

		ptr = ptr.next
	}

	return id, nil
}

func (l *list) Contains(id uuid.UUID) bool {
	if l.length == 0 {
		return false
	}

	tmp := l.head
	for tmp.next != nil && tmp.chat.Id != id {
		tmp = tmp.next
	}

	if tmp.chat.Id == id {
		return true
	}

	return false
}

func (l *list) Update(chat *crdt.Chat) error {
	if chat == nil {
		return InvalidChatErr
	}

	id, err := uuid.Parse(chat.Id.String())
	if err != nil {
		return InvalidIdentifierErr
	}

	if l.length == 0 {
		return NotFoundErr
	}

	tmp := l.head
	for tmp.next != nil && tmp.chat.Id != id {
		tmp = tmp.next
	}

	if tmp.chat.Id == id {
		tmp.chat = chat
		return nil
	}

	return NotFoundErr
}

func (l *list) GetByIndex(index int) (*crdt.Chat, error) {
	if index >= l.Len() {
		return nil, NotFoundErr
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
	if l.length == 0 {
		return nil, NotFoundErr
	}

	first := l.head
	for first.next != nil && first.chat.Id != id {
		first = first.next
	}

	if first.chat.Id == id {
		return first.chat, nil
	}

	return nil, NotFoundErr
}

func (l *list) Delete(id uuid.UUID) {
	if l.length == 0 {
		return
	}

	var previous, tmp *element

	// remove first element
	if l.head.chat.Id == id {
		l.head = l.head.next
		l.length--
		return
	}

	// iterate until element found or end of list
	previous = l.head
	tmp = l.head.next
	for tmp != nil && tmp.next != nil && tmp.chat.Id != id {
		previous = tmp
		tmp = tmp.next
	}

	// element found or end of list
	if tmp != nil && tmp.chat.Id == id {
		previous.next = tmp.next
		l.length--
	}
}
