package crdt

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type (
	Chat struct {
		Id         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		nodesSlots []uint8
		messages   []*Message // ordered by date : 0 being the oldest message, 1 coming after 0 etc ...
	}
)

var NotFoundErr = errors.New("not found")

const maxNumberOfMessages, maxNumberOfNodes = 100, 100

func NewChat(name string) *Chat {
	return &Chat{
		Id:         uuid.New(),
		Name:       name,
		nodesSlots: make([]uint8, 0, maxNumberOfNodes),
		messages:   make([]*Message, 0, maxNumberOfMessages),
	}
}

func (c *Chat) GetID() uuid.UUID {
	return c.Id
}

func (c *Chat) GetName() string {
	return c.Name
}

func (c *Chat) SaveNode(nodeSlot uint8) {

	// check if present
	for _, s := range c.nodesSlots {
		if s == nodeSlot {
			return
		}
	}

	// append if not found
	c.nodesSlots = append(c.nodesSlots, nodeSlot)
}

func (c *Chat) RemoveNode(slot uint8) error {
	// get index
	var (
		index int
		found bool
		s     uint8
	)

	for index, s = range c.nodesSlots {
		if s == slot {
			found = true
			break
		}
	}

	if !found {
		return NotFoundErr
	}

	if index == 0 && len(c.nodesSlots) == 1 {
		c.nodesSlots = make([]uint8, 0, 0)
		return nil
	}

	if index == 0 && len(c.nodesSlots) > 1 {
		c.nodesSlots = c.nodesSlots[index+1:]
		return nil
	}

	if index == len(c.nodesSlots)-1 {
		c.nodesSlots = c.nodesSlots[:len(c.nodesSlots)-1]
		return nil
	}
	var (
		newNodesSlots = make([]uint8, len(c.nodesSlots)-1)
		j             int
	)
	for i := 0; i <= len(c.nodesSlots)-1; i++ {
		if i == index {
			continue
		}

		newNodesSlots[j] = c.nodesSlots[i]
		j++
	}

	c.nodesSlots = newNodesSlots
	return nil
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

// GetSlots returns all the slots identifying active TCP connections between nodes.
func (c *Chat) GetSlots() []uint8 {
	length := 0
	if len(c.nodesSlots) > 0 {
		length = len(c.nodesSlots) - 1
	}

	slots := make([]uint8, 0, length)
	for _, s := range c.nodesSlots {
		// My own slot
		if s == 0 {
			continue
		}

		slots = append(slots, s)
	}

	return slots
}

func (c *Chat) ContainsMessage(message *Message) bool {
	for _, m := range c.messages {
		if m.Id == message.Id {
			return true
		}
	}
	return false
}

func (c *Chat) Display() {
	fmt.Printf("- %s : %d users, %d messages\n", c.Name, len(c.nodesSlots)+1, len(c.messages))
}
