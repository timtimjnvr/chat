package crdt

import "encoding/json"

type (
	infos struct {
		slot    int
		Port    string `json:"port"`
		Address string `json:"address"`
		name    string `json: "name"`
	}

	Infos interface {
		GetName() string
		ToBytes() ([]byte, error)
	}
)

func NewNodeInfos(addr string, port, name string) Infos {
	return &infos{
		slot:    -1,
		Port:    port,
		Address: addr,
		name:    name,
	}
}

func (i infos) GetName() string {
	return i.name
}

func (i infos) ToBytes() ([]byte, error) {
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

	return i, nil
}
