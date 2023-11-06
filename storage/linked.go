package storage

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
)

type (
	value interface {
		GetID() uuid.UUID
		GetName() string

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

var (
	AlreadyInListWithNameErr = errors.New("already a chat with this name in the listOld")
	AlreadyInListWithIDErr   = errors.New("already a chat with this ID in the listOld")
	InvalidChatErr           = errors.New("invalid chat")
	NotFoundErr              = errors.New("not found")
	InvalidIdentifierErr     = errors.New("invalid identifier")
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
		fmt.Printf("- %s\n", tmp.v.GetName())

		tmp = tmp.next
	}
}

// Add insert chat at the end of the listOld and return the key of the inserted chat
func (l *list[T]) Add(v T) (uuid.UUID, error) {
	e := newElement(v)

	id, err := uuid.Parse(v.GetID().String())

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
		if ptr.v.GetID() == v.GetID() {
			return uuid.UUID{}, AlreadyInListWithIDErr
		}

		if ptr.v.GetName() == v.GetName() {
			return uuid.UUID{}, AlreadyInListWithNameErr
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

func (l *list[T]) Contains(id uuid.UUID) bool {
	if l.length == 0 {
		return false
	}

	tmp := l.head
	for tmp.next != nil && tmp.v.GetID() != id {
		tmp = tmp.next
	}

	if tmp.v.GetID() == id {
		return true
	}

	return false
}

func (l *list[T]) Update(v T) error {
	if v == nil {
		return InvalidChatErr
	}

	id, err := uuid.Parse(v.GetID().String())
	if err != nil {
		return InvalidIdentifierErr
	}

	if l.length == 0 {
		return NotFoundErr
	}

	tmp := l.head
	for tmp.next != nil && tmp.v.GetID() != id {
		tmp = tmp.next
	}

	if tmp.v.GetID() == id {
		tmp.v = v
		return nil
	}

	return NotFoundErr
}

func (l *list[T]) GetByIndex(index int) (T, error) {
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

	return tmp.v, nil
}

func (l *list[T]) GetById(id uuid.UUID) (T, error) {
	if l.length == 0 {
		return nil, NotFoundErr
	}

	tmp := l.head
	for tmp.next != nil && tmp.v.GetID() != id {
		tmp = tmp.next
	}

	if tmp.v.GetID() == id {
		return tmp.v, nil
	}

	return nil, NotFoundErr
}

func (l *list[T]) Delete(id uuid.UUID) {
	if l.length == 0 {
		return
	}

	var previous, tmp *element[T]

	// remove first elementOld
	if l.head.v.GetID() == id {
		l.head = l.head.next
		l.length--
		return
	}

	// iterate until elementOld found or end of listOld
	previous = l.head
	tmp = l.head.next
	for tmp != nil && tmp.next != nil && tmp.v.GetID() != id {
		previous = tmp
		tmp = tmp.next
	}

	// elementOld found or end of listOld
	if tmp != nil && tmp.v.GetID() == id {
		previous.next = tmp.next
		l.length--
	}
}
