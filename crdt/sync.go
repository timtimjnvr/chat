package crdt

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
		getOperationData() []byte
		ToBytes() []byte
	}
)

const (
	CreateChat     = iota
	JoinChatByName = iota
	AddNode        = iota
	AddMessage     = iota
	LeaveChat      = iota
	ListUsers      = iota
	Quit           = iota
)

func NewOperation(typology OperationType, targetedChat string, data []byte) operation {
	return operation{
		typology:     typology,
		targetedChat: targetedChat,
		data:         data,
	}
}

func (op operation) ToBytes() []byte {
	var bytes []byte

	bytes = append(bytes, uint8(op.typology))
	bytes = append(bytes, []byte(op.targetedChat)...)
	bytes = append(bytes, op.data...)

	return bytes
}

func decodeOperation(bytes []byte) (operation, error) {
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

func (op operation) getTargetedChat() string {
	return op.targetedChat
}

func (op operation) getOperationData() []byte {
	return op.data
}
