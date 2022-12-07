package data

import (
	"net"
	"sync"

	"github.com/pkg/errors"
)

const notFound = "chat not found"

type (
	Chat struct {
		Index int
		Data  data
		Next  *Chat
	}
	data struct {
		Conn net.Conn
		Wg   *sync.WaitGroup
	}
)

func NewChat(wg *sync.WaitGroup, conn net.Conn) Chat {
	return Chat{
		Index: 0,
		Data: data{
			Conn: conn,
			Wg:   wg,
		},
		Next: nil,
	}
}

func AddChat(first *Chat, wg *sync.WaitGroup, conn net.Conn) *Chat {
	var p, tmp Chat

	if first == nil {
		tmp = NewChat(wg, conn)
		return &tmp
	}

	p = *first
	for p.Next != nil {
		p = *p.Next
	}
	p.Index = tmp.Index + 1
	p.Data = data{
		Wg:   wg,
		Conn: conn,
	}
	p.Next = &tmp

	return &p
}

func RemoveChat(first *Chat, index int) *Chat {
	var previous, tmp Chat

	tmp = *first
	for tmp.Index != index || tmp.Next != nil {
		previous = tmp
		tmp = *tmp.Next
	}

	if tmp.Index == index {
		previous.Next = tmp.Next
	}

	return first
}

func GetChat(first Chat, index int) (Chat, error) {
	var tmp Chat

	for tmp.Index != index || tmp.Next != nil {
		tmp = *tmp.Next
	}

	if tmp.Index == index {
		return tmp, nil
	}

	return Chat{}, errors.New(notFound)
}
