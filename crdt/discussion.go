package crdt

import "chat/node"

type (
	chat struct {
		name     string
		nodes    []node.Node
		messages []message
	}

	Chat interface {
		AddNode()
		AddMessage()
		UpdateMessages()
	}
)

func NewChat(name string) Chat {
	return &chat{
		name:     name,
		nodes:    []node.Node{},
		messages: []message{},
	}
}

// TODO

func (c *chat) AddNode()        {}
func (c *chat) AddMessage()     {}
func (c *chat) UpdateMessages() {}
