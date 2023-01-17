package crdt

import "chat/node"

type (
	chat struct {
		name     string
		nodes    []*node.Node
		messages []Message
	}

	Chat interface {
		AddNode(node *node.Node)
		AddMessage(message Message)
		UpdateMessages(messages []Message)
		GetMessages() []Message
	}
)

func NewChat(name string) Chat {
	return &chat{
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
		c.messages = append(c.messages, message)
	}
}
func (c *chat) UpdateMessages(messages []Message) {

}

func (c *chat) GetMessages() []Message {
	return c.messages
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
