package crdt

type (
	operationType int
	targetType    int
	operation     [1000]byte

	Operation interface {
		getOperationType() operationType
	}
)

const (
	add    operationType = iota
	remove operationType = iota
	update operationType = iota
)

const (
	messageType targetType = iota
	nodeType
)

func (b *operation) setOperation(op operationType) {
	b[0] = byte(op)
}

func (b *operation) getOperationType() operationType {
	return operationType(b[0])
}

/*func GetOperationFromBytes(bytes []byte) operation {
	_ := int(bytes[0])
	target := int(bytes[1])
	switch target {
	case messageType:
		m := GetMessageFromBytes(bytes[2:])
	}

}*/
