package data

import (
	"github.com/google/uuid"
	"log"
	"net"
	"sync"
)

type (
	Chat struct {
		Id    uuid.UUID
		Infos ChatInfos
		Next  *Chat
	}

	ChatInfos struct {
		Conn     net.Conn
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func NewChat(conn net.Conn) *Chat {
	id, _ := uuid.NewUUID()
	return &Chat{
		Id: id,
		Infos: ChatInfos{
			Conn:     conn,
			Wg:       &sync.WaitGroup{},
			Shutdown: make(chan struct{}, 0),
		},
		Next: nil,
	}
}

func (c *Chat) Stop() {
	close(c.Infos.Shutdown)
	c.Infos.Wg.Wait()
}

func (c *Chat) display(position int) {
	log.Printf("%d: %s <-> %s\n", position, c.Infos.Conn.LocalAddr(), c.Infos.Conn.RemoteAddr())
}
