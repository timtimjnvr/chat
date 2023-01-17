package crdt

import (
	"github.com/google/uuid"
	"time"
)

type (
	message struct {
		id      uuid.UUID `json:"id"`
		sender  string    `json:"sender"`
		content string    `json:"content"`
		date    time.Time `json:"date"`
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
		sender:  sender,
		content: content,
		date:    time.Now(),
	}
}

func (m *message) GetId() uuid.UUID {
	return m.id
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
