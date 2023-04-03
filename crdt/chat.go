package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type (
	Chat struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		nodesInfos []*NodeInfos
		messages   []*Message
	}
)

const maxNumberOfMessages, maxNumberOfNodes = 100, 100

func NewChat(id uuid.UUID, name string) *Chat {
	return &Chat{
		Id:         id.String(),
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

func (c *Chat) SaveMessage(message *Message) {
	if len(c.messages) == 0 {
		c.messages = append(c.messages, message)
		return
	}

	date, _ := time.Parse(time.RFC3339, message.Date)
	if !c.containsMessage(message) {
		var (
			i          int
			tmpDate, _ = time.Parse(time.RFC3339, c.messages[i].Date)
		)

		for date.Before(tmpDate) {
			i++
		}

		beginning := c.messages[:i-1]
		end := c.messages[i:]
		c.messages = append(beginning, message)
		c.messages = append(c.messages, end...)
	}
}

func (m *Chat) ToBytes() []byte {
	bytesChat, _ := json.Marshal(m)
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

func (c *Chat) containsMessage(message *Message) bool {
	for _, m := range c.messages {
		if m.Id == message.Id {
			return true
		}
	}
	return false
}

func (c *Chat) containsNode(node *NodeInfos) bool {
	for _, n := range c.nodesInfos {
		if n.Id == node.Id {
			return true
		}
	}

	return false
}
