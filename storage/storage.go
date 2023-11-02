package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
)

type (
	Storage struct {
		chats *list
		nodes *list
	}
)

func NewStorage() *Storage {
	return &Storage{
		chats: NewList(),
	}
}

func (s *Storage) GetNodeFromChatBySlot(chatID uuid.UUID, slot uint8) (*crdt.NodeInfos, error) {
	c, err := s.getChat(chatID.String(), true)
	if err != nil {
		return nil, err
	}

	return c.GetNodeBySlot(slot)
}

func (s *Storage) GetChatName(id uuid.UUID) (string, error) {
	c, err := s.getChat(id.String(), false)
	if err != nil {
		return "", err
	}

	return c.Name, nil
}

func (s *Storage) Exist(chatID uuid.UUID) bool {
	return s.chats.Contains(chatID)
}

func (s *Storage) GetChatID(chatName string) (uuid.UUID, error) {
	c, err := s.getChat(chatName, true)
	if err != nil {
		return uuid.UUID{}, err
	}

	return c.Id, nil
}

func (s *Storage) GetNewCurrentChatID() (uuid.UUID, error) {
	if s.chats.Len() == 0 {
		return uuid.UUID{}, errors.New("no chats in storage")
	}

	c, err := s.chats.GetByIndex(0)
	if err != nil {
		return uuid.UUID{}, err
	}

	return c.Id, nil
}

func (s *Storage) AddMessageToChat(message *crdt.Message, chatID uuid.UUID) error {

	c, err := s.getChat(chatID.String(), true)
	if err != nil {
		return err
	}

	if c.ContainsMessage(message) {
		return errors.New("message already saved")
	}

	c.SaveMessage(message)
	return nil
}

func (s *Storage) AddNewChat(chatName string) (uuid.UUID, error) {
	newChat := crdt.NewChat(chatName)
	return s.chats.Add(newChat)
}

func (s *Storage) AddChat(chat *crdt.Chat) error {
	_, err := s.chats.Add(chat)
	return err
}

func (s *Storage) RemoveChat(chatID uuid.UUID) {
	s.chats.Delete(chatID)
}

func (s *Storage) AddNodeToChat(nodeInfos *crdt.NodeInfos, chatID uuid.UUID) error {
	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return err
	}

	c.SaveNode(nodeInfos)
	return nil
}

func (s *Storage) IsSlotUsedByOtherChats(slotToFind uint8, myNodeID uuid.UUID, exceptChatID uuid.UUID) bool {
	var (
		index         = 0
		numberOfChats = s.GetNumberOfChats()
		err           error
	)

	for index < numberOfChats && err == nil {
		tmpChat, _ := s.GetChatByIndex(index)
		if tmpChat.Id == exceptChatID {
			index++
			continue
		}

		// don't kill connections in use in other chats
		tmpSlots := tmpChat.GetSlots()
		for _, slot := range tmpSlots {
			if slot == slotToFind {
				return true
			}
		}
	}

	return false
}

func (s *Storage) RemoveNodeFromChat(nodeSlot uint8, chatID uuid.UUID) error {
	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return err
	}

	_, err = c.RemoveNodeBySlot(nodeSlot)
	return err
}

func (s *Storage) GetChatByIndex(index int) (*crdt.Chat, error) {
	return s.chats.GetByIndex(index)
}

func (s *Storage) SaveChat(c *crdt.Chat) error {
	if !s.chats.Contains(c.Id) {
		_, err := s.chats.Add(c)
		if err != nil {
			return err
		}
	}

	return s.chats.Update(c)
}

func (s *Storage) DisplayChats() {
	s.chats.Display()
}

func (s *Storage) DisplayChatUsers(chatID uuid.UUID) error {
	c, err := s.chats.GetById(chatID)
	if err != nil {
		return err
	}

	c.DisplayUsers()
	return nil
}

func (s *Storage) GetNumberOfChats() int {
	return s.chats.Len()
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

func (s *Storage) GetSlots(chatID uuid.UUID) ([]uint8, error) {
	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return []uint8{}, err
	}

	return c.GetSlots(), nil
}

func (s *Storage) getChat(identifier string, byName bool) (*crdt.Chat, error) {
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
