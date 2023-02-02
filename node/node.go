package node

import (
	"github.com/google/uuid"
	"net"
	"sync"
)

type (
	Node struct {
		Infos    Infos
		Business Business
	}

	Infos struct {
		Id      uuid.UUID `json:"id"`
		Port    string    `json:"port"`
		Address string    `json:"address"`
		Name    string    `json: "name"`
	}

	Business struct {
		Conn             net.Conn
		MessagesReceived chan string
		MessagesToSend   chan string
		Wg               *sync.WaitGroup
		Shutdown         chan struct{}
	}
)

const maxSimultaneousMessages = 1000

func NewNode(conn net.Conn) *Node {
	return &Node{
		Business: Business{
			Conn:             conn,
			MessagesReceived: make(chan string, maxSimultaneousMessages),
			MessagesToSend:   make(chan string, maxSimultaneousMessages),
			Wg:               &sync.WaitGroup{},
			Shutdown:         make(chan struct{}, 0),
		},
	}
}

func (n *Node) SetNodeInfos(id uuid.UUID, addr, port string) {
	n.Infos = Infos{
		Id:      id,
		Port:    port,
		Address: addr,
	}
}

func NewNodeInfos(addr string, port, name string) Infos {
	id, _ := uuid.NewUUID()
	return Infos{
		Id:      id,
		Port:    port,
		Address: addr,
		Name:    name,
	}
}

func (n *Node) SetConn(conn net.Conn) {
	n.Business.Conn = conn
}

func (n *Node) Stop() {
	close(n.Business.Shutdown)
	n.Business.Wg.Wait()
}

func (i *Infos) ToRunes() []rune {
	var (
		idBytes   = []rune(i.Id.String())
		portBytes = []rune(i.Port)
		addrBytes = []rune(i.Address)
		nameBytes = []rune(i.Name)
		bytes     []rune
	)

	addBytes := func(destination []rune, source ...rune) []rune {
		lenBytes := []rune{int32(len(source))}
		destination = append(destination, lenBytes...)
		destination = append(destination, source...)

		return destination
	}

	bytes = addBytes(bytes, idBytes...)
	bytes = addBytes(bytes, portBytes...)
	bytes = addBytes(bytes, addrBytes...)
	bytes = addBytes(bytes, nameBytes...)

	return bytes
}
