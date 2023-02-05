package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
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
		SetSlot(slot int)
		GetName() string
		ToBytes() ([]byte, error)
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

func (i *infos) SetSlot(slot int) {
	i.slot = slot
}

func (i *infos) ToBytes() ([]byte, error) {
	bytesMessage, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	return bytesMessage, nil
}

func DecodeInfos(bytes []byte) (Infos, error) {
	var i infos
	err := json.Unmarshal(bytes, &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}
