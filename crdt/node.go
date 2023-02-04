package crdt

import "encoding/json"

type (
	infos struct {
		slot    int
		Port    string `json:"port"`
		Address string `json:"address"`
		Name    string `json:"name"`
	}

	Infos interface {
		SetSlot(slot int)
		GetName() string
		ToBytes() ([]byte, error)
	}
)

func NewNodeInfos(addr string, port, name string) Infos {
	return &infos{
		slot:    -1,
		Port:    port,
		Address: addr,
		Name:    name,
	}
}

func (i *infos) GetName() string {
	return i.Name
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
