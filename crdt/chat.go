package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
)

type (
	ConcreteChat struct {
		Id       string `json:"Id"`
		myNodeId uuid.UUID
		Name     string `json:"name"`
		nodes    []Infos
		messages []Message
	}

	Chat interface {
		GetId() string
		GetName() string
		GetNodesInfos() []Infos
		GetSlots() []uint8
		AddNode(infos Infos)
		AddMessage(message Message)
		ToBytes() ([]byte, error)
	}
)

const maxNumberOfMessages, maxNumberOfNodes = 100, 100

func NewChat(name string) Chat {
	id, _ := uuid.NewUUID()
	return &ConcreteChat{
		Id:       id.String(),
		Name:     name,
		nodes:    make([]Infos, 0, maxNumberOfNodes),
		messages: make([]Message, 0, maxNumberOfMessages),
	}
}

func (c *ConcreteChat) GetNodesInfos() []Infos {
	return c.nodes
}

func (c *ConcreteChat) GetId() string {
	return c.Id
}

func (c *ConcreteChat) GetName() string {
	return c.Name
}

func (c *ConcreteChat) AddNode(i Infos) {
	log.Println(i)
	// c.nodes = append(c.nodes, i)
}

func (c *ConcreteChat) AddMessage(message Message) {
	if !c.containsMessage(message) {
		// TODO : insert message in array by comparing dates
		c.messages = append(c.messages, message)
	}
}

func (c *ConcreteChat) containsMessage(message Message) bool {
	for _, m := range c.messages {
		if m.GetId() == message.GetId() {
			return true
		}
	}
	return false
}

func (c *ConcreteChat) ToBytes() ([]byte, error) {
	bytesChat, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return bytesChat, nil
}

func (c *ConcreteChat) GetSlots() []uint8 {
	slots := make([]uint8, 0, len(c.nodes))
	for _, i := range c.nodes {
		if i.getId() == c.myNodeId {
			slots = append(slots, i.getSlot())
		}
	}

	return slots
}
