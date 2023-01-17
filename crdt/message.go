package crdt

import (
	"github.com/google/uuid"
	"log"
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
		UpdateContent(content string)
		GetSyncBytes(operation operationType) []byte
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

func (m *message) UpdateContent(content string) {
	m.Content = content
}

func (m *message) OperationToBytes(typology operationType) {
	b := new(operation)
	b.setOperation(typology)
}

func (m *message) GetSyncBytes(operation operationType) []byte {
	// add operation
	var (
		operationByte = byte(operation)
		target        = byte(messageType)
		idBytes       = []byte(m.GetId().String())
		senderBytes   = []byte(m.GetSender())
		contentBytes  = []byte(m.GetContent())
		dateByte      = []byte(m.GetDate())
	)

	addBytes := func(destination []byte, source ...byte) []byte {
		lenBytes := byte(len(source))
		destination = append(destination, lenBytes)
		destination = append(destination, source[0:]...)
		return destination
	}

	bytes := make([]byte, 2)
	bytes[0] = operationByte
	bytes[1] = target
	bytes = addBytes(bytes, idBytes...)
	bytes = addBytes(bytes, senderBytes...)
	bytes = addBytes(bytes, contentBytes...)
	bytes = addBytes(bytes, dateByte...)

	return bytes
}

func GetMessageFromBytes(bytes []byte) (m message) {
	var (
		idBytes      []byte
		senderBytes  []byte
		contentBytes []byte
		dateByte     []byte
	)

	getField := func(offest int, source []byte) (int, []byte) {
		lenField := int(source[offest])
		return offest + lenField, source[offest : offest+lenField]
	}

	offset := 0
	offset, idBytes = getField(offset, bytes)
	log.Println(offset, string(idBytes))
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
