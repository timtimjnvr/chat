package node

import "github.com/google/uuid"

const notFound = "chat not found"

type (
	list struct {
		head   *Node
		length int
	}

	NodeList interface {
		AddNode(c *Node)
		RemoveNode(id uuid.UUID)
		GetNode(id uuid.UUID) (c *Node)
		Display()
		CloseAndWaitNode()
	}
)

func NewNodeList() (l NodeList) {
	return &list{
		head:   nil,
		length: 0,
	}
}

func (l *list) AddNode(c *Node) {
	if l.isEmpty() {
		l.head = c
		l.length += 1
		return
	}

	ptr := l.head
	for ptr.Next != nil {
		ptr = ptr.Next
	}

	l.length += 1
	ptr.Next = c
}

func (l *list) GetNode(id uuid.UUID) *Node {
	if l.isEmpty() {
		return nil
	}

	var (
		chat  = l.head
		index int
	)

	for chat.Next != nil && chat.Infos.Id != id {
		index += 1
		chat = chat.Next
	}

	if chat.Infos.Id == id {
		return chat
	}

	return nil
}

func (l *list) Display() {
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

func (l *list) RemoveNode(id uuid.UUID) {
	var previous, tmp *Node

	// remove first element
	if l.head.Infos.Id == id {
		l.head = l.head.Next
		return
	}

	// second or more
	previous = l.head
	tmp = l.head.Next
	for tmp.Next != nil && id != tmp.Infos.Id {
		previous = tmp
		tmp = tmp.Next
		tmp = previous.Next
	}

	if id != tmp.Infos.Id {
		previous.Next = tmp.Next
	}
}

func (l *list) CloseAndWaitNode() {
	var chat = l.head

	for chat.Next != nil {
		close(chat.Business.Shutdown)
	}

	for chat.Next != nil {
		chat.Business.Wg.Wait()
		l.RemoveNode(chat.Infos.Id)
	}
}

func (l *list) isEmpty() bool {
	return l.length == 0
}
