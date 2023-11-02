package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
)

type (
	value interface {
		*crdt.Chat | *crdt.NodeInfos
	}

	element[T value] struct {
		v    T
		next *element[T]
	}

	list[T value] struct {
		typeName string
		length   int
		head     *element[T]
		tail     *element[T]
	}
)

func NewChatList() *list[*crdt.Chat] {
	return &list[*crdt.Chat]{
		typeName: "chats",
	}
}

func NewNodeList() *list[*crdt.NodeInfos] {
	return &list[*crdt.NodeInfos]{
		typeName: "nodes",
	}
}

func newElement[T value](v T) *element[T] {
	return &element[T]{
		v:    v,
		next: nil,
	}
}

func (l *list[T]) Len() int {
	return l.length
}

func (l *list[T]) Display() {
	fmt.Printf("%d %s\n", l.length, l.typeName)

	tmp := l.head
	for tmp != nil {
		fmt.Printf("- %s\n", tmp.chat.Name)
		tmp = tmp.next
	}
}

// Add insert chat at the end of the listOld and return the key of the inserted chat
func (l *listOld) Add(chat *crdt.Chat) (uuid.UUID, error) {
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

func (l *listOld) Contains(id uuid.UUID) bool {
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

func (l *listOld) Update(chat *crdt.Chat) error {
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

func (l *listOld) GetByIndex(index int) (*crdt.Chat, error) {
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

func (l *listOld) GetById(id uuid.UUID) (*crdt.Chat, error) {
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

func (l *listOld) Delete(id uuid.UUID) {
	if l.length == 0 {
		return
	}

	var previous, tmp *elementOld

	// remove first elementOld
	if l.head.chat.Id == id {
		l.head = l.head.next
		l.length--
		return
	}

	// iterate until elementOld found or end of listOld
	previous = l.head
	tmp = l.head.next
	for tmp != nil && tmp.next != nil && tmp.chat.Id != id {
		previous = tmp
		tmp = tmp.next
	}

	// elementOld found or end of listOld
	if tmp != nil && tmp.chat.Id == id {
		previous.next = tmp.next
		l.length--
	}
}
