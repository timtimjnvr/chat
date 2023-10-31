package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
)

type (
	storage struct {
		chats List
	}

	List interface {
		Len() int
		Add(chat *crdt.Chat) (uuid.UUID, error)
		Contains(id uuid.UUID) bool
		Update(chat *crdt.Chat) error
		GetByIndex(index int) (*crdt.Chat, error)
		GetById(id uuid.UUID) (*crdt.Chat, error)
		Delete(key uuid.UUID)
		Display()
	}

	Storage interface {
		AddNewChat(chatName string) (uuid.UUID, error)
		AddChat(chat *crdt.Chat) error
		RemoveChat(chatID uuid.UUID)
		AddNodeToChat(nodeInfos *crdt.NodeInfos, chatID uuid.UUID) error
		RemoveNodeFromChat(nodeSlot uint8, chatID uuid.UUID) error
		RemoveNodeSlotFromStorage(slot uint8)
		GetNumberOfChats() int
		DisplayChatUsers(chatID uuid.UUID) error
		DisplayChats()
	}
)

func NewStorage() Storage {
	return &storage{
		chats: NewList(),
	}
}

func (s *storage) AddNewChat(chatName string) (uuid.UUID, error) {
	newChat := crdt.NewChat(chatName)
	return s.chats.Add(newChat)
}

func (s *storage) AddChat(chat *crdt.Chat) error {
	_, err := s.chats.Add(chat)
	return err
}

func (s *storage) RemoveChat(chatID uuid.UUID) {
	s.chats.Delete(chatID)
}

func (s *storage) AddNodeToChat(nodeInfos *crdt.NodeInfos, chatID uuid.UUID) error {
	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return err
	}

	c.SaveNode(nodeInfos)
	return nil
}

func (s *storage) RemoveNodeFromChat(nodeSlot uint8, chatID uuid.UUID) error {
	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return err
	}

	_, err = c.RemoveNodeBySlot(nodeSlot)
	return err
}

func (s *storage) GetChatByIndex(index int) (*crdt.Chat, error) {
	return s.chats.GetByIndex(index)
}

func (s *storage) SaveChat(c *crdt.Chat) error {
	id, err := uuid.Parse(c.Id)
	if err != nil {
		return InvalidIdentifierErr
	}
	if !s.chats.Contains(id) {
		_, err = s.chats.Add(c)
		if err != nil {
			return err
		}
	}

	return s.chats.Update(c)
}

func (s *storage) DisplayChats() {
	s.chats.Display()
}

func (s *storage) DisplayChatUsers(chatID uuid.UUID) error {
	c, err := s.chats.GetById(chatID)
	if err != nil {
		return err
	}

	c.DisplayUsers()
	return nil
}

func (s *storage) GetNumberOfChats() int {
	return s.chats.Len()
}

func (s *storage) RemoveNodeSlotFromStorage(slot uint8) {
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

func (s *storage) getChat(identifier string, byName bool) (*crdt.Chat, error) {
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
