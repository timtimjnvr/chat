package node

import (
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
)

type (
	Node struct {
		Id    uuid.UUID
		Infos Infos
		Next  *Node
	}

	Infos struct {
		Conn     net.Conn
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func NewChat(conn net.Conn) *Node {
	id, _ := uuid.NewUUID()
	return &Node{
		Id: id,
		Infos: Infos{
			Conn:     conn,
			Wg:       &sync.WaitGroup{},
			Shutdown: make(chan struct{}, 0),
		},
		Next: nil,
	}
}

func (c *Node) Stop() {
	close(c.Infos.Shutdown)
	c.Infos.Wg.Wait()
}

func (c *Node) display(position int) {
	log.Printf("%d: %s <-> %s\n", position, c.Infos.Conn.LocalAddr(), c.Infos.Conn.RemoteAddr())
}
