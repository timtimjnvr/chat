package crdt

import "time"

type (
	message struct {
		sender, content string
		date            time.Time
	}

	Message interface {
		GetSender() string
		GetContent() string
		UpdateContent(content string)
	}
)

func NewMessage(sender, content string) Message {
	return &message{
		sender:  sender,
		content: content,
		date:    time.Now(),
	}
}

func (m *message) GetSender() string {
	return m.sender
}

func (m *message) GetContent() string {
	return m.content
}

func (m *message) UpdateContent(content string) {
	m.content = content
}
