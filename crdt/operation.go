package crdt

import (
	"encoding/json"
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
	AddNode
	AddMessage
	LeaveChat
	ListUsers
	ListChatsCommand
	Quit
)

func NewOperation(typology OperationType, targetedChat string, data Data) *Operation {
	return &Operation{
		Slot:         0,
		Typology:     typology,
		TargetedChat: targetedChat,
		Data:         data,
	}
}

// *------*-------------------*--------------*----------*---------*------*
// | Slot | lenTargetedChat   | TargetedChat | Typology | lenData | Data |
// *------*-------------------*--------------*----------*---------*------*
//							 lenTargetedChat				  	  lenData
//	 1 byte	 	1 byte	  	 	bytes		  1 byte	 1 byte	   bytes

func (op *Operation) ToBytes() []byte {
	var bytes []byte
	bytes = append(bytes, op.Slot)
	bytes = append(bytes, uint8(len(op.TargetedChat)))
	bytes = append(bytes, []byte(op.TargetedChat)...)
	bytes = append(bytes, uint8(op.Typology))
	var dataBytes = []byte{}
	if  op.Data != nil {
		dataBytes = op.Data.ToBytes()

	}

	bytes = append(bytes, uint8(len(dataBytes)))
	bytes = append(bytes, dataBytes...)

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

	// decode data into concrete type
	switch typology {
	case AddNode, LeaveChat, JoinChatByName:
		var result *NodeInfos
		err := DecodeData(dataBytes, result)
		if err != nil {
			return nil, err
		}

		op.Data = result

	case AddMessage:
		var result *Message
		err := DecodeData(dataBytes, result)
		if err != nil {
			return nil, err
		}

		op.Data = result
	default:
	}

	return op, nil
}

func DecodeData(bytes []byte, result any) error {
	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	return nil
}

func getField(offset int, source []byte) (int, []byte) {
	lenField := int(source[offset])
	return offset + lenField + 1, source[offset+1 : offset+lenField+1]
}
