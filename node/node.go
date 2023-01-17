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
		Port     int
		Addr     int
		Conn     net.Conn
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func NewChat(conn net.Conn) *Node {
	return &Node{
		Business: Business{
			Conn:     conn,
			Wg:       &sync.WaitGroup{},
			Shutdown: make(chan struct{}, 0),
		},
		Next: nil,
	}
}

func (c *Node) Stop() {
	close(c.Business.Shutdown)
	c.Business.Wg.Wait()
}

func (c *Node) display(position int) {
	log.Printf("%d: %s <-> %s\n", position, c.Business.Conn.LocalAddr(), c.Business.Conn.RemoteAddr())
}
