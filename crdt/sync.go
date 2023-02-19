package crdt

type (
	operation struct {
		slot         uint8
		typology     OperationType
		targetedChat string
		data         []byte
	}

	OperationType uint8

	Operation interface {
		GetOperationType() OperationType
		GetTargetedChat() string
		GetOperationData() []byte
		SetSlot(slot uint8)
		ToBytes() []byte
	}
)

const (
	CreateChat     OperationType = iota
	JoinChatByName OperationType = iota
	AddNode        OperationType = iota
	AddMessage     OperationType = iota
	LeaveChat      OperationType = iota
	ListUsers      OperationType = iota
	Quit           OperationType = iota
)

func NewOperation(typology OperationType, targetedChat string, data []byte) Operation {
	return &operation{
		slot:         0,
		typology:     typology,
		targetedChat: targetedChat,
		data:         data,
	}
}

func (op *operation) GetOperationType() OperationType {
	return op.typology
}

func (op *operation) GetTargetedChat() string {
	return op.targetedChat
}

func (op *operation) GetOperationData() []byte {
	return op.data
}

func (op *operation) SetSlot(slot uint8) {
	op.slot = slot
}

func (op *operation) ToBytes() []byte {
	var bytes []byte
	bytes = append(bytes, op.slot)
	bytes = append(bytes, uint8(len(op.targetedChat)))
	bytes = append(bytes, []byte(op.targetedChat)...)
	bytes = append(bytes, uint8(op.typology))
	bytes = append(bytes, uint8(len(op.data)))
	bytes = append(bytes, op.data...)
	return bytes
}

func DecodeOperation(bytes []byte) Operation {
	slot := bytes[0]
	offset, targetedChat := getField(1, bytes)
	typology := OperationType(bytes[offset])
	_, data := getField(offset+1, bytes)

	return &operation{
		slot:         slot,
		typology:     typology,
		targetedChat: string(targetedChat),
		data:         data,
	}
}

func getField(offset int, source []byte) (int, []byte) {
	lenField := int(source[offset])
	return offset + lenField + 1, source[offset+1 : offset+lenField+1]
}
