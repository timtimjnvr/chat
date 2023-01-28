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

func orchestrate(wg *sync.WaitGroup, initialChat *string, stdin, fromConnections <-chan []byte, newNodes chan *node.Node, chats linked.List, nodes  linked.List, shutdown <-chan struct{} ){
	defer func() {
		wg.Done()
	}()

	var (
		currentChat   = crdt.NewChat(*initialChat)
		connectionsDone = make(chan uuid.UUID, maxSimultaneousConnections)
	)

	for {
		select {
		case <-shutdown:
			return

		/* Save the connection and handle connection*/
		case newNode := <-newNodes:
			nodes.Add(newNode)
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
					addr     = args[parsestdin.AddrArg]
					chatRoom = args[parsestdin.ChatRoomArg]
					newConn  net.Conn
					pt       int
				)

				pt, err = strconv.Atoi(args[parsestdin.PortArg])
				if err != nil {
					log.Println(err)
				}

				if addr == localhost || addr == localhostDecimalPointed {
					addr = ""
				}

				/* Open connection */
				newConn, err = conn.OpenConnection(transportProtocol, addr, pt)
				if err != nil {
					log.Println("[ERROR] ", err)
					continue
				}

				/* Saves the new connection into newNode List */
				newNode := node.NewNode(newConn)
				newNode.Business.Wg.Add(1)
				go conn.HandleConnection(newNode, connectionsDone, newNode.Business.Shutdown)

				/* Add the newNode to chatRoom */
				if chatRoom == *initialChat {
					currentChat.AddNode(newNode)
				}

				/* sync -> add the new newNode to other nodes */

			case parsestdin.MsgCommandType:
				content := args[parsestdin.MessageArg]
				if currentChat == nil {
					log.Println(noDiscussionSelected)
					continue
				}

				/* Add the message to discussion */
				message := crdt.NewMessage(*initialChat, content)
				currentChat.AddMessage(message)

				syncMessage := crdt.NewOperation(crdt.AddMessage, message.ToRunes())
				bytesSyncMessage := []byte(string(syncMessage.ToRunes()))
				currentChat.Send(bytesSyncMessage)

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
