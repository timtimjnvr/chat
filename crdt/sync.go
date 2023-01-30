package crdt

type (
	operation struct {
		typology operationType
		data     []rune
	}

	operationType int32

	Operable interface {
		ToRunes() []rune
	}
)

const (
	AddMessage    operationType = iota
	removeMessage operationType = iota
	updateMessage operationType = iota
)

func NewOperation(typology operationType, data []rune) operation {
	return operation{
		typology: typology,
		data:     data,
	}
}

func (op *operation) GetOperationType() operationType {
	return op.typology
}

func (op *operation) GetOperationData() []rune {
	return op.data
}

func (op *operation) SetOperationType(typology operationType) {
	op.typology = typology
}

func (op *operation) SetOperationData(data Operable) {
	op.data = data.ToRunes()
}

func DecodeOperation(bytes []rune) Operable {
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

	return &operation{
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
