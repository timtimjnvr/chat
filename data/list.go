package data

import "github.com/google/uuid"

const notFound = "chat not found"

type (
	ChatList struct {
		head   *Chat
		length int
	}
)

func NewChatList() (l *ChatList) {
	return &ChatList{
		head:   nil,
		length: 0,
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

	l.length += 1
	ptr.Next = c
}

func (l *ChatList) GetChat(position int) *Chat {
	if l.isEmpty() || l.length-1 < position {
		return nil
	}

	var (
		chat  = l.head
		index int
	)

	for index != position {
		index += 1
		chat = chat.Next
	}

	return chat
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

func (l *ChatList) RemoveChat(id uuid.UUID) {
	var previous, tmp *Chat

	// remove first element
	if l.head.Id == id {
		l.head = l.head.Next
		return
	}

	// second or more
	previous = l.head
	tmp = l.head.Next
	for tmp.Next != nil && id != tmp.Id {
		previous = tmp
		tmp = tmp.Next
		tmp = previous.Next
	}

	if id != tmp.Id {
		previous.Next = tmp.Next
	}
}

func (l *ChatList) CloseAndWaitChats() {
	var chat = l.head

	for chat.Next != nil {
		close(chat.Infos.Shutdown)
	}

	for chat.Next != nil {
		chat.Infos.Wg.Wait()
		l.RemoveChat(chat.Id)
	}
}

func (l *ChatList) isEmpty() bool {
	return l.length == 0
}
