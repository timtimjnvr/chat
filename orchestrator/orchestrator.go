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

	"github.com/google/uuid"
)

type (
	Orchestrator struct {
		*sync.RWMutex
		debugMode    bool
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
	MaxMessagesStdin       = 100
	logErrFormat           = "[ERROR] %s\n"
	logFormat              = "[INFO] %s\n"
	logOpperationErrFormat = "[ERROR] [%s] %s\n"
	typeCommand            = "type a Command :"
)

func NewOrchestrator(storage *storage.Storage, myInfos *crdt.NodeInfos) *Orchestrator {
	var (
		s     = storage
		id, _ = s.AddNewChat(myInfos.Name)
		o     = &Orchestrator{
			RWMutex: &sync.RWMutex{},
			myInfos: myInfos,
			storage: s,
		}
	)

	o.updateCurrentChat(id)

	return o
}

// HandleChats maintains chat infos consistency by executing and propagating operations received
// from stdin or TCP connections through the channel toExecute
func (o *Orchestrator) HandleChats(wg *sync.WaitGroup, toExecute chan *crdt.Operation, toSend chan<- *crdt.Operation) {
	defer func() {
		close(toSend)
		wg.Done()
	}()

	for {
		select {
		case op, more := <-toExecute:
			if !more {

				return
			}

			switch op.Typology {
			case crdt.JoinChatByName:
				chatID, err := o.storage.GetChatID(op.TargetedChat)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
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
						fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
						continue
					}

					addNodeOperation := crdt.NewOperation(crdt.AddNode, chatID.String(), nodeInfo)
					addNodeOperation.Slot = newNodeSlot
					toSend <- addNodeOperation
				}

				// add new node
				newNodeInfos.Slot = op.Slot
				err = o.storage.AddNodeToChat(newNodeInfos, chatID)

				fmt.Printf(logFormat, fmt.Sprintf("%s joined chat", newNodeInfos.Name))

			case crdt.CreateChat:
				id, err := o.storage.AddNewChat(op.TargetedChat)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				// don't care about error since we just added the given chat
				_ = o.storage.AddNodeToChat(o.myInfos, id)

			case crdt.AddChat:
				newChatInfos, ok := op.Data.(*crdt.Chat)
				if !ok {
					fmt.Println("[ERROR] can't parse op data to Chat")
					continue
				}

				err := o.storage.AddChat(newChatInfos)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				o.updateCurrentChat(newChatInfos.Id)
				fmt.Printf(logFormat, fmt.Sprintf("you joined a new chat : %s", newChatInfos.Name))

			case crdt.AddNode, crdt.SaveNode:
				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
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
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}
				// in case of node we just added we need to ask the remote node to save us
				if op.Typology == crdt.AddNode {
					addMe := crdt.NewOperation(crdt.SaveNode, chatID.String(), o.myInfos)
					addMe.Slot = op.Slot
					toSend <- addMe
				}

			case crdt.AddMessage:
				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				newMessage, ok := op.Data.(*crdt.Message)
				if !ok {
					log.Println("[ERROR] can't parse op data to Message")
					break
				}

				err = o.storage.AddMessageToChat(newMessage, chatID)
				if err != nil {
					continue
				}

				// No error so we effectively got a new message
				fmt.Printf("%s (%s): %s", newMessage.Sender, newMessage.Date, newMessage.Content)

				slots, err := o.storage.GetSlots(chatID)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}
				for _, s := range slots {
					// Send message to all slots except the sender
					if op.Slot != s {
						copied := op.Copy()
						copied.Slot = s
						toSend <- copied
					}
				}

			case crdt.RemoveNode:
				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				err = o.storage.RemoveNodeFromChat(op.Slot, chatID)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
				}

			case crdt.KillNode:
				o.storage.RemoveNodeSlotFromStorage(op.Slot)

			case crdt.RemoveChat:
				// Only one chat in storage
				if o.storage.GetNumberOfChats() <= 1 {
					fmt.Printf("[ERROR] You can't leave the current c\n")
					continue
				}

				chatID, err := uuid.Parse(op.TargetedChat)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				chatNodeSlots, err := o.storage.GetSlots(chatID)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				// Killing needed connections and removing node from chat
				for _, s := range chatNodeSlots {
					if o.storage.IsSlotUsedByOtherChats(s, chatID) {
						leaveOperation := crdt.NewOperation(crdt.RemoveNode, chatID.String(), nil)
						leaveOperation.Slot = s
						toSend <- leaveOperation
					} else {
						removeNode := crdt.NewOperation(crdt.KillNode, "", nil)
						removeNode.Slot = s
						toSend <- removeNode
					}
				}

				chatName, err := o.storage.GetChatName(chatID)
				if err != nil {
					fmt.Printf(logOpperationErrFormat, crdt.GetOperationName(op.Typology), err)
					continue
				}

				//Removing chat from storage
				o.storage.RemoveChat(chatID)
				fmt.Printf(logFormat, fmt.Sprintf("Leaving %s", chatName))

				// Getting new current chat
				newID, _ := o.storage.GetNewCurrentChatID()
				o.updateCurrentChat(newID)
				newCurrentName, _ := o.storage.GetChatName(newID)
				fmt.Printf("Switched to chat %s\n", newCurrentName)

			case crdt.Quit:
				// Node handler need to close all TCP connections (node slot 0)
				toSend <- crdt.NewOperation(crdt.KillNode, "", nil)
				return
			}
		}
	}
}

