package main

import (
	"chat/conn"
	"chat/crdt"
	"chat/linked"
	"chat/node"
	"chat/parsestdin"
	"github.com/google/uuid"
	"log"
	"net"
	"strconv"
	"sync"
)

// orchestrate maintains node list & chat infos consistency by parsing the different inputs (stdin, fromConnections, newConnections), it propagates sync operations to other nodes if needed
func orchestrate(wg *sync.WaitGroup, myInfos node.Infos, chats linked.List, nodes linked.List, stdin, fromConnections <-chan []byte, newConnections chan net.Conn, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	var (
		chatNameId      = make(map[string]uuid.UUID)
		currentChat     = crdt.NewChat(myInfos.Name)
		connectionsDone = make(chan uuid.UUID, maxSimultaneousConnections)
	)

	// used by users join chat by name
	chatNameId[currentChat.GetName()] = currentChat.GetId()

	for {
		select {
		case <-shutdown:
			return

		case newConn := <-newConnections:
			newNode := node.NewNode(newConn)
			newNode.Business.Wg.Add(1)
			go conn.HandleConnection(newNode, connectionsDone, newNode.Business.Shutdown)

		case id := <-connectionsDone:
			nodes.Delete(id)

			// TODO update all chats where the node was inside

		case line := <-stdin:
			cmd, err := parsestdin.NewCommand(string(line))
			if err != nil {
				log.Println("[ERROR] ", err)
			}

			args := cmd.GetCommandArgs()

			switch typology := cmd.GetCommandType(); typology {
			case parsestdin.CreateChatCommandType:
				var (
					chatName = args[parsestdin.ChatRoomArg]
					newChat  = crdt.NewChat(chatName)
				)

				chats.Add(newChat)

			case parsestdin.ConnectCommandType:
				var (
					addr     = args[parsestdin.AddrArg]
					chatRoom = args[parsestdin.ChatRoomArg]
					pt       int
				)

				pt, err = strconv.Atoi(args[parsestdin.PortArg])
				if err != nil {
					log.Println(err)
				}

				/* Open connection */
				var newConn net.Conn
				newConn, err = conn.OpenConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				newConnections <- newConn

				/* Sends a sync operation to the connected node to enter the chat room with all his infos */
				syncMessage := crdt.NewOperation(crdt.JoinChat, chatNameId[chatRoom], myInfos.ToRunes())
				bytesSyncMessage := []byte(string(syncMessage.ToRunes()))
				conn.Send(newConn, bytesSyncMessage)

			case parsestdin.MsgCommandType:
				content := args[parsestdin.MessageArg]
				if currentChat == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				/* Add the message to discussion & sync with other nodes */
				message := crdt.NewMessage(myInfos.Name, content)
				currentChat.AddMessage(message)
				syncMessage := crdt.NewOperation(crdt.AddMessage, currentChat.GetId(), message.ToRunes())
				bytesSyncMessage := []byte(string(syncMessage.ToRunes()))

				var nodeIds = make([]uuid.UUID, 0, chats.Len())
				for _, nodesInfos := range currentChat.GetNodesInfos() {
					if nodesInfos.Id == myInfos.Id {
						continue
					}

					nodeIds = append(nodeIds, nodesInfos.Id)
				}

				// TODO send to nodes

				/* for _, i := range nodesInfos {
					var n interface{}
					n, err = nodes.GetById(i.Id)
					if err != nil {
						continue
					}

					node := n.(node.Node)

					conn.Send(node.Business.Conn, bytesSyncMessage)
				}
				*/

			case parsestdin.CloseCommandType:
				/* TODO
				leave discussion and gracefully shutdown connection with all nodes
				*/

			case parsestdin.QuitCommandType:
				/* TODO
				leave all discussions and gracefully shutdown connection with all nodes
				*/
			}

		case <-fromConnections:
			/*operation := crdt.DecodeOperation(operationBytes)

			var targetedChat interface{}
			targetedChat, err := chats.GetById(operation.targetedChat)
			if err != nil {
				log.Println("[ERROR] no chat id in operation")
			}

			chat := targetedChat.(crdt.Chat)

			switch operation.GetOperationType() {
			case crdt.JoinChat:
				/*nodeInfos := DecodeNodeInfos(operation.GetOperationData())

				for _, i := range nodesInfos {
					var n interface{}
					n, err = nodes.GetById(i.Id)
					if err != nil {
						continue
					}

					node := n.(node.Node)

					conn.Send(node.Business.Conn, bytesSyncMessage)
				}

			case crdt.AddMessage:
			case crdt.RemoveMessage:
			case crdt.UpdateMessage:
			case crdt.LeaveChat:
			}
			*/
		}
	}

}
