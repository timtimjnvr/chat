package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type (
	message struct {
		Id      uuid.UUID `json:"Id"`
		Sender  string    `json:"sender"`
		Content string    `json:"content"`
		Date    time.Time `json:"date"`
	}

	Message interface {
		GetId() uuid.UUID
		GetSender() string
		GetContent() string
		GetDate() string
		ToBytes() []byte
	}
)

func NewMessage(sender, content string) Message {
	return &message{
		Sender:  sender,
		Content: content,
		Date:    time.Now().UTC(),
	}
}

func (m message) GetId() uuid.UUID {
	return m.Id
}

func (m message) GetSender() string {
	return m.Sender
}

func (m message) GetContent() string {
	return m.Content
}

func (m message) GetDate() string {
	return m.Date.Format(time.RFC3339)
}

func (m message) ToBytes() []byte {
	bytesMessage, _ := json.Marshal(m)
	return bytesMessage
}

func DecodeMessage(bytes []byte) (Message, error) {
	var m message
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
