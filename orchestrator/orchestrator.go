package orchestrator

import (
	"fmt"
	"github.com/google/uuid"
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
		myInfos      *crdt.NodeInfos
		currenChatID uuid.UUID
		storage      storage.Storage
	}
)

func (o *Orchestrator) updateCurrentChat(currenChatID uuid.UUID) {
	o.Lock()
	defer o.Unlock()
	o.currenChatID = currenChatID
}

const (
	MaxMessagesStdin = 100
	logErrFrmt       = "[ERROR] %s\n"
	logFrmt          = "[INFO] %s\n"
	typeCommand      = "type a Command :"
)

func NewOrchestrator(myInfos *crdt.NodeInfos) *Orchestrator {
	s := storage.NewStorage()
	id, _ := s.AddNewChat(myInfos.Name)
	_ = s.AddNodeToChat(myInfos, id)

	o := &Orchestrator{
		RWMutex: &sync.RWMutex{},
		myInfos: myInfos,
		storage: s,
	}

	o.updateCurrentChat(id)
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
				id, err := o.storage.AddNewChat(op.TargetedChat)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				// don't care about error since we just added the given chat
				_ = o.storage.AddNodeToChat(o.myInfos, id)
				continue
			}

			if op.Typology == crdt.AddChat {
				newChatInfos, ok := op.Data.(*crdt.Chat)
				if !ok {
					fmt.Println("[ERROR] can't parse op data to Chat")
					continue
				}

				err := o.storage.AddChat(newChatInfos)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				id := newChatInfos.Id

				err = o.storage.AddNodeToChat(o.myInfos, id)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				o.updateCurrentChat(id)

				fmt.Printf("you joined a new chat : %s\n", newChatInfos.Name)
				continue
			}

			if op.Typology == crdt.ListChats {
				o.storage.DisplayChats()
				continue
			}

			if op.Typology == crdt.ListUsers {
				o.storage.DisplayChatUsers(o.currenChatID)
				continue
			}

			// there is no chat specified in operation in this case we need to remove node identified by slot from all chats
			if op.Typology == crdt.RemoveNode {
				o.storage.RemoveNodeSlotFromStorage(op.Slot)
				continue
			}

			// for other operation we need to get a chat from storage
			c, err := o.storage.GetChat(op.TargetedChat, op.Typology == crdt.JoinChatByName)
			if err != nil {
				fmt.Printf(logErrFrmt, err)
				continue
			}

			id := c.Id

			switch op.Typology {
			case crdt.JoinChatByName:
				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}

				newNodeInfos.Slot = op.Slot
				err = o.storage.AddNodeToChat(newNodeInfos, id)

				fmt.Printf("%s joined chat\n", newNodeInfos.Name)
				fmt.Printf("connection established with %s\n", newNodeInfos.Name)

			// connection just established
			case crdt.AddNode, crdt.SaveNode:
				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				newNodeInfos.Slot = op.Slot
				c.SaveNode(newNodeInfos)
				o.storage.AddNodeToChat(newNodeInfos, id)

				log.Println(fmt.Sprintf("connection established with %s", newNodeInfos.Name))

			case crdt.AddMessage:
				newMessage, ok := op.Data.(*crdt.Message)
				if !ok {
					log.Println("[ERROR] can't parse op data to Message")
					break
				}

				err = o.storage.AddMessageToChat(newMessage, id)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				fmt.Printf("%s (%s): %s", newMessage.Sender, newMessage.Date, newMessage.Content)

				for syncOp := range o.getPropagationOperations(op, c) {
					toSend <- syncOp
				}

			case crdt.LeaveChat:
				// Only one chat in storage
				if o.storage.GetNumberOfChats() <= 1 {
					fmt.Printf("[ERROR] You can't leave the current chat\n")
					continue
				}

				var (
					chatNodeSlots = c.GetSlots(o.myInfos.Id)
					toDelete      = make(map[uint8]bool)
				)

				for _, slot := range chatNodeSlots {
					leaveOperation := crdt.NewOperation(crdt.RemoveNode, op.TargetedChat, nil)
					leaveOperation.Slot = slot
					toSend <- leaveOperation
				}

				for _, slot := range chatNodeSlots {
					toDelete[slot] = true
				}

				// Verify that slots are not used by any other chats
				for s, _ := range toDelete {
					if o.storage.IsUsedByOtherChats(s, o.myInfos.Id, id) {
						toDelete[s] = false
					}
				}

				// Operation used by node handler to kill connections if it is not used anymore
				for s, killConnection := range toDelete {
					if killConnection {
						removeNode := crdt.NewOperation(crdt.KillNode, "", nil)
						removeNode.Slot = s
						toSend <- removeNode
					}
				}

				// Removing chat from storage and getting new current
				fmt.Printf("Leaving chat %s\n", c.Name)
				o.storage.RemoveChat(id)
				newID, _ := o.storage.GetNewCurrent()
				o.updateCurrentChat(newID)
				newCurrent, _ := o.storage.GetChat(newID.String(), false)
				fmt.Printf("Switched to chat %s\n", newCurrent.Name)

			case crdt.RemoveNode:
				o.storage.RemoveNodeSlotFromStorage(op.Slot)

			case crdt.Quit:
				process, err := os.FindProcess(os.Getpid())
				if err != nil {
					log.Fatal(err)
				}

				// signal main to stop
				process.Signal(os.Interrupt)
			}
		}
	}
}

