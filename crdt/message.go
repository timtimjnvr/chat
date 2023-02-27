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
		GetTime() time.Time
		ToBytes() []byte
	}
)

func NewMessage(sender, content string) Message {
	id, _ := uuid.NewUUID()
	return &message{
		Id:      id,
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

func (m message) GetTime() time.Time {
	return m.Date
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
