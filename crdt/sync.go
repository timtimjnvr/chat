package crdt

type (
	operationType int
	targetType    int
	operation     [1000]byte

	Operable interface {
		toRunes(operation operationType) []rune
	}
)

/*
ADD A MESSAGE
operationType targetType message
*/

const (
	add    operationType = iota
	remove operationType = iota
	update operationType = iota
)

const (
	messageType targetType = iota
	nodeType
)
