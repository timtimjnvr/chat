package crdt

import (
	"chat/node"
	"encoding/json"
	"log"
)

type (
	messageOperation struct {
		OperationTypology operationType `json:"operation"`
		Data              []message     `json:"data"`
	}

	nodeOperation struct {
		OperationTypology operationType `json:"operation"`
		Data              []node.Infos  `json:"data"`
	}

	operationType int

	Operation interface {
		getOperationType() operationType
	}
)

const (
	add    operationType = iota
	remove operationType = iota
	update operationType = iota
)

func deserializeOperation(b []byte) Operation {
	var msgOp messageOperation
	var nodeOp nodeOperation
	var bytesCopy = b

	// try to unmarshall into a message operation
	err := json.Unmarshal(b, &msgOp)
	if err == nil {
		return msgOp
	}

	log.Println(err)

	// try to unmarshall into a node operation
	err = json.Unmarshal(bytesCopy, &nodeOp)
	if err != nil {
		log.Println(err)
	}
	return nodeOp
}

func serializeOperation(operation interface{}) []byte {
	output, err := json.Marshal(operation)
	if err != nil {
		log.Println(err)
	}
	return output
}

func (op nodeOperation) getOperationType() operationType {
	return op.OperationTypology
}

func (op messageOperation) getOperationType() operationType {
	return op.OperationTypology
}
