package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
)

type (
	NodeInfos struct {
		Slot    uint8
		Id      uuid.UUID `json:"id"`
		Port    string    `json:"port"`
		Address string    `json:"address"`
		Name    string    `json:"name"`
	}
)

func NewNodeInfos(addr string, port, name string) *NodeInfos {
	id, _ := uuid.NewUUID()

	return &NodeInfos{
		Slot:    0,
		Id:      id,
		Port:    port,
		Address: addr,
		Name:    name,
	}
}

func (i *NodeInfos) GetID() uuid.UUID {
	return i.Id
}

func (i *NodeInfos) GetName() string {
	return i.Name
}

func (i *NodeInfos) ToBytes() []byte {
	bytesMessage, _ := json.Marshal(i)
	return bytesMessage
}
