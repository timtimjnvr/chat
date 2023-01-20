package crdt

type (
	operationType int32

	// message or node
	Operable interface {
		ToRunes() []rune
		SendNodes(content []byte)
	}
)

const (
	AddMessage    operationType = iota
	removeMessage operationType = iota
	updateMessage operationType = iota
)

/*
ADD A MESSAGE
operationType message
*/

func toRunes(operation operationType) []rune {
	return []rune{int32(operation)}
}

func GetOperationRunes(operation operationType, data Operable) []rune {
	return append(toRunes(operation), data.ToRunes()...)
}
