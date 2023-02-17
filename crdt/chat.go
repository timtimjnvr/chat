package crdt

import (
	"encoding/json"
	"github.com/google/uuid"
	"github/timtimjnvr/chat/linked"
	"log"
	"sync"
)

type (
	chat struct {
		Id       string `json:"Id"`
		myNodeId uuid.UUID
		Name     string `json:"name"`
		nodes    []Infos
		messages []Message
	}

	Chat interface {
		GetId() string
		GetName() string
		GetNodesInfos() []Infos
		getSlots() []int
		AddNode(infos Infos)
		AddMessage(message Message)
		ToBytes() ([]byte, error)
	}
)

func NewChat(name string) Chat {
	id, _ := uuid.NewUUID()
	return &chat{
		Id:       id.String(),
		Name:     name,
		nodes:    []Infos{},
		messages: []Message{},
	}
}

func (c *chat) GetNodesInfos() []Infos {
	return c.nodes
}

func (c *chat) GetId() string {
	return c.Id
}

func (c *chat) GetName() string {
	return c.Name
}

func (c *chat) AddNode(i Infos) {
	c.nodes = append(c.nodes, i)
}

func (c *chat) AddMessage(message Message) {
	if !c.containsMessage(message) {
		// TODO : insert message in array by comparing dates
		c.messages = append(c.messages, message)
	}
}

func (c *chat) containsMessage(message Message) bool {
	for _, m := range c.messages {
		if m.GetId() == message.GetId() {
			return true
		}
	}
	return false
}

func (c *chat) ToBytes() ([]byte, error) {
	bytesChat, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return bytesChat, nil
}

func (c *chat) getSlots() []int {
	slots := make([]int, 0, len(c.nodes))
	for _, i := range c.nodes {
		if i.getId() == c.myNodeId {
			slots = append(slots, i.getSlot())
		}
	}

	return slots
}

// HandleChats maintains chat infos consistency by executing operation and propagating operations to other nodes if needed
func HandleChats(wg *sync.WaitGroup, myInfos Infos, toSend chan<- []byte, toExecute <-chan []byte, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	var (
		chats = linked.NewList()
	)

	chats.Add(NewChat(myInfos.GetName()))

	for {
		select {
		case <-shutdown:
			return

		case operationBytes := <-toExecute:
			var (
				slot = int(operationBytes[0])
				op   = DecodeOperation(operationBytes[1:])
				c    Chat
				err  error
			)

			// get targeted chat
			switch op.GetOperationType() {
			// by name
			case JoinChatByName:
				c, err = getChat(chats, op.GetTargetedChat(), true)

			// by id
			default:
				c, err = getChat(chats, op.GetTargetedChat(), false)

			// no targeted chat needed
			case Quit:
			}

			// execute operation
			switch op.GetOperationType() {
			case JoinChatByName:
				var newNodeInfos Infos
				newNodeInfos, err = DecodeInfos(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
				}

				newNodeInfos.SetSlot(slot)
				c.AddNode(newNodeInfos)

				var chatInfos []byte
				chatInfos, err = c.ToBytes()
				if err != nil {
					log.Println("[ERROR]", err)
				}

				var myInfosByte []byte
				myInfosByte = myInfos.ToBytes()
				if err != nil {
					log.Println("[ERROR]", err)
				}

				createChatOperation := NewOperation(CreateChat, c.GetId(), chatInfos).ToBytes()
				toSend <- AddSlot(slot, createChatOperation)

				addNodeOperation := NewOperation(AddNode, c.GetId(), myInfosByte).ToBytes()

				// propagates new node to other chats
				slots := c.getSlots()
				for _, s := range slots {
					toSend <- AddSlot(s, addNodeOperation)
				}

			case AddNode:
				var newNodeInfos Infos
				newNodeInfos, err = DecodeInfos(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
				}

				newNodeInfos.SetSlot(slot)
				c.AddNode(newNodeInfos)

			case AddMessage:
				newMessage, err := DecodeMessage(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
				}
				c.AddMessage(newMessage)

				slots := c.getSlots()
				for _, s := range slots {
					toSend <- AddSlot(s, operationBytes)
				}
			}
		}
	}
}

func getChat(chats linked.List, identifier string, byName bool) (Chat, error) {
	var (
		numberOfChats = chats.Len()
		c             Chat
		err           error
	)

	if byName {
		for index := 0; index < numberOfChats; index++ {
			var chatValue interface{}
			chatValue, _ = chats.GetByIndex(index)
			c = chatValue.(*chat)

			if c.GetName() == identifier {
				return c, nil
			}
		}

		if err != nil || c == nil {
			return nil, linked.NotFound
		}

	}
	// by uuid
	var id uuid.UUID
	id, err = uuid.Parse(identifier)
	if err != nil {
		return nil, linked.InvalidIdentifier
	}

	var chatValue interface{}
	chatValue, err = chats.GetById(id)

	if err != nil {
		return nil, linked.NotFound
	}

	return chatValue.(*chat), nil
}

func AddSlot(slot int, operation []byte) []byte {
	return append([]byte{uint8(slot)}, operation...)
}
