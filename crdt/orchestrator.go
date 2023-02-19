package crdt

import (
	"github.com/google/uuid"
	"github/timtimjnvr/chat/linked"
	"log"
	"sync"
)

type (
	orchestrator struct {
		myInfos Infos
		chats   linked.List
	}
)

func NewOrchestrator(myInfos Infos) *orchestrator {
	return &orchestrator{
		myInfos: myInfos,
		chats:   linked.NewList(),
	}
}

func (o *orchestrator) getChat(identifier string, byName bool) (Chat, error) {
	var (
		numberOfChats = o.chats.Len()
		c             Chat
		err           error
	)

	if byName {
		for index := 0; index < numberOfChats; index++ {
			var chatValue interface{}
			chatValue, _ = o.chats.GetByIndex(index)
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
	chatValue, err = o.chats.GetById(id)

	if err != nil {
		return nil, linked.NotFound
	}

	return chatValue.(*chat), nil
}

func (o *orchestrator) addNewChat(name string) {
	o.chats.Add(NewChat(name))
}

// HandleChats maintains chat infos consistency by executing operation and building operations to send to other nodes if needed
func (o *orchestrator) HandleChats(wg *sync.WaitGroup, toSend chan<- Operation, toExecute <-chan Operation, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return

		case op := <-toExecute:
			var (
				slot = op.GetSlot()
				c    Chat
				err  error
			)

			// get targeted chat
			switch op.GetOperationType() {
			// by name
			case JoinChatByName:
				c, err = o.getChat(op.GetTargetedChat(), true)
				log.Println("[ERROR]", err)
				continue

			// by id
			default:
				c, err = o.getChat(op.GetTargetedChat(), false)
				if err != nil {
					log.Println("[ERROR]", err)
					continue
				}

			// no targeted chat needed
			case Quit:
			}

			// execute op
			switch op.GetOperationType() {
			case JoinChatByName:
				var newNodeInfos Infos
				newNodeInfos, err = DecodeInfos(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
					break
				}

				newNodeInfos.SetSlot(slot)
				c.AddNode(newNodeInfos)

				var chatInfos []byte
				chatInfos, err = c.ToBytes()
				if err != nil {
					log.Println("[ERROR]", err)
					break
				}

				var myInfosByte []byte
				myInfosByte = o.myInfos.ToBytes()
				if err != nil {
					log.Println("[ERROR]", err)
					break
				}

				createChatOperation := NewOperation(CreateChat, c.GetId(), chatInfos)
				createChatOperation.SetSlot(slot)
				toSend <- createChatOperation

				addNodeOperation := NewOperation(AddNode, c.GetId(), myInfosByte)

				// propagates new node to other chats
				slots := c.getSlots()
				for _, s := range slots {
					addNodeOperation.SetSlot(s)
					toSend <- addNodeOperation
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
				var newMessage Message
				newMessage, err = DecodeMessage(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
					break
				}
				c.AddMessage(newMessage)

				slots := c.getSlots()
				for _, s := range slots {
					op.SetSlot(s)
					toSend <- op
				}
			}
		}
	}
}
