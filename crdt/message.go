package crdt

import (
	"github.com/google/uuid"
	"time"
)

type (
	message struct {
		id      uuid.UUID `json:"id"`
		Sender  string    `json:"sender"`
		Content string    `json:"content"`
		Date    time.Time `json:"date"`
	}

	Message interface {
		GetId() uuid.UUID
		GetSender() string
		GetContent() string
		UpdateContent(content string)
	}
)

func NewMessage(sender, content string) Message {
	return &message{
		Sender:  sender,
		Content: content,
		Date:    time.Now(),
	}
}

func (m *message) GetId() uuid.UUID {
	return m.id
}

func (m *message) GetSender() string {
	return m.Sender
}

func (m *message) GetContent() string {
	return m.Content
}

func (m *message) UpdateContent(content string) {
	m.Content = content
}
