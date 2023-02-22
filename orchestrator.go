package main

import (
	"github.com/google/uuid"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/linked"
	"github/timtimjnvr/chat/parsestdin"
	"log"
	"sync"
)

type (
	orchestrator struct {
		myInfos     crdt.Infos
		currentChat crdt.Chat
		chats       linked.List
	}
)

func NewOrchestrator(myInfos crdt.Infos) *orchestrator {
	return &orchestrator{
		myInfos:     myInfos,
		currentChat: crdt.NewChat(myInfos.GetName()),
		chats:       linked.NewList(),
	}
}

func (o *orchestrator) getChat(identifier string, byName bool) (crdt.Chat, error) {
	var (
		numberOfChats = o.chats.Len()
		c             crdt.Chat
		err           error
	)

	if byName {
		for index := 0; index < numberOfChats; index++ {
			var chatValue interface{}
			chatValue, _ = o.chats.GetByIndex(index)
			c = chatValue.(*crdt.ConcreteChat)

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

	return chatValue.(*crdt.ConcreteChat), nil
}

func (o *orchestrator) addNewChat(newChat crdt.Chat) {
	newChat.AddNode(o.myInfos)
	o.chats.Add(newChat)
}

// HandleChats maintains chat infos consistency by executing operation and building operations to send to other nodes if needed
func (o *orchestrator) HandleChats(wg *sync.WaitGroup, incomingCommands chan parsestdin.Command, toSend chan<- crdt.Operation, toExecute chan crdt.Operation, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return

		case cmd := <-incomingCommands:
			args := cmd.GetArgs()

			switch cmd.GetTypology() {
			case crdt.CreateChat:
				var (
					chatName = args[parsestdin.ChatRoomArg]
					newChat  = crdt.NewChat(chatName)
				)

				o.addNewChat(newChat)

			case crdt.AddMessage:
				content := args[parsestdin.MessageArg]

				/* Add the messageBytes to discussion & sync with other nodes */
				var messageBytes []byte
				messageBytes = crdt.NewMessage(o.myInfos.GetName(), content).ToBytes()
				toExecute <- crdt.NewOperation(crdt.AddMessage, o.currentChat.GetId(), messageBytes)

			case crdt.LeaveChat:
				toExecute <- crdt.NewOperation(crdt.LeaveChat, o.currentChat.GetId(), o.myInfos.ToBytes())

			case crdt.Quit:
				toExecute <- crdt.NewOperation(crdt.Quit, "", o.myInfos.ToBytes())
			}

		case op := <-toExecute:
			var (
				slot = op.GetSlot()
				c    crdt.Chat
				err  error
			)

			// get targeted chat
			switch op.GetOperationType() {
			// by name
			case crdt.JoinChatByName:
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
			case crdt.Quit:
			}

			// execute op
			switch op.GetOperationType() {
			case crdt.JoinChatByName:
				var newNodeInfos crdt.Infos
				newNodeInfos, err = crdt.DecodeInfos(op.GetOperationData())
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

				createChatOperation := crdt.NewOperation(crdt.CreateChat, c.GetId(), chatInfos)
				createChatOperation.SetSlot(slot)
				toSend <- createChatOperation

				addNodeOperation := crdt.NewOperation(crdt.AddNode, c.GetId(), myInfosByte)

				// propagates new node to other chats
				slots := c.GetSlots()
				for _, s := range slots {
					addNodeOperation.SetSlot(s)
					toSend <- addNodeOperation
				}

			case crdt.AddNode:
				var newNodeInfos crdt.Infos
				newNodeInfos, err = crdt.DecodeInfos(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
				}

				newNodeInfos.SetSlot(slot)
				c.AddNode(newNodeInfos)

			case crdt.AddMessage:
				var newMessage crdt.Message
				newMessage, err = crdt.DecodeMessage(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
					break
				}
				c.AddMessage(newMessage)

				slots := c.GetSlots()
				for _, s := range slots {
					op.SetSlot(s)
					toSend <- op
				}
			}
		}
	}
}
