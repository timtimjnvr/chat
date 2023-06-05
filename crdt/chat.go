package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"time"
)

type (
	Chat struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		nodesInfos []*NodeInfos
		messages   []*Message // ordered by date : 0 being the oldest message, 1 coming after 0 etc ...
	}
)

const maxNumberOfMessages, maxNumberOfNodes = 100, 100

func NewChat(name string) *Chat {
	return &Chat{
		Id:         uuid.New().String(),
		Name:       name,
		nodesInfos: make([]*NodeInfos, 0, maxNumberOfNodes),
		messages:   make([]*Message, 0, maxNumberOfMessages),
	}
}

func (c *Chat) SaveNode(nodeInfo *NodeInfos) {
	// update if found
	for i, n := range c.nodesInfos {
		if n.Id == nodeInfo.Id {
			c.nodesInfos[i] = nodeInfo
			return
		}
	}

	// append if not found
	c.nodesInfos = append(c.nodesInfos, nodeInfo)
}

func (c *Chat) RemoveNodeBySlot(slot uint8) {
	// get index
	var (
		index int
		n     NodeInfos
	)
	for index, n = range c.nodesInfos {
		if n.Slot == slot {
			break
		}
	}

	// node identified by slot found
	if index != len(c.nodesInfos) {
		c.nodesInfos = append(c.nodesInfos[:index-1], c.nodesInfos[index+1])
	}
}

func (c *Chat) SaveMessage(message *Message) {
	if len(c.messages) == 0 {
		c.messages = append(c.messages, message)
		return
	}

	messageToSaveDate, _ := time.Parse(time.RFC3339, message.Date)
	if !c.ContainsMessage(message) {
		var (
			i              int
			messageDate, _ = time.Parse(time.RFC3339, c.messages[i].Date)
		)

		for messageDate, _ = time.Parse(time.RFC3339, c.messages[i].Date); messageToSaveDate.After(messageDate) && i < len(c.messages); i++ {
		}

		beginning := c.messages[:i]
		end := c.messages[i:]
		newMessages := make([]*Message, len(c.messages)+1)

		j := 0
		if len(beginning) > 0 {
			var m *Message
			for j, m = range beginning {
				newMessages[j] = m
			}

			j += 1
		}

		newMessages[j] = message

		if len(end) > 0 {
			for k, m := range end {
				newMessages[k+j+1] = m
			}
		}

		c.messages = newMessages
	}
}

func (c *Chat) ToBytes() []byte {
	bytesChat, _ := json.Marshal(c)
	return bytesChat
}

func (c *Chat) GetSlots(myId uuid.UUID) []uint8 {
	slots := make([]uint8, 0, len(c.nodesInfos))
	for _, i := range c.nodesInfos {
		if i.Id != myId {
			slots = append(slots, i.Slot)
		}
	}

	return slots
}

func (c *Chat) DisplayUsers() {
	log.Printf("chat name : %s\n", c.Name)
	for _, n := range c.nodesInfos {
		log.Printf("- %s (Address: %s, Port: %s, Slot: %d)\n", n.Name, n.Address, n.Port, n.Slot)
	}
}

func (c *Chat) ContainsMessage(message *Message) bool {
	for _, m := range c.messages {
		if m.Id == message.Id {
			return true
		}
	}
	return false
}

func (c *Chat) GetMessageOperationsForPropagation() []*Operation {
	addMessageOperations := make([]*Operation, 0, 0)
	for _, m := range c.messages {
		addMessageOperations = append(addMessageOperations, NewOperation(AddMessage, c.Id, m))
	}
	return addMessageOperations
}

func (c *Chat) containsNode(node *NodeInfos) bool {
	for _, n := range c.nodesInfos {
		if n.Id == node.Id {
			return true
		}
	}

	return false
}
