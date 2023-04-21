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
		Messages   []*Message // ordered by date : 0 being the oldest message, 1 coming after 0 etc ...
	}
)

const maxNumberOfMessages, maxNumberOfNodes = 100, 100

func NewChat(name string) *Chat {
	return &Chat{
		Id:         uuid.New().String(),
		Name:       name,
		nodesInfos: make([]*NodeInfos, 0, maxNumberOfNodes),
		Messages:   make([]*Message, 0, maxNumberOfMessages),
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
	if len(c.Messages) == 0 {
		c.Messages = append(c.Messages, message)
		return
	}

	messageToSaveDate, _ := time.Parse(time.RFC3339, message.Date)
	if !c.ContainsMessage(message) {
		var (
			i              int
			messageDate, _ = time.Parse(time.RFC3339, c.Messages[i].Date)
		)

		for messageDate, _ = time.Parse(time.RFC3339, c.Messages[i].Date); messageToSaveDate.After(messageDate) && i < len(c.Messages); i++ {
		}

		beginning := c.Messages[:i]
		end := c.Messages[i:]
		newMessages := make([]*Message, len(c.Messages)+1)

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

		c.Messages = newMessages
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
	for _, m := range c.Messages {
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
