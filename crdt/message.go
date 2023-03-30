package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type (
	Message struct {
		Id      uuid.UUID `json:"id"`
		Sender  string    `json:"sender"`
		Content string    `json:"content"`
		Date    string    `json:"date"`
	}
)

func NewMessage(sender, content string) *Message {
	id, _ := uuid.NewUUID()
	return &Message{
		Id:      id,
		Sender:  sender,
		Content: content,
		Date:    time.Now().Format(time.RFC3339),
	}
}

func (m *Message) ToBytes() []byte {
	bytesMessage, _ := json.Marshal(m)
	return bytesMessage
}
