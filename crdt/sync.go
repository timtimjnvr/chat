package crdt

import "github.com/google/uuid"

type (
	operation struct {
		typology     OperationType
		targetedChat string
		data         []byte
	}

	OperationType int32

	Operable interface {
		ToRunes() []rune
	}

	Operation interface {
		GetOperationType() OperationType
		GetTargetedChat() uuid.UUID
		GetOperationData() []byte
		ToBytes() []byte
	}
)

const (
	AddChat = iota
	JoinChatByName               = iota
	LeaveChat                    = iota
	AddMessage     OperationType = iota
	AddNode                      = iota
)

func NewOperation(typology OperationType, targetedChat string, data []byte) operation {
	return operation{
		typology:     typology,
		targetedChat: targetedChat,
		data:         data,
	}
}

func (op operation) GetOperationType() OperationType {
	return op.typology
}

func (op operation) GetOperationData() []byte {
	return op.data
}

func (op operation) GetTargetedChat() string {
	return op.targetedChat
}

func (op operation) ToBytes() []byte {
	var bytes []byte

	bytes = append(bytes, uint8(op.typology))
	bytes = append(bytes, []byte(op.targetedChat)...)
	bytes = append(bytes, op.data...)

	return bytes
}

func DecodeOperation(bytes []byte) (operation, error) {
	getField := func(offset int, source []byte) (int, []byte) {
		lenField := int(source[offset])
		return offset + lenField + 1, source[offset+1 : offset+lenField+1]
	}

	var (
		offset             = 0
		data, targetedChat []byte
	)

	typology := OperationType(bytes[offset])
	offset, targetedChat = getField(offset, bytes)
	_, data = getField(offset+1, bytes)

	return operation{
		typology:     typology,
		targetedChat: string(targetedChat),
		data:         data,
	}, nil
}
