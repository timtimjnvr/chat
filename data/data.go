package data

import (
	"log"
	"net"
	"sync"
)

const notFound = "chat not found"

type (
	ChatList struct {
		head   *Chat
		length int
	}

	Chat struct {
		Infos ChatInfos
		Next  *Chat
	}

	ChatInfos struct {
		Conn     net.Conn
		Wg       *sync.WaitGroup
		Shutdown chan struct{}
	}
)

func (l *ChatList) isEmpty() bool {
	return l.length == 0
}

func NewChatList() (l *ChatList) {
	return &ChatList{
		head:   nil,
		length: 0,
	}
}

func NewChat(conn net.Conn) *Chat {
	return &Chat{
		Infos: ChatInfos{
			Conn:     conn,
			Wg:       &sync.WaitGroup{},
			Shutdown: make(chan struct{}, 0),
		},
		Next: nil,
	}
}

func (l *ChatList) AddChat(c *Chat) {
	if l.isEmpty() {
		l.head = c
		l.length += 1
		return
	}

	ptr := l.head
	for ptr.Next != nil {
		ptr = ptr.Next
	}

	ptr.Next = c
}

func (l *ChatList) GetChat(position int) *Chat {
	if l.isEmpty() || l.length-1 < position {
		return nil
	}
	var chat = l.head

	for position != 0 {
		position -= 1
		chat = chat.Next
	}

	return chat
}
func (c *Chat) display(position int) {
	log.Printf("%d: %s <-> %s\n", position, c.Infos.Conn.LocalAddr(), c.Infos.Conn.RemoteAddr())
}

func (l *ChatList) Display() {
	var (
		chat     = l.head
		position int
	)

	for chat.Next != nil {
		chat.display(position)
		chat = chat.Next
		position++
	}

	// last of the list
	chat.display(position)
}

/*func (l *ChatList) RemoveChat(index int) {
	var previous, tmp ChatList

	tmp = *l
	for tmp.Index != index || tmp.Next != nil {
		tmp = *tmp.Next
	}

	if tmp.Index == index {
		previous.Next = tmp.Next
	}
}*/
