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
		myInfos     *crdt.NodeInfos
		currentChat *crdt.Chat
		storage     *storage.Storage
	}
)

func newOrchestrator(myInfos *crdt.NodeInfos) *orchestrator {
	currentChat := crdt.NewChat(myInfos.Name)
	storage := storage.NewStorage()
	storage.SaveChat(currentChat)

	return &orchestrator{
		myInfos:     myInfos,
		currentChat: currentChat,
		storage:     storage,
	}
}

func (o *orchestrator) getPropagationOperations(op *crdt.Operation, chat *crdt.Chat) <-chan *crdt.Operation {
	var syncOps = make(chan *crdt.Operation, 1)

	go func(syncOps chan *crdt.Operation) {
		defer close(syncOps)

		switch op.Typology {
		case crdt.JoinChatByName:
			slot := op.Slot
			createChatOperation := crdt.NewOperation(crdt.CreateChat, chat.Name, nil)
			createChatOperation.Slot = slot
			syncOps <- createChatOperation

			addNodeOperation := crdt.NewOperation(crdt.AddNode, chat.Id, o.myInfos)
			createChatOperation.Slot = slot
			syncOps <- addNodeOperation

			// propagates new node to other chats
			slots := chat.GetSlots(o.myInfos.Id)
			for _, s := range slots {
				addNodeOperation.Slot = s
				syncOps <- addNodeOperation
			}

		case crdt.AddMessage:
			slots := chat.GetSlots(o.myInfos.Id)
			for _, s := range slots {
				op.Slot = s
				syncOps <- op
			}
		}

	}(syncOps)

	return syncOps
}

func (o *orchestrator) getChatFromStorage(op crdt.Operation) (*crdt.Chat, error) {
	var (
		c   *crdt.Chat
		err error
	)

	if op.Typology == crdt.Quit{
		return nil, nil
	}

	c, err = o.storage.GetChat(op.TargetedChat, op.Typology==crdt.JoinChatByName)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// HandleChats maintains chat infos consistency by executing operation and building operations to send to other nodes if needed
func (o *orchestrator) handleChats(wg *sync.WaitGroup, incomingCommands chan parsestdin.Command, toExecute chan *crdt.Operation, toSend chan<- *crdt.Operation, shutdown <-chan struct{}) {
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

				newChat.SaveNode(o.myInfos)
				o.storage.SaveChat(newChat)

			case crdt.AddMessage:

				/* Add the messageBytes to discussion & sync with other nodes */
				toExecute <- crdt.NewOperation(crdt.AddMessage,
												o.currentChat.Id,
												crdt.NewMessage(o.myInfos.Name, args[parsestdin.MessageArg]))
			case crdt.ListChatsCommand:
				o.storage.DisplayChats()

			case crdt.LeaveChat:
				toExecute <- crdt.NewOperation(crdt.LeaveChat, o.currentChat.Id, o.myInfos)

			case crdt.Quit:
				toExecute <- crdt.NewOperation(crdt.Quit, "", o.myInfos)
			}

		case op := <-toExecute:
			c, err := o.getChatFromStorage(*op)
			if err != nil {
				log.Println("[ERROR]", err)
				continue
			}

			// execute op
			switch op.Typology {
			case crdt.CreateChat:

			case crdt.JoinChatByName:
				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				newNodeInfos.Slot = op.Slot
				c.SaveNode(newNodeInfos)
				o.storage.SaveChat(c)

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}

			case crdt.AddNode:
				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				newNodeInfos.Slot = op.Slot
				c.SaveNode(newNodeInfos)
				o.storage.SaveChat(c)

			case crdt.AddMessage:
				newMessage, ok := op.Data.(*crdt.Message)
				if !ok {
					log.Println("[ERROR] can't parse op data to Message")
					break
				}

				c.SaveMessage(newMessage)
				o.storage.SaveChat(c)

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}
			}
		}
	}
}
