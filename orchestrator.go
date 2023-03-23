package main

import (
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"github/timtimjnvr/chat/storage"
	"log"
	"sync"
)

type (
	orchestrator struct {
		myInfos     crdt.Infos
		currentChat crdt.Chat
		storage     *storage.Storage
	}
)

func newOrchestrator(myInfos crdt.Infos) *orchestrator {
	currentChat := crdt.NewChat(myInfos.GetName())
	storage := storage.NewStorage()
	storage.SaveChat(currentChat)

	return &orchestrator{
		myInfos:     myInfos,
		currentChat: currentChat,
		storage:     storage,
	}
}

func (o *orchestrator) getPropagationOperations(op crdt.Operation, chat crdt.Chat) <-chan crdt.Operation {
	var (
		syncOps = make(chan crdt.Operation, 1)
		err     error
	)

	go func(syncOps chan crdt.Operation) {
		defer close(syncOps)

		switch op.GetOperationType() {
		case crdt.JoinChatByName:
			var chatInfos []byte
			chatInfos, err = chat.ToBytes()
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

			slot := op.GetSlot()
			createChatOperation := crdt.NewOperation(crdt.CreateChat, chat.GetId(), chatInfos)
			createChatOperation.SetSlot(slot)
			syncOps <- createChatOperation

			addNodeOperation := crdt.NewOperation(crdt.AddNode, chat.GetId(), myInfosByte)
			createChatOperation.SetSlot(slot)
			syncOps <- addNodeOperation

			// propagates new node to other chats
			slots := chat.GetSlots()
			for _, s := range slots {
				addNodeOperation.SetSlot(s)
				syncOps <- addNodeOperation
			}

		case crdt.AddMessage:
			slots := chat.GetSlots()
			for _, s := range slots {
				op.SetSlot(s)
				syncOps <- op
			}
		}

	}(syncOps)

	return syncOps
}

func (o *orchestrator) getChatFromStorage(op crdt.Operation) (crdt.Chat, error) {
	var (
		c   crdt.Chat
		err error
	)

	// get targeted chat
	switch op.GetOperationType() {
	// by name
	case crdt.JoinChatByName:
		c, err = o.storage.GetChat(op.GetTargetedChat(), true)
		if err != nil {
			return nil, err
		}

		return c, nil

	// by id
	default:
		c, err = o.storage.GetChat(op.GetTargetedChat(), false)
		if err != nil {
			return nil, err
		}

		return c, nil

	// no targeted chat needed
	case crdt.Quit:
		return nil, nil
	}
}

// HandleChats maintains chat infos consistency by executing operation and building operations to send to other nodes if needed
func (o *orchestrator) handleChats(wg *sync.WaitGroup, incomingCommands chan parsestdin.Command, toExecute chan crdt.Operation, toSend chan<- crdt.Operation, shutdown <-chan struct{}) {
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

				newChat.AddNode(o.myInfos)
				o.storage.SaveChat(newChat)

			case crdt.AddMessage:
				content := args[parsestdin.MessageArg]

				/* Add the messageBytes to discussion & sync with other nodes */
				var messageBytes []byte
				messageBytes = crdt.NewMessage(o.myInfos.GetName(), content).ToBytes()
				toExecute <- crdt.NewOperation(crdt.AddMessage, o.currentChat.GetId(), messageBytes)
			case crdt.ListChatsCommand:
				o.storage.DisplayChats()

			case crdt.LeaveChat:
				toExecute <- crdt.NewOperation(crdt.LeaveChat, o.currentChat.GetId(), o.myInfos.ToBytes())

			case crdt.Quit:
				toExecute <- crdt.NewOperation(crdt.Quit, "", o.myInfos.ToBytes())
			}

		case op := <-toExecute:
			c, err := o.getChatFromStorage(op)
			if err != nil {
				log.Println("[ERROR]", err)
				continue
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

				newNodeInfos.SetSlot(op.GetSlot())
				c.AddNode(newNodeInfos)
				o.storage.SaveChat(c)

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}

			case crdt.AddNode:
				var newNodeInfos crdt.Infos
				newNodeInfos, err = crdt.DecodeInfos(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
				}

				newNodeInfos.SetSlot(op.GetSlot())
				c.AddNode(newNodeInfos)
				o.storage.SaveChat(c)

			case crdt.AddMessage:
				var newMessage crdt.Message
				newMessage, err = crdt.DecodeMessage(op.GetOperationData())
				if err != nil {
					log.Println("[ERROR]", err)
					break
				}

				c.AddMessage(newMessage)
				o.storage.SaveChat(c)

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}
			}
		}
	}
}
