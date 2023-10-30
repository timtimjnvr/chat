package storage

import (
	"fmt"
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
		Update(chat *crdt.Chat) error
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
			return nil, NotFoundErr
		}
	}

	// by uuid
	var id uuid.UUID
	id, err = uuid.Parse(identifier)
	if err != nil {
		return nil, InvalidIdentifierErr
	}

	c, err = s.chats.GetById(id)
	if err != nil {
		return nil, NotFoundErr
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

	s.chats.Update(c)
}

func (s *Storage) DeleteChatById(identifier string) {
	id, _ := uuid.Parse(identifier)

	s.chats.Delete(id)
}

func (s *Storage) DisplayChats() {
	s.chats.Display()
}

func (s *Storage) GetNumberOfChats() int {
	return s.chats.Len()
}

func (s *Storage) CreateNewChat(name string, myInfos *crdt.NodeInfos) *crdt.Chat {
	newChat := crdt.NewChat(name)
	newChat.SaveNode(myInfos)
	s.SaveChat(newChat)
	return newChat
}

func (s *Storage) AddChat(name string, id string, myInfos *crdt.NodeInfos) *crdt.Chat {
	newChat := crdt.NewChat(name)
	newChat.Id = id
	newChat.SaveNode(myInfos)
	s.SaveChat(newChat)
	return newChat
}

func (s *Storage) RemoveNodeSlotFromStorage(slot uint8) {
	var (
		index         = 0
		numberOfChats = s.GetNumberOfChats()
		c             *crdt.Chat
		err           error
	)

	for index < numberOfChats && err == nil {
		c, err = s.GetChatByIndex(index)
		if err != nil {
			index++
			continue
		}

		nodeName, err2 := c.RemoveNodeBySlot(slot)
		if err2 == nil {
			fmt.Printf("%s leaved chat %s\n", nodeName, c.Name)
		}

		index++
	}
}
