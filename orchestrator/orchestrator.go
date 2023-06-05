package orchestrator

import (
	"fmt"
	"github/timtimjnvr/chat/conn"
	"github/timtimjnvr/chat/crdt"
	"github/timtimjnvr/chat/parsestdin"
	"github/timtimjnvr/chat/reader"
	"github/timtimjnvr/chat/storage"
	"log"
	"os"
	"sync"
)

type (
	Orchestrator struct {
		*sync.RWMutex
		myInfos     *crdt.NodeInfos
		currentChat *crdt.Chat
		storage     *storage.Storage
	}
)

func (o *Orchestrator) getCurrentChat() *crdt.Chat {
	o.RLock()
	defer o.RUnlock()

	return o.currentChat
}

func (o *Orchestrator) updateCurrentChat(c *crdt.Chat) {
	o.Lock()
	defer o.Unlock()
	var valueChat = *c
	o.currentChat = &valueChat
}

const (
	MaxMessagesStdin = 100

	logFrmt     = "[INFO] %s\n"
	typeCommand = "type a Command :"
)

func NewOrchestrator(myInfos *crdt.NodeInfos) *Orchestrator {
	currentChat := crdt.NewChat(myInfos.Name)
	storage := storage.NewStorage()

	o := &Orchestrator{
		RWMutex: &sync.RWMutex{},
		myInfos: myInfos,
		storage: storage,
	}

	o.updateCurrentChat(currentChat)
	storage.SaveChat(currentChat)

	return o
}

// HandleChats maintains chat infos consistency by executing operation and building operations to send to other nodes if needed
func (o *Orchestrator) HandleChats(wg *sync.WaitGroup, toExecute chan *crdt.Operation, toSend chan<- *crdt.Operation, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return

		case op := <-toExecute:
			// execute op
			if op.Typology == crdt.CreateChat {
				newChat := crdt.NewChat(op.TargetedChat)
				newChat.SaveNode(o.myInfos)
				o.storage.SaveChat(newChat)
				o.updateCurrentChat(newChat)
				continue
			}

			if op.Typology == crdt.AddChat {
				newChatInfos, ok := op.Data.(*crdt.Chat)
				if !ok {
					log.Println("[ERROR] can't parse op data to Chat")
					continue
				}

				newChat := crdt.NewChat(newChatInfos.Name)
				newChat.Id = newChatInfos.Id
				o.storage.SaveChat(newChat)
				o.updateCurrentChat(newChat)

				log.Println(fmt.Sprintf("you joined a new chat : %s", newChatInfos.Name))
				continue
			}

			if op.Typology == crdt.ListChats {
				o.storage.DisplayChats()
				continue
			}

			if op.Typology == crdt.ListUsers {
				o.currentChat.DisplayUsers()
				continue
			}

			if op.Typology == crdt.Quit {
				var (
					index = 0
					c     *crdt.Chat
					err   error
				)

				for err != nil {
					c, err = o.storage.GetChatByIndex(index)
					if err != nil {
						continue
					}

					c.RemoveNodeBySlot(op.Slot)
				}

				continue
			}

			// for other operation we need to get a chat from storage
			c, err := o.getChatFromStorage(*op)
			if err != nil {
				log.Println("[ERROR]", err)
				continue
			}

			switch op.Typology {
			case crdt.JoinChatByName:
				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				newNodeInfos.Slot = op.Slot
				c.SaveNode(newNodeInfos)
				o.storage.SaveChat(c)
				o.updateCurrentChat(c)

				fmt.Printf("%s joined chat", newNodeInfos.Name)

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

				o.updateCurrentChat(c)

				log.Println(fmt.Sprintf("connection established with %s", newNodeInfos.Name))

			case crdt.AddMessage:
				newMessage, ok := op.Data.(*crdt.Message)
				if !ok {
					log.Println("[ERROR] can't parse op data to Message")
					break
				}

				if c.ContainsMessage(newMessage) {
					continue
				}

				c.SaveMessage(newMessage)
				o.storage.SaveChat(c)
				o.updateCurrentChat(c)

				fmt.Printf("%s (%s): %s", newMessage.Sender, newMessage.Date, newMessage.Content)

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}
			}
		}
	}
}

