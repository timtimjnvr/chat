package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github/timtimjnvr/chat/crdt"
	"log"
)

type (
	Storage struct {
		chats *List[*crdt.Chat]
		nodes *List[*crdt.NodeInfos]
	}
)

func NewStorage() *Storage {
	return &Storage{
		chats: NewChatList(),
		nodes: NewNodeList(),
	}
}

func (s *Storage) GetNumberOfChats() int {
	return s.chats.Len()
}

func (s *Storage) GetChatID(chatName string) (uuid.UUID, error) {
	c, err := s.getChat(chatName, true)
	if err != nil {
		return uuid.UUID{}, err
	}

	return c.Id, nil
}

func (s *Storage) GetChatName(id uuid.UUID) (string, error) {
	c, err := s.getChat(id.String(), false)
	if err != nil {
		return "", err
	}

	return c.Name, nil
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

// AddNodeToChat add a node to a given chat identified by id. The node slot need to be set
func (s *Storage) AddNodeToChat(node *crdt.NodeInfos, chatID uuid.UUID) error {
	if !s.nodes.Contains(node.Id) {
		_, _ = s.nodes.Add(node)
	}

	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return err
	}

	c.SaveNode(node.Slot)
	return nil
}

func (s *Storage) RemoveNodeFromChat(nodeSlot uint8, chatID uuid.UUID) error {
	c, err := s.getChat(chatID.String(), false)
	if err != nil {
		return err
	}

	err = c.RemoveNode(nodeSlot)
	n, err := s.GetNodeBySlot(nodeSlot)
	if err != nil {
		return err
	}

	fmt.Printf("%s leaved chat\n", n.Name)
	return nil
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

func (s *Storage) GetNodeBySlot(slot uint8) (*crdt.NodeInfos, error) {
	var (
		numberOfNodes = s.nodes.Len()
		n             *crdt.NodeInfos
		err           error
	)
	if s.nodes.length == 0 {
		return nil, NotFoundErr
	}

	for index := 0; index < numberOfNodes; index++ {
		n, _ = s.nodes.GetByIndex(index)
		if n.Slot == slot {
			return n, nil
		}

		if err != nil || n == nil {
			return nil, NotFoundErr
		}
	}

	return n, nil
}

func (s *Storage) IsSlotUsedByOtherChats(slotToFind uint8, excludeChatForSearch uuid.UUID) bool {
	var (
		numberOfChats = s.GetNumberOfChats()
		err           error
	)

	for index := 0; index < numberOfChats && err == nil; index++ {
		tmpChat, _ := s.chats.GetByIndex(index)
		if tmpChat.Id == excludeChatForSearch {
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

func (s *Storage) RemoveNodeSlotFromStorage(slot uint8) {
	var (
		index         = 0
		numberOfChats = s.GetNumberOfChats()
		c             *crdt.Chat
		err           error
	)

	node, err := s.GetNodeBySlot(slot)
	if err != nil {
		return
	}

	for index < numberOfChats && err == nil {
		c, _ = s.chats.GetByIndex(index)

		err = c.RemoveNode(slot)
		if err == nil {
			fmt.Printf("%s leaved chat %s\n", node.Name, c.Name)
		}

		index++
	}

	s.nodes.Delete(node.Id)
}

func (s *Storage) DisplayChats() {
	s.chats.Display()
}

func (s *Storage) DisplayNodes() {
	s.nodes.Display()
}

func (s *Storage) DisplayChatUsers(chatID uuid.UUID) error {
	c, err := s.chats.GetById(chatID)
	if err != nil {
		return err
	}

	log.Printf("chat name : %s\n", c.Name)
	for slot := range c.GetSlots() {
		n, _ := s.nodes.GetByIndex(slot)
		log.Printf("- %s (Address: %s, Port: %s, Slot: %d)\n", n.Name, n.Address, n.Port, n.Slot)
	}

	return nil
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