func (o *Orchestrator) HandleStdin(wg *sync.WaitGroup, toExecute chan *crdt.Operation, outgoingConnectionRequests chan<- conn.ConnectionRequest, shutdown chan struct{}) {
	var (
		wgReadStdin = sync.WaitGroup{}
		stdin       = make(chan []byte, MaxMessagesStdin)
		stopReading = make(chan struct{}, 0)
	)

	defer func() {
		close(stopReading)
		wgReadStdin.Wait()
		wg.Done()
	}()

	isDone := make(chan struct{})
	go reader.Read(os.Stdin, stdin, reader.Separator, stopReading, isDone)

	for {
		fmt.Printf(logFrmt, typeCommand)

		select {
		case <-shutdown:
			return

		case line := <-stdin:
			cmd, err := parsestdin.NewCommand(string(line))
			if err != nil {
				fmt.Printf(logErrFrmt, err)
				continue
			}

			args := cmd.GetArgs()
			switch cmd.GetTypology() {
			case crdt.JoinChatByName:
				if args[parsestdin.PortArg] == o.myInfos.Port && sameAddress(o.myInfos.Address, args[parsestdin.AddrArg]) {
					fmt.Printf(logErrFrmt, "You are trying to connect to yourself")
					continue
				}

				outgoingConnectionRequests <- conn.NewConnectionRequest(args[parsestdin.PortArg], args[parsestdin.AddrArg], args[parsestdin.ChatRoomArg])

			default:
				switch cmd.GetTypology() {
				case crdt.CreateChat:
					toExecute <- crdt.NewOperation(crdt.CreateChat, args[parsestdin.ChatRoomArg], nil)

				case crdt.SwitchChat:
					chatName := args[parsestdin.ChatRoomArg]
					c, err := o.storage.GetChat(chatName, true)
					if err != nil {
						fmt.Printf(logErrFrmt, err)
					}

					o.updateCurrentChat(c.Id)

					fmt.Printf(logFrmt, fmt.Sprintf("Switched to chat %s", chatName))

				case crdt.AddMessage:
					/* Add the messageBytes to discussion & sync with other nodes */
					toExecute <- crdt.NewOperation(crdt.AddMessage,
						o.currenChatID.String(),
						crdt.NewMessage(o.myInfos.Name, args[parsestdin.MessageArg]))

				case crdt.ListChats:
					toExecute <- crdt.NewOperation(crdt.ListChats, "", nil)

				case crdt.ListUsers:
					toExecute <- crdt.NewOperation(crdt.ListUsers, "", nil)

				case crdt.LeaveChat:
					toExecute <- crdt.NewOperation(crdt.LeaveChat, o.currenChatID.String(), o.myInfos)

				case crdt.Quit:
					process, err := os.FindProcess(os.Getpid())
					if err != nil {
						log.Fatal(err)
					}

					// signal main to stop
					process.Signal(os.Interrupt)
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

			// add me
			addMeOperation := crdt.NewOperation(crdt.SaveNode, chat.Id.String(), o.myInfos)
			addMeOperation.Slot = slot
			syncOps <- addMeOperation

			// add other nodes
			slots := chat.GetSlots(o.myInfos.Id)
			for _, s := range slots {
				nodeInfo, err := chat.GetNodeBySlot(s)
				if err != nil {
					log.Println(err)
				}
				addNodeOperation := crdt.NewOperation(crdt.AddNode, chat.Id.String(), nodeInfo)
				addNodeOperation.Slot = slot
				syncOps <- addNodeOperation
			}

			// sending chat history
			addMessageOperations := chat.GetMessageOperationsForPropagation()
			for _, addMessageOperation := range addMessageOperations {
				addMessageOperation.Slot = slot
				syncOps <- addMessageOperation
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

func sameAddress(addr1, addr2 string) bool {
	if addr1 == addr2 {
		return true
	}

	isLocalhost := func(addr string) bool {
		return addr == "" || addr == "localhost" || addr == "127.0.0.1"
	}

	// TODO later : get address from domain and compare addresses

	return isLocalhost(addr1) && isLocalhost(addr2)
}
