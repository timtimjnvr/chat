package crdt

import "github.com/google/uuid"

type (
	operation struct {
		typology     operationType
		targetedChat uuid.UUID
		data         []rune
	}

	operationType int32

	Operable interface {
		ToRunes() []rune
	}
)

const (
	JoinChat                    = iota
	LeaveChat                   = iota
	AddMessage    operationType = iota
	RemoveMessage operationType = iota
	UpdateMessage operationType = iota
)

func NewOperation(typology operationType, targetedChat uuid.UUID, data []rune) operation {
	return operation{
		typology:     typology,
		targetedChat: targetedChat,
		data:         data,
	}
}

func (op *operation) GetOperationType() operationType {
	return op.typology
}

func (op *operation) GetOperationData() []rune {
	return op.data
}

func (op *operation) GetTargetedChat() uuid.UUID {
	return op.targetedChat
}

func (op *operation) SetOperationType(typology operationType) {
	op.typology = typology
}

func (op *operation) SetOperationData(data Operable) {
	op.data = data.ToRunes()
}

func DecodeOperation(bytes []rune) operation {
	getField := func(offset int, source []rune) (int, []rune) {
		lenField := int(source[offset])
		return offset + lenField + 1, source[offset+1 : offset+lenField+1]
	}
	var (
		offset = 0
		data   []rune
	)

	typology := operationType(bytes[offset])
	_, data = getField(offset+1, bytes)

	return operation{
		typology: typology,
		data:     data,
	}
}

func (op *operation) ToRunes() []rune {
	var bytes []rune

	// 1st add operation id
	bytes = append(bytes, int32(op.typology))

	//2nd add operation data
	bytes = append(bytes, op.data...)

	return bytes
}
