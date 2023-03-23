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
				chatInfos, _ := newChat.ToBytes()
				toExecute <- crdt.NewOperation(crdt.CreateChat, args[parsestdin.ChatRoomArg], chatInfos)

			case crdt.AddMessage:
				content := args[parsestdin.MessageArg]

				var messageBytes []byte
				messageBytes = crdt.NewMessage(o.myInfos.GetName(), content).ToBytes()
				toExecute <- crdt.NewOperation(crdt.AddMessage, o.currentChat.GetId(), messageBytes)

			// will only need to be displayed
			case crdt.ListUsers:
				o.currentChat.DisplayUser()

			case crdt.ListChatsCommand:
				o.storage.DisplayChats()
			}

		case op := <-toExecute:

			var (
				c   crdt.Chat
				err error
			)

			switch op.GetOperationType() {
			// execute
			case crdt.CreateChat:
				c = crdt.NewChat(op.GetTargetedChat())
				o.storage.SaveChat(c)

			// get chat from storage
			default:
				c, err = o.getChatFromStorage(op)
				if err != nil {
					log.Println("[ERROR]", err)
					continue
				}
			}

			// execute operation
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
