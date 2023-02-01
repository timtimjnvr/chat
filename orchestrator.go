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

func orchestrate(wg *sync.WaitGroup, myInfos node.Infos, stdin, fromConnections <-chan []byte, newConnections chan net.Conn, chats linked.List, nodes linked.List, shutdown <-chan struct{}) {
	defer func() {
		wg.Done()
	}()

	var (
		currentChat     = crdt.NewChat(myInfos.Name)
		connectionsDone = make(chan uuid.UUID, maxSimultaneousConnections)
	)

	for {
		select {
		case <-shutdown:
			return

		case newConn := <-newConnections:
			/* Saves the new connection into newNode List */
			newNode := node.NewNode(newConn)
			newNode.Business.Wg.Add(1)
			go conn.HandleConnection(newNode, connectionsDone, newNode.Business.Shutdown)

		/* Node done */
		case id := <-connectionsDone:
			nodes.Delete(id)

		/* Command input */
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
					addr = args[parsestdin.AddrArg]
					// chatRoom = args[parsestdin.ChatRoomArg]
					pt int
				)

				pt, err = strconv.Atoi(args[parsestdin.PortArg])
				if err != nil {
					log.Println(err)
				}

				if addr == localhost || addr == localhostDecimalPointed {
					addr = ""
				}

				/* Open connection */
				var newConn net.Conn
				newConn, err = conn.OpenConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				newConnections <- newConn

				/* Sends a sync operation to enter the chat room with all his infos */
				syncMessage := crdt.NewOperation(crdt.JoinChat, myInfos.ToRunes())
				bytesSyncMessage := []byte(string(syncMessage.ToRunes()))
				conn.Send(newConn, bytesSyncMessage)

			case parsestdin.MsgCommandType:
				content := args[parsestdin.MessageArg]
				if currentChat == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				/* Add the message to discussion */
				message := crdt.NewMessage(myInfos.Name, content)
				currentChat.AddMessage(message)

				syncMessage := crdt.NewOperation(crdt.AddMessage, message.ToRunes())
				bytesSyncMessage := []byte(string(syncMessage.ToRunes()))
				nodesInfos := currentChat.GetNodesInfos()

				for _, i := range nodesInfos {
					var n interface{}
					n, err = nodes.GetById(i.Id)
					if err != nil {
						continue
					}

					node := n.(node.Node)

					conn.Send(node.Business.Conn, bytesSyncMessage)
				}

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
			// decode & execute operation
		}

	}
}
