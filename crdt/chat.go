package crdt

import (
	"chat/conn"
	"chat/node"
	"github.com/google/uuid"
)

type (
	chat struct {
		id       uuid.UUID
		myNodeId uuid.UUID
		name     string
		nodes    []*node.Node
		messages []Message
	}

	Chat interface {
		AddNode(node *node.Node)
		AddMessage(message Message)
		Send(data []byte)
	}
)

func NewChat(name string) Chat {
	id, _ := uuid.NewUUID()
	return &chat{
		id:       id,
		name:     name,
		nodes:    []*node.Node{},
		messages: []Message{},
	}
}

func (c *chat) AddNode(node *node.Node) {
	if !c.containsNode(node) {
		c.nodes = append(c.nodes, node)
	}
}

func (c *chat) AddMessage(message Message) {
	if !c.containsMessage(message) {
		// TODO : insert message in array by comparing dates
		c.messages = append(c.messages, message)
	}
}

func (c *chat) containsNode(node *node.Node) bool {
	for _, n := range c.nodes {
		if n.Infos.Id == node.Infos.Id {
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

func (c *chat) Send(data []byte) {
	for _, node := range c.nodes {
		if c.myNodeId == node.Infos.Id {
			continue
		}

		conn.Send(node.Business.Conn, data)
	}
}
