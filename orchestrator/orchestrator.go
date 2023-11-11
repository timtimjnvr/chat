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
		storage      *storage.Storage
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
	var (
		s     = storage.NewStorage()
		id, _ = s.AddNewChat(myInfos.Name)
		o     = &Orchestrator{
			RWMutex: &sync.RWMutex{},
			myInfos: myInfos,
			storage: s,
		}
	)
	_ = s.AddNodeToChat(myInfos, id)
	o.updateCurrentChat(id)

	return o
}

// HandleChats maintains chat infos consistency by executing and propagating operations received
// from stdin or TCP connections through the channel toExecute
func (o *Orchestrator) HandleChats(wg *sync.WaitGroup, toExecute chan *crdt.Operation, toSend chan<- *crdt.Operation, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case <-shutdown:
			return

		case op := <-toExecute:
			switch op.Typology {
			case crdt.JoinChatByName:
				chatID, err := o.storage.GetChatID(op.TargetedChat)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}
				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				newNodeSlot := op.Slot

				// create chat
				createChatOperation := crdt.NewOperation(crdt.AddChat, op.TargetedChat, &crdt.Chat{Id: chatID, Name: op.TargetedChat})
				createChatOperation.Slot = newNodeSlot
				toSend <- createChatOperation

				// add me
				addMeOperation := crdt.NewOperation(crdt.SaveNode, chatID.String(), o.myInfos)
				addMeOperation.Slot = newNodeSlot
				toSend <- addMeOperation

				// add other nodes
				slots, _ := o.storage.GetSlots(chatID)
				for _, s := range slots {
					nodeInfo, err := o.storage.GetNodeBySlot(s)
					if err != nil {
						fmt.Printf(logErrFrmt, err)
					}

					addNodeOperation := crdt.NewOperation(crdt.AddNode, chatID.String(), nodeInfo)
					addNodeOperation.Slot = newNodeSlot
					toSend <- addNodeOperation
				}

				// add new node
				newNodeInfos.Slot = op.Slot
				err = o.storage.AddNodeToChat(newNodeInfos, chatID)

				fmt.Printf("%s joined c\n", newNodeInfos.Name)
				fmt.Printf("connection established with %s\n", newNodeInfos.Name)
			case crdt.CreateChat:
				id, err := o.storage.AddNewChat(op.TargetedChat)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				// don't care about error since we just added the given c
				_ = o.storage.AddNodeToChat(o.myInfos, id)
				continue
			case crdt.AddChat:
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

				fmt.Printf("you joined a new c : %s\n", newChatInfos.Name)
				continue
			case crdt.ListChats:
				o.storage.DisplayChats()
			case crdt.ListUsers:
				err := o.storage.DisplayChatUsers(o.currenChatID)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
				}
			case crdt.AddNode, crdt.SaveNode:
				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				newNodeInfos, ok := op.Data.(*crdt.NodeInfos)
				if !ok {
					log.Println("[ERROR] can't parse op data to NodeInfos")
					continue
				}

				newNodeInfos.Slot = op.Slot
				err = o.storage.AddNodeToChat(newNodeInfos, chatID)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				log.Println(fmt.Sprintf("connection established with %s", newNodeInfos.Name))
			case crdt.AddMessage:
				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				newMessage, ok := op.Data.(*crdt.Message)
				if !ok {
					log.Println("[ERROR] can't parse op data to Message")
					break
				}

				err = o.storage.AddMessageToChat(newMessage, chatID)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				fmt.Printf("%s (%s): %s", newMessage.Sender, newMessage.Date, newMessage.Content)

				slots, err := o.storage.GetSlots(chatID)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				for _, s := range slots {
					op.Slot = s
					toSend <- op
				}
			case crdt.RemoveNode:
				o.storage.RemoveNodeSlotFromStorage(op.Slot)
			case crdt.LeaveChat:
				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logErrFrmt, err)
					continue
				}

				// Only one chat in storage
				if o.storage.GetNumberOfChats() <= 1 {
					fmt.Printf("[ERROR] You can't leave the current c\n")
					continue
				}

				chatNodeSlots, err := o.storage.GetSlots(chatID)
				if err != nil {
					fmt.Printf(logFrmt, err)
				}

				toDelete := make(map[uint8]bool)
				for _, slot := range chatNodeSlots {
					leaveOperation := crdt.NewOperation(crdt.RemoveNode, chatID.String(), nil)
					leaveOperation.Slot = slot
					toSend <- leaveOperation
				}

				for _, slot := range chatNodeSlots {
					toDelete[slot] = true
				}

				// Verify that slots are not used by any other chats
				for s, _ := range toDelete {
					if o.storage.IsSlotUsedByOtherChats(s, chatID) {
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

				// Removing c from storage and getting new current
				chatName, _ := o.storage.GetChatName(chatID)
				fmt.Printf("Leaving c %s\n", chatName)
				o.storage.RemoveChat(chatID)

				// Getting new current chat
				newID, _ := o.storage.GetNewCurrentChatID()
				o.updateCurrentChat(newID)
				newCurrentName, _ := o.storage.GetChatName(newID)
				fmt.Printf("Switched to c %s\n", newCurrentName)
			case crdt.Quit:
				process, err := os.FindProcess(os.Getpid())
				if err != nil {
					log.Fatal(err)
				}

				// signal main to stop
				err = process.Signal(os.Interrupt)
				if err != nil {
					log.Fatal(err)
				}
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
					id, err := o.storage.GetChatID(chatName)
					if err != nil {
						fmt.Printf(logErrFrmt, err)
					}

					o.updateCurrentChat(id)

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
