package storage

import (
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
	"log"
)

type (
	Storage struct {
		chats List
	}

	List interface {
		Len() int
		Add(value interface{}) uuid.UUID
		Contains(id uuid.UUID) bool
		Update(id uuid.UUID, value interface{})
		GetByIndex(index int) (interface{}, error)
		GetById(id uuid.UUID) (interface{}, error)
		Delete(key uuid.UUID)
		Display()
	}
)

func NewStorage() *Storage {
	return &Storage{
		chats: NewList(),
	}
}

func (s *Storage) GetChat(identifier string, byName bool) (crdt.Chat, error) {
	log.Println(identifier)
	var (
		numberOfChats = s.chats.Len()
		c             crdt.Chat
		err           error
	)

	if byName {
		for index := 0; index < numberOfChats; index++ {
			var chatValue interface{}
			chatValue, _ = s.chats.GetByIndex(index)
			c = chatValue.(*crdt.ConcreteChat)

			if c.GetName() == identifier {
				return c, nil
			}
		}

		if err != nil || c == nil {
			return nil, NotFound
		}

	}

	// by uuid
	var id uuid.UUID
	id, err = uuid.Parse(identifier)
	if err != nil {
		return nil, InvalidIdentifier
	}

	var chatValue interface{}
	chatValue, err = s.chats.GetById(id)

	if err != nil {
		return nil, NotFound
	}

	return chatValue.(*crdt.ConcreteChat), nil
}

func (s *Storage) SaveChat(c crdt.Chat) {
	id, _ := uuid.Parse(c.GetId())
	if !s.chats.Contains(id) {
		s.chats.Add(c)
		return
	}

	s.chats.Update(id, c)
}

func (s *Storage) DisplayChats() {
	s.chats.Display()
}
