package crdt

import (
	"chat/linked"
	"encoding/json"
	"github.com/google/uuid"
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

// HandleChats maintains chat infos consistency by parsing the different inputs (stdi & fromConnections), it propagates sync operations to node handler nodes if needed
func HandleChats(wg *sync.WaitGroup, myInfos Infos, toExecute <-chan []byte, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	var (
		chats = linked.NewList()
	)

	for {
		select {
		case <-shutdown:
			return

		case operationBytes := <-toExecute:
			slot := int(operationBytes[0])
			op, err := DecodeOperation(operationBytes[1:])

			var c Chat

			switch op.typology {

			case JoinChatByName:
				var (
					chatName      = op.targetedChat
					numberOfChats = chats.Len()
				)

				for index := 0; index < numberOfChats; index++ {
					var chatValue interface{}
					chatValue, _ = chats.GetByIndex(index)
					c = chatValue.(Chat)

					if c.GetName() == chatName {
						break
					}
				}

				if err != nil || c == nil {
					log.Println("[ERROR] ", err)
					continue
				}

			default:
				var id uuid.UUID
				id, err = uuid.Parse(op.GetTargetedChat())
				if err != nil {
					log.Println("[ERROR]", err)
					continue
				}

				var chatValue interface{}
				chatValue, err = chats.GetById(id)
				if err != nil {
					log.Println("[ERROR]", err)
					continue
				}

				c = chatValue.(Chat)
			}

			switch op.typology {
			case JoinChatByName:
				newNodeInfos, err := DecodeInfos(op.GetOperationData())
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
			}
		}
	}

}