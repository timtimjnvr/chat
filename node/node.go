package node

import (
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
)

type (
	Node struct {
		Infos    Infos
		Business Business
		Next     *Node
	}

	Infos struct {
		Id      uuid.UUID `json:"id"`
		Port    int       `json:"port"`
		Address string    `json:"address"`
	}

	Business struct {
		Conn     net.Conn
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func NewNode(conn net.Conn) *Node {
	id := uuid.New()
	return &Node{
		Business: Business{
			Conn:     conn,
			Wg:       &sync.WaitGroup{},
			Shutdown: make(chan struct{}, 0),
		},
		Infos: Infos{
			Id: id,
		},
		Next: nil,
	}
}

func NewNodeInfos(id uuid.UUID, addr string, pt int) Infos {
	return Infos{
		Id:      id,
		Port:    pt,
		Address: addr,
	}
}

func (c *Node) SetConn(conn net.Conn) {
	c.Business.Conn = conn
}

func (c *Node) Stop() {
	close(c.Business.Shutdown)
	c.Business.Wg.Wait()
}

func (c *Node) display(position int) {
	log.Printf("%d: %s <-> %s\n", position, c.Business.Conn.LocalAddr(), c.Business.Conn.RemoteAddr())
}
