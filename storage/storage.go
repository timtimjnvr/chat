package storage

import (
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
)

type (
	Storage struct {
		chats List
	}

	List interface {
		Len() int
		Add(chat *crdt.Chat) uuid.UUID
		Contains(id uuid.UUID) bool
		Update(id uuid.UUID, chat *crdt.Chat)
		GetByIndex(index int) (*crdt.Chat, error)
		GetById(id uuid.UUID) (*crdt.Chat, error)
		Delete(key uuid.UUID)
		Display()
	}
)

func NewStorage() *Storage {
	return &Storage{
		chats: NewList(),
	}
}

func (s *Storage) GetChat(identifier string, byName bool) (*crdt.Chat, error) {
	var (
		numberOfChats = s.chats.Len()
		c             *crdt.Chat
		err           error
	)

	if byName {
		for index := 0; index < numberOfChats; index++ {
			c, _ = s.chats.GetByIndex(index)
			if c.Name == identifier {
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

	c, err = s.chats.GetById(id)
	if err != nil {
		return nil, NotFound
	}

	return c, nil
}

func (s *Storage) GetChatByIndex(index int) (*crdt.Chat, error) {
	return s.chats.GetByIndex(index)
}

func (s *Storage) SaveChat(c *crdt.Chat) {
	id, _ := uuid.Parse(c.Id)
	if !s.chats.Contains(id) {
		s.chats.Add(c)
		return
	}

	s.chats.Update(id, c)
}

func (s *Storage) DisplayChats() {
	s.chats.Display()
}
