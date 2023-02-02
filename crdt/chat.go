package crdt

import (
	"chat/node"
	"github.com/google/uuid"
)

type (
	chat struct {
		id       uuid.UUID
		myNodeId uuid.UUID
		name     string
		nodes    []*node.Infos
		messages []Message
	}

	Chat interface {
		GetId() uuid.UUID
		GetName() string
		GetNodesInfos() []*node.Infos
		AddNode(infos *node.Infos)
		AddMessage(message Message)
	}
)

func NewChat(name string) Chat {
	id, _ := uuid.NewUUID()
	return &chat{
		id:       id,
		name:     name,
		nodes:    []*node.Infos{},
		messages: []Message{},
	}
}

func (c *chat) GetNodesInfos() []*node.Infos {
	return c.nodes
}

func (c *chat) GetId() uuid.UUID {
	return c.id
}

func (c *chat) GetName() string {
	return c.name
}

func (c *chat) AddNode(infos *node.Infos) {
	if !c.containsNode(infos.Id) {
		c.nodes = append(c.nodes, infos)
	}
}

func (c *chat) AddNodeInfos(i *node.Infos) {
	if !c.containsNode(i.Id) {
		c.nodes = append(c.nodes, i)
	}
}

func (c *chat) AddMessage(message Message) {
	if !c.containsMessage(message) {
		// TODO : insert message in array by comparing dates
		c.messages = append(c.messages, message)
	}
}

func (c *chat) containsNode(id uuid.UUID) bool {
	for _, n := range c.nodes {
		if n.Id == id {
			return true
		}
	}
	return false
}

func (c *chat) containsMessage(message Message) bool {
	for _, m := range c.messages {
		if m.GetId() == message.GetId() {
			return true
		}
	}
	return false
}