func (o *Orchestrator) HandleStdin(osStdin *os.File, toExecute chan *crdt.Operation, outgoingConnectionRequests chan<- conn.ConnectionRequest, shutdown chan struct{}, sigC chan os.Signal) {
	var (
		wgReadStdin = sync.WaitGroup{}
		stdinChann  = make(chan []byte, MaxMessagesStdin)
		stopReading = make(chan struct{}, 0)
	)

	defer func() {

		close(stopReading)
		wgReadStdin.Wait()
	}()

	go reader.Read(osStdin, stdinChann, reader.Separator, stopReading)

	for {
		fmt.Printf(logFormat, typeCommand)

		select {
		case <-sigC:
			quit(toExecute, shutdown)
			return

		case line := <-stdinChann:
			cmd, err := parsestdin.NewCommand(string(line))
			if err != nil {
				fmt.Printf(logErrFormat, err)
				continue
			}

			args := cmd.GetArgs()
			switch cmd.GetTypology() {
			case crdt.JoinChatByName:
				if args[parsestdin.PortArg] == o.myInfos.Port && sameAddress(o.myInfos.Address, args[parsestdin.AddrArg]) {
					fmt.Printf(logErrFormat, "You are trying to connect to yourself")
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
						fmt.Printf(logErrFormat, err)
					}

					o.updateCurrentChat(id)

					fmt.Printf(logFormat, fmt.Sprintf("Switched to chat %s", chatName))

				case crdt.AddMessage:
					/* Add the messageBytes to discussion & sync with other nodes */
					toExecute <- crdt.NewOperation(crdt.AddMessage,
						o.currenChatID.String(),
						crdt.NewMessage(o.myInfos.Name, args[parsestdin.MessageArg]))

				case crdt.ListChats:
					o.storage.DisplayChats()

				case crdt.ListUsers:
					o.storage.DisplayNodes()

				case crdt.ListChatUsers:
					err = o.storage.DisplayChatUsers(o.currenChatID)
					if err != nil {
						fmt.Printf(logErrFormat, err)
					}

				case crdt.RemoveChat:
					toExecute <- crdt.NewOperation(crdt.RemoveChat, o.currenChatID.String(), o.myInfos)

				case crdt.Quit:
					quit(toExecute, shutdown)
					return
				}
			}
		}
	}
}

func quit(toExecute chan *crdt.Operation, shutdown chan struct{}) {
	toExecute <- crdt.NewOperation(crdt.Quit, "", nil)
	close(shutdown)
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
