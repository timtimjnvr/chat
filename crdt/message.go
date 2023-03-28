package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type (
	Message struct {
		Id      uuid.UUID `json:"Id"`
		Sender  string    `json:"sender"`
		Content string    `json:"content"`
		Date    time.Time `json:"date"`
	}
)

func NewMessage(sender, content string) *Message {
	id, _ := uuid.NewUUID()
	return &Message{
		Id:      id,
		Sender:  sender,
		Content: content,
		Date:    time.Now().UTC(),
	}
}

func (m *Message) ToBytes() []byte {
	bytesMessage, _ := json.Marshal(m)
	return bytesMessage
}
