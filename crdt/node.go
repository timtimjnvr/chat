package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
)

type (
	infos struct {
		slot    int
		Id      uuid.UUID `json:"Id"`
		Port    string    `json:"port"`
		Address string    `json:"address"`
		Name    string    `json:"name"`
	}

	Infos interface {
		getId() uuid.UUID
		getSlot() int
		GetAddr() string
		GetPort() string
		GetName() string
		SetSlot(slot int)
		ToBytes() []byte
	}
)

func NewNodeInfos(id uuid.UUID, addr string, port, name string) Infos {
	return &infos{
		slot:    -1,
		Id:      id,
		Port:    port,
		Address: addr,
		Name:    name,
	}
}

func (i *infos) getId() uuid.UUID {
	return i.Id
}

func (i *infos) GetName() string {
	return i.Name
}

func (i *infos) getSlot() int {
	return i.slot
}

func (i *infos) GetAddr() string {
	return i.Address
}

func (i *infos) GetPort() string {
	return i.Port
}

func (i *infos) SetSlot(slot int) {
	i.slot = slot
}

func (i *infos) ToBytes() []byte {
	bytesMessage, err := json.Marshal(i)
	if err != nil {
		log.Println("[ERROR] ", err)
		return nil
	}

	return bytesMessage
}

func DecodeInfos(bytes []byte) (Infos, error) {
	var i infos
	err := json.Unmarshal(bytes, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}
