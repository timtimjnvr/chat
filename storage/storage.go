package storage

import (
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
)

type (
	Storage struct {
		chats List
	}
)

func NewStorage() *Storage {
	return &Storage{
		chats: NewList(),
	}
}

func (s *Storage) GetChat(identifier string, byName bool) (crdt.Chat, error) {
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
	// TODO : save chat into chained list
}