func (o *Orchestrator) HandleStdin(wg *sync.WaitGroup, toExecute chan *crdt.Operation, outgoingConnectionRequests chan<- conn.ConnectionRequest, shutdown chan struct{}) {
	var wgReadStdin = sync.WaitGroup{}

	defer func() {
		wgReadStdin.Wait()
		wg.Done()
	}()

	wgReadStdin.Add(1)
	var stdin = make(chan []byte, MaxMessagesStdin)
	go reader.Read(&wgReadStdin, os.Stdin, stdin, reader.Separator, shutdown)

	for {
		fmt.Printf(logFrmt, typeCommand)

		select {
		case <-shutdown:
			return

		case line := <-stdin:
			cmd, err := parsestdin.NewCommand(string(line))
			if err != nil {
				log.Println("[ERROR] ", err)
				continue
			}

			args := cmd.GetArgs()
			switch cmd.GetTypology() {
			case crdt.JoinChatByName:
				outgoingConnectionRequests <- conn.NewConnectionRequest(args[parsestdin.PortArg], args[parsestdin.AddrArg], args[parsestdin.ChatRoomArg])

			default:
				switch cmd.GetTypology() {
				case crdt.CreateChat:
					var chatName = args[parsestdin.ChatRoomArg]
					toExecute <- crdt.NewOperation(crdt.CreateChat,
						chatName, nil)

				case crdt.AddMessage:
					/* Add the messageBytes to discussion & sync with other nodes */
					toExecute <- crdt.NewOperation(crdt.AddMessage,
						o.getCurrentChat().Id,
						crdt.NewMessage(o.myInfos.Name, args[parsestdin.MessageArg]))

				case crdt.ListChats:
					toExecute <- crdt.NewOperation(crdt.ListChats, "", nil)

				case crdt.ListUsers:
					toExecute <- crdt.NewOperation(crdt.ListUsers, "", nil)

				case crdt.LeaveChat:
					toExecute <- crdt.NewOperation(crdt.LeaveChat, o.getCurrentChat().Id, o.myInfos)

				case crdt.Quit:
					toExecute <- crdt.NewOperation(crdt.Quit, "", o.myInfos)
				}
			}
		}
	}
}

func (o *Orchestrator) getPropagationOperations(op *crdt.Operation, chat *crdt.Chat) <-chan *crdt.Operation {
	var syncOps = make(chan *crdt.Operation)

	go func(syncOps chan *crdt.Operation) {
		defer close(syncOps)

		switch op.Typology {
		case crdt.JoinChatByName:
			slot := op.Slot
			createChatOperation := crdt.NewOperation(crdt.AddChat, chat.Name, chat)
			createChatOperation.Slot = slot
			syncOps <- createChatOperation

			addNodeOperation := crdt.NewOperation(crdt.AddNode, chat.Id, o.myInfos)
			syncOps <- addNodeOperation

			// propagates new node to other chats
			slots := chat.GetSlots(o.myInfos.Id)
			for _, s := range slots {
				addNodeOperation.Slot = s
				syncOps <- addNodeOperation
			}

			// sending chat messages
			addMessageOperations := chat.GetMessageOperationsForPropagation()

			for _, s := range slots {
				for _, addMessageOperation := range addMessageOperations {
					addMessageOperation.Slot = s
					syncOps <- addMessageOperation
				}
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

func (o *Orchestrator) getChatFromStorage(op crdt.Operation) (*crdt.Chat, error) {
	var (
		c   *crdt.Chat
		err error
	)

	if op.Typology == crdt.Quit {
		return nil, nil
	}

	c, err = o.storage.GetChat(op.TargetedChat, op.Typology == crdt.JoinChatByName)
	if err != nil {
		return nil, err
	}

	return c, nil
}
