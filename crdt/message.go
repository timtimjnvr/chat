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
		GetDate() string
		ToRunes() []rune
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

func (m *message) GetDate() string {
	return m.Date.Format(time.RFC3339)
}

func (m *message) ToRunes() []rune {
	var (
		idBytes      = []rune(m.GetId().String())
		senderBytes  = []rune(m.GetSender())
		contentBytes = []rune(m.GetContent())
		dateByte     = []rune(m.GetDate())
		bytes        []rune
	)

	addBytes := func(destination []rune, source ...rune) []rune {
		lenBytes := []rune{int32(len(source))}
		destination = append(destination, lenBytes...)
		destination = append(destination, source...)

		return destination
	}

	bytes = addBytes(bytes, idBytes...)
	bytes = addBytes(bytes, senderBytes...)
	bytes = addBytes(bytes, contentBytes...)
	bytes = addBytes(bytes, dateByte...)

	return bytes
}

func DecodeMessage(bytes []rune) (m message) {
	var (
		idBytes      []rune
		senderBytes  []rune
		contentBytes []rune
		dateByte     []rune
	)

	getField := func(offset int, source []rune) (int, []rune) {
		lenField := int(source[offset])
		return offset + lenField + 1, source[offset+1 : offset+lenField+1]
	}

	offset := 0
	offset, idBytes = getField(offset, bytes)
	offset, senderBytes = getField(offset, bytes)
	offset, contentBytes = getField(offset, bytes)
	_, dateByte = getField(offset, bytes)

	id, _ := uuid.Parse(string(idBytes))
	date, _ := time.Parse(time.RFC3339, string(dateByte))

	return message{
		id:      id,
		Sender:  string(senderBytes),
		Content: string(contentBytes),
		Date:    date,
	}
}
