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
		mode             int
		myInfos          crdt.Infos
		currentChat      crdt.Chat
		chats            linked.List
		toExecute        chan crdt.Operation
		toSend           chan<- crdt.Operation
		incomingCommands <-chan parsestdin.Command
	}
)

func newOrchestrator(myInfos crdt.Infos, incomingCommands <-chan parsestdin.Command, toExecute chan crdt.Operation, toSend chan<- crdt.Operation) *orchestrator {
	return &orchestrator{
		myInfos:          myInfos,
		currentChat:      crdt.NewChat(myInfos.GetName()),
		chats:            linked.NewList(),
		toExecute:        toExecute,
		toSend:           toSend,
		incomingCommands: incomingCommands,
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

func (o *orchestrator) getOperationFromCommand(cmd parsestdin.Command) <-chan crdt.Operation {
	var (
		op = make (chan crdt.Operation, 1)
		args = cmd.GetArgs()
	)

	go func(op chan crdt.Operation) {
		defer close(op)

		switch cmd.GetTypology() {
		case crdt.CreateChat:
			var (
				chatName = args[parsestdin.ChatRoomArg]
				newChat  = crdt.NewChat(chatName)
			)

			newChat.AddNode(o.myInfos)
			o.chats.Add(newChat)

		case crdt.AddMessage:
			content := args[parsestdin.MessageArg]

			/* Add the messageBytes to discussion & sync with other nodes */
			var messageBytes []byte
			messageBytes = crdt.NewMessage(o.myInfos.GetName(), content).ToBytes()
			op <-crdt.NewOperation(crdt.AddMessage, o.currentChat.GetId(), messageBytes)

		case crdt.LeaveChat:
			op <- crdt.NewOperation(crdt.LeaveChat, o.currentChat.GetId(), o.myInfos.ToBytes())

		case crdt.Quit:
			op <- crdt.NewOperation(crdt.Quit, "", o.myInfos.ToBytes())
		}

	}(op)

	return op
}

// HandleChats maintains chat infos consistency by executing operation and building operations to send to other nodes if needed
func (o *orchestrator) handleChats(wg *sync.WaitGroup, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return

		case cmd := <-o.incomingCommands:
			op, ok := <- o.getOperationFromCommand(cmd)
			if ok {
				o.toExecute <-op
			}

		case op := <-o.toExecute:
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
				o.toSend <- createChatOperation

				addNodeOperation := crdt.NewOperation(crdt.AddNode, c.GetId(), myInfosByte)

				// propagates new node to other chats
				slots := c.GetSlots()
				for _, s := range slots {
					addNodeOperation.SetSlot(s)
					o.toSend <- addNodeOperation
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
					o.toSend <- op
				}
			}
		}
	}
}
