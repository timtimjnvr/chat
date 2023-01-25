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

func (l *list) AddNode(node *Node) {
	if l.isEmpty() {
		l.head = node
		l.length += 1
		return
	}

	ptr := l.head
	for ptr.Next != nil {
		ptr = ptr.Next
	}

	l.length += 1
	ptr.Next = node
}

func (l *list) GetNode(id uuid.UUID) *Node {
	if l.isEmpty() {
		return nil
	}

	var (
		node  = l.head
		index int
	)

	for node.Next != nil && node.Infos.Id != id {
		index += 1
		node = node.Next
	}

	if node.Infos.Id == id {
		return node
	}

	return nil
}

func (l *list) Display() {
	var (
		node     = l.head
		position int
	)

	for node.Next != nil {
		node.display(position)
		node = node.Next
		position++
	}

	// last of the list
	node.display(position)
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
	var node = l.head

	for node.Next != nil {
		close(node.Business.Shutdown)
	}

	for node.Next != nil {
		node.Business.Wg.Wait()
		l.RemoveNode(node.Infos.Id)
	}
}

func (l *list) isEmpty() bool {
	return l.length == 0
}
