package crdt

import (
	"encoding/json"
	"github/timtimjnvr/chat/reader"
)

type (
	Operation struct {
		Slot         uint8 // Slot of the node who forwarded the operation
		Typology     OperationType
		TargetedChat string // uuid or chat name
		Data         Data
	}

	OperationType uint8

	Data interface {
		ToBytes() []byte
	}
)

const (
	CreateChat OperationType = iota
	JoinChatByName
	SaveNode
	KillNode
	AddNode
	RemoveNode
	AddChat
	RemoveChat
	SwitchChat
	AddMessage
	ListChatUsers
	ListUsers
	ListChats
	Quit
)

var operationNames = map[OperationType]string{
	CreateChat:     "create chat",
	JoinChatByName: "join chat by name",
	SaveNode:       "save node",
	KillNode:       "kill node",
	AddNode:        "add node",
	RemoveNode:     "remove node",
	AddChat:        "add chat",
	RemoveChat:     "leave chat",
	SwitchChat:     "switch chat",
	AddMessage:     "add message",
	ListChatUsers:  "list chat users",
	ListUsers:      "list users",
	ListChats:      "list chats",
	Quit:           "quit",
}

func NewOperation(typology OperationType, targetedChat string, data Data) *Operation {
	return &Operation{
		Slot:         0,
		Typology:     typology,
		TargetedChat: targetedChat,
		Data:         data,
	}
}

// Slot :
// identifies the TCP connection slot
//
// TargetedChat :
// uuid of the chat, name in case of JoinChatByName operation
//
// Typology :
// Typology of the operation
//
// Data :
// bytes that can be deserialized into a Chat or NodeInfo according to operation typology
// *------*-------------------*--------------*----------*---------*------*
// | Slot | lenTargetedChat   | TargetedChat | Typology | lenData | Data |
// *------*-------*------------*--------------*----------*---------*----*
//							 lenTargetedChat				  	  lenData
//	 1 byte	 	1 byte	  	 	bytes		  1 byte	 1 byte	   bytes

func (op *Operation) ToBytes() []byte {
	var bytes []byte
	bytes = append(bytes, op.Slot)
	bytes = append(bytes, uint8(len(op.TargetedChat)))
	bytes = append(bytes, []byte(op.TargetedChat)...)
	bytes = append(bytes, uint8(op.Typology))

	var dataBytes []byte
	if op.Data != nil {
		dataBytes = op.Data.ToBytes()
	}

	bytes = append(bytes, uint8(len(dataBytes)))
	bytes = append(bytes, dataBytes...)
	bytes = append(bytes, reader.Separator...)

	return bytes
}

func DecodeOperation(bytes []byte) (*Operation, error) {
	slot := bytes[0]
	offset, targetedChat := getField(1, bytes)
	typology := OperationType(bytes[offset])
	_, dataBytes := getField(offset+1, bytes)

	op := &Operation{
		Slot:         slot,
		Typology:     typology,
		TargetedChat: string(targetedChat),
		Data:         nil,
	}

	// decode data into concrete type when needed
	switch typology {
	case AddNode, SaveNode, RemoveChat, JoinChatByName:
		var result NodeInfos
		err := decodeData(dataBytes, &result)
		if err != nil {
			return nil, err
		}

		op.Data = &result

	case AddMessage:
		var result Message
		err := decodeData(dataBytes, &result)
		if err != nil {
			return nil, err
		}

		op.Data = &result

	case AddChat:
		var result Chat
		err := decodeData(dataBytes, &result)
		if err != nil {
			return nil, err
		}

		op.Data = &result
	}

	return op, nil
}

func decodeData(bytes []byte, result any) error {
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	return nil
}

func getField(offset int, source []byte) (int, []byte) {
	if len(source) <= offset-1 {
		return 0, []byte{}
	} else {
		lenField := int(source[offset])
		return offset + lenField + 1, source[offset+1 : offset+lenField+1]
	}
}

func GetOperationName(typology OperationType) string {
	return operationNames[typology]
}

func (o *Operation) Copy() *Operation {
	newOp := &Operation{}
	newOp.Slot = o.Slot
	newOp.Typology = o.Typology
	newOp.TargetedChat = o.TargetedChat
	newOp.Data = o.Data
	return newOp
}
