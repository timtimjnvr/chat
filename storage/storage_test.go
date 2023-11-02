package storage

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github/timtimjnvr/chat/crdt"
	"testing"
)

func Test_storage_AddNewChat(t *testing.T) {
	s := NewStorage()
	name := "my-chat"
	id, err := s.AddNewChat(name)
	assert.Nil(t, err)
	assert.Equal(t, 1, s.GetNumberOfChats())

	// try to get by id
	c, err := s.getChat(id.String(), false)
	assert.Nil(t, err)
	assert.Equal(t, name, c.Name)

	// try to get by index
	c, err = s.GetChatByIndex(0)
	assert.Nil(t, err)
	assert.Equal(t, name, c.Name)

	// add a chat already in storage
	id, err = s.AddNewChat(name)
	assert.True(t, errors.Is(err, AlreadyInListWithNameErr))
}

func Test_storage_AddChat(t *testing.T) {
	s := NewStorage()
	name := "my-chat"
	chat := crdt.NewChat(name)
	idString := chat.Id.String()

	err := s.AddChat(chat)
	assert.Nil(t, err)
	assert.Equal(t, 1, s.GetNumberOfChats())

	// try to get by idString
	c, err := s.getChat(idString, false)
	assert.Nil(t, err)
	assert.Equal(t, name, c.Name)

	// try to get by index
	c, err = s.GetChatByIndex(0)
	assert.Nil(t, err)
	assert.Equal(t, name, c.Name)

	// add a chat already in storage with name
	newChat := crdt.NewChat(name)
	err = s.AddChat(newChat)
	assert.True(t, errors.Is(err, AlreadyInListWithNameErr))

	// add a chat already in storage with same id
	newChat.Name = "different name"
	newChat.Id = chat.Id
	err = s.AddChat(newChat)
	assert.True(t, errors.Is(err, AlreadyInListWithIDErr))
}

func Test_storage_RemoveChat(t *testing.T) {
	s := NewStorage()
	name := "my-chat"
	chat := crdt.NewChat(name)
	idString := chat.Id.String()
	err := s.AddChat(chat)
	assert.Nil(t, err)

	assert.Equal(t, 1, s.GetNumberOfChats())

	id, err := uuid.Parse(idString)
	assert.Nil(t, err)

	s.RemoveChat(id)
	assert.Equal(t, 0, s.GetNumberOfChats())
}

func Test_storage_getChat(t *testing.T) {
	chat := crdt.NewChat("chat name")
	s := NewStorage()

	err := s.AddChat(chat)
	assert.Nil(t, err)

	gotFromStorage, err := s.getChat(chat.Id.String(), false)
	assert.Nil(t, err)
	assert.Equal(t, chat, gotFromStorage)

	gotFromStorage, err = s.getChat(chat.Name, true)
	assert.Nil(t, err)
	assert.Equal(t, chat, gotFromStorage)
}

func Test_storage_AddNodeToChat(t *testing.T) {
	s := NewStorage()
	chatName := "my-chat"
	id, err := s.AddNewChat(chatName)
	assert.Nil(t, err)

	c, err := s.getChat(id.String(), false)
	assert.Nil(t, err)

	addr := "127.0.0.1"
	port := "8080"
	nodeName := "toto"
	slot := uint8(1)
	node := crdt.NewNodeInfos(addr, port, nodeName)
	node.Slot = slot

	numberOfSlots := c.GetSlots()
	assert.Equal(t, 0, len(numberOfSlots))
	err = s.AddNodeToChat(node, id)
	assert.Nil(t, err)

	assert.Nil(t, err)
	numberOfSlots = c.GetSlots()
	assert.Equal(t, 1, len(numberOfSlots))

	// Verify node
	n, err := c.GetNodeBySlot(slot)
	assert.Nil(t, err)
	assert.Equal(t, n.Name, nodeName)
	assert.Equal(t, n.Port, port)
	assert.Equal(t, n.Address, addr)
	assert.Equal(t, n.Slot, slot)

	// Try to get un existent node
	n, err = c.GetNodeBySlot(uint8(10))
	assert.Equal(t, err.Error(), NotFoundErr.Error())
}

func Test_storage_RemoveNodeFromChat(t *testing.T) {
	s := NewStorage()
	name := "my-chat"
	id, err := s.AddNewChat(name)
	assert.Nil(t, err)

	c, err := s.getChat(id.String(), false)
	assert.Nil(t, err)

	node := crdt.NewNodeInfos("127.0.0.1", "8080", "toto")
	// Setting slot to identify active TCP connection
	node.Slot = 1
	err = s.AddNodeToChat(node, id)
	assert.Nil(t, err)

	numberOfSlots := c.GetSlots()
	assert.Equal(t, 1, len(numberOfSlots))

	chatId, err := uuid.Parse(c.Id.String())
	assert.Nil(t, err)
	err = s.RemoveNodeFromChat(node.Slot, chatId)
	assert.Nil(t, err)

	// try to remove in existent node slot
	err = s.RemoveNodeFromChat(node.Slot, chatId)
	assert.NotNil(t, err)
}

func Test_storage_GetNumberOfChats(t *testing.T) {
	s := NewStorage()
	_, err := s.AddNewChat("chat name")
	assert.Nil(t, err)

	assert.Equal(t, 1, s.GetNumberOfChats())
}

func TestStorage_RemoveNodeSlotFromStorage(t *testing.T) {
	s := NewStorage()

	first, err := s.AddNewChat("first")
	assert.Nil(t, err)
	second, err := s.AddNewChat("second")
	assert.Nil(t, err)

	firstNode := crdt.NewNodeInfos("", "", "first")
	firstNode.Slot = uint8(1)

	secondNode := crdt.NewNodeInfos("", "", "second")
	secondNode.Slot = uint8(2)

	err = s.AddNodeToChat(firstNode, first)
	assert.Nil(t, err)

	err = s.AddNodeToChat(secondNode, first)

	c1, err := s.getChat(first.String(), false)
	assert.Nil(t, err)

	_, err = c1.GetNodeBySlot(uint8(1))
	assert.Nil(t, err)

	_, err = c1.GetNodeBySlot(uint8(2))
	assert.Nil(t, err)

	s.RemoveNodeSlotFromStorage(2)
	_, err = c1.GetNodeBySlot(uint8(2))
	assert.Equal(t, err.Error(), NotFoundErr.Error())

	err = s.AddNodeToChat(secondNode, first)
	assert.Nil(t, err)
	err = s.AddNodeToChat(secondNode, second)
	assert.Nil(t, err)

	c2, err := s.getChat(second.String(), false)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(c1.GetSlots()))
	assert.Equal(t, 1, len(c2.GetSlots()))

	s.RemoveNodeSlotFromStorage(2)
	assert.Equal(t, 1, len(c1.GetSlots()))
	assert.Equal(t, 0, len(c2.GetSlots()))
}
